// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"errors"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/ibc"
	evmostypes "github.com/evmos/evmos/v20/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmante "github.com/evmos/evmos/v20/x/evm/ante"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/precompiles/callback"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
	callbacktypes "github.com/sagaxyz/saga-sdk/x/transferrouter/v10types"
)

// Execute executes a call to another contract through the Gateway precompile.
func (p Precompile) Execute(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) (retBz []byte, retErr error) {
	packetQueueItem, err := p.popNextPacket(ctx)
	if err != nil {
		return nil, err
	}

	packet := *packetQueueItem.Packet

	cachedCtx, writeFn := ctx.CacheContext()
	cachedCtx = evmante.BuildEvmExecutionCtx(cachedCtx)

	var packetData transfertypes.FungibleTokenPacketData

	// This defer is used so we can write the cached context events back to the main context, but also to clear the returned error,
	// so it can still remove the packet from the queue if the execution fails.
	defer func() {
		success := retErr == nil
		if retErr == nil {
			// Write cachedCtx events back to ctx only if the execution is successful
			writeFn()
		}

		// The tx must always succeed in order to be shown in the block explorer, so we never return an error here
		retErr = nil

		// Create the acknowledgement for the packet
		var ack channeltypes.Acknowledgement
		if success {
			ack = channeltypes.NewResultAcknowledgement([]byte{1})
		} else {
			ack = channeltypes.NewErrorAcknowledgement(errors.New("failed to execute call"))
		}

		err = p.transferKeeper.WriteAcknowledgementForPacket(ctx, packet, packetData, ack)
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("failed to write IBC acknowledgment", "error", err)
			retErr = err
			return
		}
		p.transferKeeper.Logger(ctx).Info("Successfully wrote IBC acknowledgment", "success", success)
	}()

	// parse the packet data, TODO: do not use this cdc
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &packetData); err != nil {
		p.transferKeeper.Logger(ctx).Error("Failed to unmarshal packet data", "error", err)
		return nil, err
	}

	// Check if the token pair exists and get the ERC20 contract address
	// for the native ERC20 or the precompile.
	// This call fails if the token does not exist or is not registered.
	coin := ibc.GetReceivedCoin(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), packetData.Denom, packetData.Amount)

	tokenPairID := p.transferKeeper.Erc20Keeper.GetTokenPairID(ctx, coin.Denom)
	tokenPair, found := p.transferKeeper.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		p.transferKeeper.Logger(ctx).Error("Token pair not found", "denom", packetData.Denom, "tokenPairID", tokenPairID)
		return nil, errorsmod.Wrapf(erc20types.ErrTokenPairNotFound, "token pair for denom %s not found", packetData.Denom)
	}

	var (
		resp *evmtypes.MsgEthereumTxResponse
		logs []*ethtypes.Log
	)

	// if the packet is a callback packet we process it as such, if not, we assume it's a normal erc20 transfer
	cbData, isCbPacket, err := callbacktypes.GetCallbackData(packetData, callbacktypes.V1, packet.GetDestPort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().Limit(), callbacktypes.DestinationCallbackKey)
	if isCbPacket {
		p.transferKeeper.Logger(ctx).Debug("Processing callback packet")
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("failed to get callback data", "error", err)
			retErr = err
			return
		}
		p.transferKeeper.Logger(ctx).Info("Successfully retrieved callback data", "callbackAddress", cbData.CallbackAddress, "senderAddress", packetData.Sender, "commitGasLimit", cbData.CommitGasLimit)

		resp, logs, err = p.executeDestinationCallback(ctx, cachedCtx, packet, packetData, cbData, tokenPair)
		retErr = err
	} else {
		resp, logs, err = p.executeERC20Transfer(ctx, cachedCtx, stateDB, packet, packetData, tokenPair)
		retErr = err
	}

	// Emit event for the packet, regardless of success or failure, as we want to show the result in the block explorer.
	// Note that we are doing it on the original context, we must not use the cached context here.
	if err := p.emitGatewayExecuteEvent(ctx, stateDB, p.Address(), packet.Sequence, retErr == nil, packetQueueItem.OriginalTxHash, isCbPacket, false, resp.Ret); err != nil {
		p.transferKeeper.Logger(ctx).Error("failed to emit gateway execute event", "error", err)
		return nil, err
	}

	if err != nil {
		p.transferKeeper.Logger(ctx).Error("failed to execute call", "error", err)
		return nil, err
	}

	for _, log := range logs {
		stateDB.AddLog(log)
	}

	return resp.Ret, nil
}

