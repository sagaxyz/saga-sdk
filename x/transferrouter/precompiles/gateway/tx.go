// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"errors"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/ibc"
	evmostypes "github.com/evmos/evmos/v20/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmante "github.com/evmos/evmos/v20/x/evm/ante"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
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
	// Parse the execute arguments
	p.transferKeeper.Logger(ctx).Info("=== GATEWAY EXECUTE START ===", "origin", origin.Hex(), "contract", contract.Address().Hex())

	var packet channeltypes.Packet
	var sequence uint64
	err := p.transferKeeper.PacketQueue.Walk(ctx, nil, func(key uint64, value channeltypes.Packet) (stop bool, err error) {
		p.transferKeeper.Logger(ctx).Info("Processing packet from queue", "key", key, "value", value)
		packet = value
		sequence = key
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	// get next packet data
	p.transferKeeper.Logger(ctx).Info("Retrieved packet from queue", "sequence", sequence, "packetSourcePort", packet.GetSourcePort(), "packetDestPort", packet.GetDestPort())

	// parse the packet data
	var packetData transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &packetData); err != nil {
		p.transferKeeper.Logger(ctx).Error("Failed to unmarshal packet data", "error", err)
		return nil, err
	}
	p.transferKeeper.Logger(ctx).Info("Parsed packet data", "denom", packetData.Denom, "amount", packetData.Amount, "sender", packetData.Sender, "receiver", packetData.Receiver)

	cachedCtx, writeFn := ctx.CacheContext()
	cachedCtx = evmante.BuildEvmExecutionCtx(cachedCtx)

	success := false
	// This defer is used so we can write the cached context events back to the main context, but also to clear the returned error,
	// so it can still remove the packet from the queue if the execution fails.
	defer func() {
		if retErr == nil {
			// Write cachedCtx events back to ctx.
			writeFn()
			p.transferKeeper.Logger(ctx).Info("Wrote cached context events back to main context")

		}

		retErr = nil

		// delete the packet from the queue
		err = p.transferKeeper.PacketQueue.Remove(ctx, sequence)
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("failed to delete packet from queue", "error", err)
			retErr = err
			return
		}

		var ack channeltypes.Acknowledgement
		if success {
			p.transferKeeper.Logger(ctx).Info("Receipt status successful, creating result acknowledgement")
			ack = channeltypes.NewResultAcknowledgement([]byte{1})
		} else {
			p.transferKeeper.Logger(ctx).Info("Receipt status unsuccessful, creating error acknowledgement")
			ack = channeltypes.NewErrorAcknowledgement(errors.New("failed to execute call"))
		}
		p.transferKeeper.Logger(ctx).Info("Created acknowledgment", "ack", ack, "receipt", success)

		err = p.transferKeeper.WriteAcknowledgementForPacket(ctx, packet, packetData, ack)
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("failed to write IBC acknowledgment", "error", err)
			retErr = err
			return
		}
		p.transferKeeper.Logger(ctx).Info("Successfully wrote IBC acknowledgment")
	}()

	// Check if the token pair exists and get the ERC20 contract address
	// for the native ERC20 or the precompile.
	// This call fails if the token does not exist or is not registered.
	coin := ibc.GetReceivedCoin(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), packetData.Denom, packetData.Amount)
	p.transferKeeper.Logger(ctx).Info("Retrieved coin from IBC", "denom", coin.Denom, "amount", coin.Amount)

	tokenPairID := p.transferKeeper.Erc20Keeper.GetTokenPairID(ctx, coin.Denom)
	tokenPair, found := p.transferKeeper.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		p.transferKeeper.Logger(ctx).Error("Token pair not found", "denom", packetData.Denom, "tokenPairID", tokenPairID)
		return nil, errorsmod.Wrapf(erc20types.ErrTokenPairNotFound, "token pair for denom %s not found", packetData.Denom)
	}
	p.transferKeeper.Logger(ctx).Info("Found token pair", "tokenPairID", tokenPairID, "erc20Contract", tokenPair.GetERC20Contract().Hex())

	var callData []byte
	var fromAddress common.Address

	fromAddress = p.Address() // TODO: for now we use the gateway contract address as the from address
	p.transferKeeper.Logger(ctx).Info("Set initial from address", "fromAddress", fromAddress.Hex())

	// if the packet is a callback packet we process it as such, if not, we assume it's a normal erc20 transfer
	cbData, isCbPacket, err := callbacktypes.GetCallbackData(packetData, callbacktypes.V1, packet.GetDestPort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().Limit(), callbacktypes.DestinationCallbackKey)
	if isCbPacket {
		p.transferKeeper.Logger(ctx).Info("Processing callback packet")
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("failed to get callback data", "error", err)
			retErr = err
			return
		}
		p.transferKeeper.Logger(ctx).Info("Successfully retrieved callback data", "callbackAddress", cbData.CallbackAddress, "senderAddress", cbData.SenderAddress, "commitGasLimit", cbData.CommitGasLimit)

		target := common.HexToAddress(cbData.CallbackAddress)
		p.transferKeeper.Logger(ctx).Info("Set callback target address", "target", target.Hex())

		// Generate secure isolated address from sender, we know this address is initialized in the IBC OnRecvPacket
		isolatedAddr := utils.GenerateIsolatedAddress(packet.GetDestChannel(), cbData.SenderAddress)
		fromAddress = common.BytesToAddress(isolatedAddr.Bytes())
		p.transferKeeper.Logger(ctx).Info("Generated isolated address", "isolatedAddress", fromAddress.Hex(), "destChannel", packet.GetDestChannel())

		cachedCtx = cachedCtx.WithGasMeter(evmostypes.NewInfiniteGasMeterWithLimit(cbData.CommitGasLimit))
		p.transferKeeper.Logger(ctx).Info("Created cached context with gas limit", "gasLimit", cbData.CommitGasLimit)

		amountInt, ok := math.NewIntFromString(packetData.Amount)
		if !ok {
			p.transferKeeper.Logger(ctx).Error("Failed to parse amount", "amount", packetData.Amount)
			return nil, errors.New("error when parsing amount")
		}
		p.transferKeeper.Logger(ctx).Info("Parsed amount", "amount", amountInt.String())

		erc20 := contracts.ERC20MinterBurnerDecimalsContract

		remainingGas := math.NewIntFromUint64(cachedCtx.GasMeter().GasRemaining()).BigInt()
		p.transferKeeper.Logger(ctx).Info("Starting ERC20 approve call", "remainingGas", remainingGas.String(), "fromAddress", fromAddress.Hex(), "tokenContract", tokenPair.GetERC20Contract().Hex(), "target", target.Hex(), "amount", amountInt.String())

		// Call the EVM with the remaining gas as the maximum gas limit.
		// Up to now, the remaining gas is equal to the callback gas limit set by the user.
		// NOTE: use the cached ctx for the EVM calls.
		res, err := p.evmKeeper.CallEVM(cachedCtx, erc20.ABI, fromAddress, tokenPair.GetERC20Contract(), true, "approve", target, amountInt.BigInt())
		if err != nil {
			p.transferKeeper.Logger(ctx).Error("ERC20 approve call failed", "error", err)
			return nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance: %v", err)
		}
		p.transferKeeper.Logger(ctx).Info("ERC20 approve call completed", "gasUsed", res.GasUsed, "success", !res.Failed())

		// Consume the actual used gas on the original callback context.
		ctx.GasMeter().ConsumeGas(res.GasUsed, "callback allowance")
		remainingGas = remainingGas.Sub(remainingGas, math.NewIntFromUint64(res.GasUsed).BigInt())
		p.transferKeeper.Logger(ctx).Info("Consumed gas for approve", "gasUsed", res.GasUsed, "remainingGas", remainingGas.String())
		if ctx.GasMeter().IsOutOfGas() || remainingGas.Cmp(big.NewInt(0)) < 0 {
			p.transferKeeper.Logger(ctx).Error("Out of gas after approve", "remainingGas", remainingGas.String())
			return nil, errorsmod.Wrapf(errortypes.ErrOutOfGas, "out of gas")
		}

		var approveSuccess bool
		err = erc20.ABI.UnpackIntoInterface(&approveSuccess, "approve", res.Ret)
		if err != nil {
			return nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to unpack approve return: %v", err)
		}

		if !approveSuccess {
			return nil, errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance")
		}
		p.transferKeeper.Logger(ctx).Info("Approve call successful")
		// NOTE: use the cached ctx for the EVM calls.
		p.transferKeeper.Logger(ctx).Info("Starting callback EVM call", "fromAddress", fromAddress.Hex(), "target", target.Hex(), "calldataLength", len(cbData.Calldata))
		res, err = p.evmKeeper.CallEVMWithData(cachedCtx, fromAddress, &target, cbData.Calldata, true)
		if err != nil {
			return nil, errorsmod.Wrapf(ErrEVMCallFailed, "EVM returned error: %s", err.Error())
		}
		p.transferKeeper.Logger(ctx).Info("Callback EVM call completed", "gasUsed", res.GasUsed, "success", !res.Failed())

		// Consume the actual gas used on the original callback context.
		ctx.GasMeter().ConsumeGas(res.GasUsed, "callback function")
		if ctx.GasMeter().IsOutOfGas() {
			p.transferKeeper.Logger(ctx).Error("Out of gas after callback function")
			return nil, errorsmod.Wrapf(errortypes.ErrOutOfGas, "out of gas")
		}
		p.transferKeeper.Logger(ctx).Info("Consumed gas for callback function", "gasUsed", res.GasUsed)

		// Check that the sender no longer has tokens after the callback.
		// NOTE: contracts must implement an IERC20(token).transferFrom(msg.sender, address(this), amount)
		// for the total amount, or the callback will fail.
		// This check is here to prevent funds from getting stuck in the isolated address,
		// since they would become irretrievable.
		p.transferKeeper.Logger(ctx).Info("Checking token balance after callback", "fromAddress", fromAddress.Hex(), "tokenContract", tokenPair.GetERC20Contract().Hex())
		receiverTokenBalance := p.transferKeeper.Erc20Keeper.BalanceOf(ctx, erc20.ABI, tokenPair.GetERC20Contract(), fromAddress) // here,
		// we can use the original ctx and skip manually adding the gas
		p.transferKeeper.Logger(ctx).Info("Token balance after callback", "balance", receiverTokenBalance.String())
		if receiverTokenBalance.Cmp(big.NewInt(0)) != 0 {
			p.transferKeeper.Logger(ctx).Error("Receiver still has tokens after callback", "balance", receiverTokenBalance.String())
			return nil, errorsmod.Wrapf(erc20types.ErrEVMCall,
				"receiver has %d unrecoverable tokens after callback", receiverTokenBalance)
		}
		p.transferKeeper.Logger(ctx).Info("Callback processing completed successfully")

		return res.Ret, nil
	}

	// if it doesn't have a callback we handle it as a normal erc20 transfer
	p.transferKeeper.Logger(ctx).Info("Processing normal ERC20 transfer (non-callback)")
	callData, err = CreateERC20TransferExecuteCallDataFromPacket(ctx, p.transferKeeper, packet, packetData)
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("Failed to create gateway execute call data", "error", err)
		return nil, errorsmod.Wrapf(ErrEVMCallFailed, "failed to create gateway execute call data: %v", err)
	}
	p.transferKeeper.Logger(ctx).Info("Created ERC20 transfer call data", "callDataLength", len(callData))

	// Execute the call logic here
	// This is where you would call your keeper methods to perform the actual execution
	nonce, err := p.transferKeeper.AccountKeeper.GetSequence(ctx, p.Address().Bytes())
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("Failed to get account sequence", "error", err)
		return nil, err
	}
	p.transferKeeper.Logger(ctx).Info("Retrieved account sequence", "nonce", nonce)

	fromAddress = common.BytesToAddress(fromAddress.Bytes())
	target := tokenPair.GetERC20Contract()
	p.transferKeeper.Logger(ctx).Info("Creating EVM message", "fromAddress", fromAddress.Hex(), "target", target.Hex(), "nonce", nonce, "gasLimit", 6000000)

	p.transferKeeper.Logger(ctx).Info("Applying EVM message")
	result, err := p.evmKeeper.CallEVMWithData(cachedCtx, fromAddress, &target, callData, true)
	if err != nil {
		p.transferKeeper.Logger(ctx).Error("EVM message application failed", "error", err)
		return nil, err
	}
	p.transferKeeper.Logger(ctx).Info("EVM message applied", "gasUsed", result.GasUsed, "failed", result.Failed())

	if !result.Failed() {
		p.transferKeeper.Logger(ctx).Info("Adding EVM logs to state", "logCount", len(result.Logs))
		for _, log := range result.Logs {
			stateDB.AddLog(log.ToEthereum())
		}
	} else {
		p.transferKeeper.Logger(ctx).Error("EVM message application failed", "error", result.Logs, "result", result, "vmerror", result.VmError)
	}

	success = !result.Failed()
	p.transferKeeper.Logger(ctx).Info("Set execution success status", "success", success)

	// Emit the gateway execute event
	// if err := p.emitGatewayExecuteEvent(ctx, stateDB, p.Address(), p.Address(), executeArgs.Target, executeArgs.Value, executeArgs.Data, executeArgs.Note, !result.Failed(), result.Ret); err != nil {
	// return nil, err
	// }

	p.transferKeeper.Logger(ctx).Info("=== GATEWAY EXECUTE COMPLETE ===", "success", success)
	return []byte{}, nil
}