// popNextPacket gets the next packet from the queue and removes it
func (p Precompile) popNextPacket(ctx sdk.Context) (types.PacketQueueItem, error) {
	var packet types.PacketQueueItem
	logger := p.transferKeeper.Logger(ctx)

	if err := p.transferKeeper.PacketQueue.Walk(ctx, nil, func(key uint64, value types.PacketQueueItem) (bool, error) {
		logger.Debug("Processing packet from queue", "key", key, "value", value)
		packet = value
		return true, nil // stop after first
	}); err != nil {
		return types.PacketQueueItem{}, err
	}

	// remove the packet from the queue
	err := p.transferKeeper.PacketQueue.Remove(ctx, packet.Packet.Sequence)
	if err != nil {
		return types.PacketQueueItem{}, err
	}

	return packet, nil
}

func (p Precompile) executeERC20Transfer(ctx, cachedCtx sdk.Context, stateDB vm.StateDB, packet channeltypes.Packet, packetData transfertypes.FungibleTokenPacketData, tokenPair erc20types.TokenPair) (*evmtypes.MsgEthereumTxResponse, []*ethtypes.Log, error) {
	callData, err := CreateERC20TransferExecuteCallDataFromPacket(ctx, p.transferKeeper, packet, packetData)
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("Failed to create gateway execute call data", "error", err)
		return nil, nil, errorsmod.Wrapf(ErrEVMCallFailed, "failed to create gateway execute call data: %v", err)
	}

	// Execute the call logic here
	// This is where you would call your keeper methods to perform the actual execution
	fromAddress := common.BytesToAddress(p.Address().Bytes()) // the sender for normal ERC20 transfers is the gateway contract address
	target := tokenPair.GetERC20Contract()

	result, err := p.evmKeeper.CallEVMWithData(cachedCtx, fromAddress, &target, callData, true)
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("EVM message call failed", "error", err)
		return nil, nil, err
	}

	logs := evmtypes.LogsToEthereum(result.Logs)

	// consume gas in the original context
	ctx.GasMeter().ConsumeGas(result.GasUsed, "ERC20 transfer")
	if ctx.GasMeter().IsOutOfGas() {
		p.transferKeeper.Logger(ctx).Error("Out of gas after ERC20 transfer", "gasUsed", result.GasUsed)
		return nil, nil, errorsmod.Wrapf(errortypes.ErrOutOfGas, "out of gas")
	}

	return result, logs, nil
}