// EmitNote handles the emitNote method for emitting metadata notes
func (p Precompile) EmitNote(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse the emitNote arguments
	var noteArgs emitNote
	if err := method.Inputs.Copy(&noteArgs, args); err != nil {
		return nil, fmt.Errorf("error parsing emitNote arguments: %w", err)
	}

	// Validate the note arguments
	if err := validateNoteArgs(noteArgs); err != nil {
		return nil, err
	}

	// if err := p.emitNoteEvent(ctx, stateDB, p.Address(), sender, noteArgs.Ref, noteArgs.Data); err != nil {
	// 	return nil, err
	// }

	return method.Outputs.Pack()
}

// validateNoteArgs validates the note arguments
func validateNoteArgs(args emitNote) error {
	if args.Ref == [32]byte{} {
		return fmt.Errorf(ErrInvalidRef)
	}
	return nil
}

// Pause handles the pause method for pausing the contract
func (p Precompile) Pause(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// TODO: Implement pause logic
	// This would typically involve calling keeper methods to pause the contract
	return method.Outputs.Pack()
}

// Unpause handles the unpause method for unpausing the contract
func (p Precompile) Unpause(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// TODO: Implement unpause logic
	// This would typically involve calling keeper methods to unpause the contract
	return method.Outputs.Pack()
}