// executeDestinationCallback executes a callback packet, the cachedCtx must be a cached context and ctx must be the original context that we can use to consume gas
func (p Precompile) executeDestinationCallback(ctx, cachedCtx sdk.Context, packet channeltypes.Packet, packetData transfertypes.FungibleTokenPacketData, cbData callbacktypes.CallbackData, tokenPair erc20types.TokenPair) (*evmtypes.MsgEthereumTxResponse, []*ethtypes.Log, error) {
	target := common.HexToAddress(cbData.CallbackAddress)

	// Generate secure isolated address from sender, we know this address is initialized in the IBC OnRecvPacket
	p.transferKeeper.Logger(ctx).Debug("Generating isolated address", "senderAddress", packetData.Sender, "destChannel", packet.GetDestChannel())
	isolatedAddr := utils.GenerateIsolatedAddress(packet.GetDestChannel(), packetData.Sender)

	ctx = ctx.WithGasMeter(evmostypes.NewInfiniteGasMeterWithLimit(cbData.CommitGasLimit))

	amountInt, ok := math.NewIntFromString(packetData.Amount)
	if !ok {
		return nil, nil, errors.New("invalid amount")
	}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract

	// TODO: remaining gas not used until we update to Cosmos EVM
	remainingGas := math.NewIntFromUint64(cachedCtx.GasMeter().GasRemaining()).BigInt()

	// Call the EVM with the remaining gas as the maximum gas limit.
	// Up to now, the remaining gas is equal to the callback gas limit set by the user.
	// NOTE: use the cached ctx for the EVM calls.
	res, err := p.evmKeeper.CallEVM(cachedCtx, erc20.ABI, common.Address(isolatedAddr), tokenPair.GetERC20Contract(), true, "approve", target, amountInt.BigInt())
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("ERC20 approve call failed", "error", err)
		return nil, nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance: %w", err)
	}

	// only add logs if the call was successful
	logs := evmtypes.LogsToEthereum(res.Logs)

	// Consume the actual used gas on the original callback context.
	ctx.GasMeter().ConsumeGas(res.GasUsed, "callback allowance")
	remainingGas = remainingGas.Sub(remainingGas, math.NewIntFromUint64(res.GasUsed).BigInt())
	p.transferKeeper.Logger(ctx).Debug("Consumed gas for approve", "gasUsed", res.GasUsed, "remainingGas", remainingGas.String())
	if ctx.GasMeter().IsOutOfGas() || remainingGas.Cmp(big.NewInt(0)) < 0 {
		p.transferKeeper.Logger(ctx).Error("Out of gas after approve", "remainingGas", remainingGas.String())
		return nil, nil, errorsmod.Wrapf(errortypes.ErrOutOfGas, "out of gas")
	}

	var approveSuccess bool
	err = erc20.ABI.UnpackIntoInterface(&approveSuccess, "approve", res.Ret)
	if err != nil {
		return nil, nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to unpack approve return: %w", err)
	}

	if !approveSuccess {
		return nil, nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance")
	}
	// NOTE: use the cached ctx for the EVM calls.
	p.transferKeeper.Logger(ctx).Debug("Starting callback EVM call", "fromAddress", isolatedAddr.String(), "target", target.Hex(), "calldataLength", len(cbData.Calldata))
	res, err = p.evmKeeper.CallEVMWithData(cachedCtx, common.Address(isolatedAddr), &target, cbData.Calldata, true)
	if err != nil {
		return nil, nil, errorsmod.Wrapf(ErrEVMCallFailed, "EVM returned error: %w", err)
	}

	// only add logs if the call was successful
	logs = append(logs, evmtypes.LogsToEthereum(res.Logs)...)

	// Consume the actual gas used on the original callback context.
	ctx.GasMeter().ConsumeGas(res.GasUsed, "callback function")
	if ctx.GasMeter().IsOutOfGas() {
		return nil, nil, errorsmod.Wrapf(errortypes.ErrOutOfGas, "out of gas")
	}

	// Check that the sender no longer has tokens after the callback.
	// NOTE: contracts must implement an IERC20(token).transferFrom(msg.sender, address(this), amount)
	// for the total amount, or the callback will fail.
	// This check is here to prevent funds from getting stuck in the isolated address,
	// since they would become irretrievable.
	receiverTokenBalance := p.transferKeeper.Erc20Keeper.BalanceOf(cachedCtx, erc20.ABI, tokenPair.GetERC20Contract(), common.Address(isolatedAddr)) // here,
	// we can use the original ctx and skip manually adding the gas
	if receiverTokenBalance.Cmp(big.NewInt(0)) != 0 {
		p.transferKeeper.Logger(ctx).Error("Receiver still has tokens after callback", "balance", receiverTokenBalance.String())
		return nil, nil, errorsmod.Wrapf(erc20types.ErrEVMCall,
			"receiver has %d unrecoverable tokens after callback", receiverTokenBalance)
	}
	p.transferKeeper.Logger(ctx).Debug("Callback processing completed successfully")

	return res, logs, nil
}

// ExecuteSrcCallback executes a source callback packet, the process is similar to Execute
func (p Precompile) ExecuteSrcCallback(ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) (retBz []byte, retErr error) {

	packetQueueItem, err := p.popNextSrcCallback(ctx)
	if err != nil {
		return nil, err
	}

	// cache ctx
	cachedCtx, writeFn := ctx.CacheContext()
	cachedCtx = evmante.BuildEvmExecutionCtx(cachedCtx)

	// the from address is the IBC module address, this is only so the contracts can verify the caller
	acc, _ := p.transferKeeper.AccountKeeper.GetModuleAccountAndPermissions(ctx, "txrouter")

	cbData, err := getSourceCallbackData(ctx, packetQueueItem)
	if err != nil {
		return nil, err
	}

	target := common.HexToAddress(cbData.CallbackAddress)

	// Here we assemble the calldata depending on if it's an acknowledgement or a timeout
	var calldata []byte
	if packetQueueItem.IsTimeout {
		calldata, err = callback.ABI.Pack("onPacketTimeout", packetQueueItem.Packet.SourceChannel, packetQueueItem.Packet.SourcePort, packetQueueItem.Packet.Sequence, packetQueueItem.Packet.Data)
		if err != nil {
			return nil, err
		}
	} else {
		calldata, err = callback.ABI.Pack("onPacketAcknowledgement", packetQueueItem.Packet.SourceChannel, packetQueueItem.Packet.SourcePort, packetQueueItem.Packet.Sequence, packetQueueItem.Packet.Data, packetQueueItem.Acknowledgement)
		if err != nil {
			return nil, err
		}
	}

	res, resErr := p.evmKeeper.CallEVMWithData(cachedCtx, common.Address(acc.GetAddress().Bytes()), &target, calldata, true)

	var returnBz []byte
	if resErr == nil {
		returnBz = res.Ret
	} else {
		returnBz = []byte(resErr.Error())
	}

	// emit the event
	if err := p.emitGatewayExecuteEvent(ctx, stateDB, p.Address(), packetQueueItem.Packet.Sequence, retErr == nil, packetQueueItem.OriginalTxHash, true, true, returnBz); err != nil {
		return nil, err
	}

	// only add logs if the call was successful
	if resErr == nil && !res.Failed() {
		logs := evmtypes.LogsToEthereum(res.Logs)
		for _, log := range logs {
			stateDB.AddLog(log)
		}
	}
	writeFn()

	return nil, nil
}

func getSourceCallbackData(ctx sdk.Context, packetQueueItem types.PacketQueueItem) (*callbacktypes.CallbackData, error) {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packetQueueItem.Packet.Data, &data); err != nil {
		return nil, err
	}
	cbData, isCbPacket, err := callbacktypes.GetCallbackData(data, callbacktypes.V1, packetQueueItem.Packet.GetSourcePort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().Limit(), callbacktypes.SourceCallbackKey)
	if isCbPacket {
		if err != nil {
			return nil, err
		}

		return &cbData, nil
	}
	return nil, errors.New("packet is not a callback packet")
}

func (p Precompile) popNextSrcCallback(ctx sdk.Context) (types.PacketQueueItem, error) {
	var (
		packet   types.PacketQueueItem
		sequence uint64
	)
	logger := p.transferKeeper.Logger(ctx)

	if err := p.transferKeeper.SrcCallbackQueue.Walk(ctx, nil, func(key uint64, value types.PacketQueueItem) (bool, error) {
		logger.Info("Processing packet from queue", "key", key, "value", value)
		sequence = key
		packet = value
		return true, nil // stop after first
	}); err != nil {
		return types.PacketQueueItem{}, err
	}

	// remove the packet from the queue
	err := p.transferKeeper.SrcCallbackQueue.Remove(ctx, sequence)
	if err != nil {
		return types.PacketQueueItem{}, err
	}
	return packet, nil
}
