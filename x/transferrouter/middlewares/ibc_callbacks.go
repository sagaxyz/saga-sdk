package middlewares

import (
	"bytes"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/evmos/evmos/v20/ibc"
	evmostypes "github.com/evmos/evmos/v20/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmante "github.com/evmos/evmos/v20/x/evm/ante"
	callbacktypes "github.com/sagaxyz/saga-sdk/x/transferrouter/v10types"
)

const (
	// ModuleName for future compatibility with the original IBC callbacks module
	ModuleName = "ibc-callbacks"
)

// EVM sentinel callback errors
var (
	ErrInvalidReceiverAddress = errorsmod.Register(ModuleName, 1, "invalid receiver address")
	ErrCallbackFailed         = errorsmod.Register(ModuleName, 2, "callback failed")
	ErrInvalidCalldata        = errorsmod.Register(ModuleName, 3, "invalid calldata in callback data")
	ErrContractHasNoCode      = errorsmod.Register(ModuleName, 4, "contract has no code")
	ErrTokenPairNotFound      = errorsmod.Register(ModuleName, 5, "token not registered")
	ErrNumberOverflow         = errorsmod.Register(ModuleName, 6, "number overflow")
	ErrAllowanceFailed        = errorsmod.Register(ModuleName, 7, "allowance failed")
	ErrEVMCallFailed          = errorsmod.Register(ModuleName, 8, "evm call failed")
	ErrOutOfGas               = errorsmod.Register(ModuleName, 9, "out of gas")
)

// GenerateIsolatedAddress generates an isolated address for the given channel ID and sender address.
// This provides a safe address to call the receiver contract address with custom calldata
func GenerateIsolatedAddress(channelID string, sender string) sdk.AccAddress {
	return sdk.AccAddress(address.Module(ModuleName, []byte(channelID), []byte(sender))[:20])
}

// IBCReceivePacketCallback handles IBC packet callbacks for cross-chain contract execution.
// This function processes incoming IBC packets that contain callback data and executes
// the specified contract with the transferred tokens.
//
// The function performs the following operations:
// 1. Unmarshals and validates the IBC packet data
// 2. Extracts callback data from the packet
// 3. Generates an isolated address for security
// 4. Validates the receiver address matches the isolated address
// 5. Verifies the target contract exists and contains code
// 6. Sets up ERC20 token allowance for the contract
// 7. Executes the callback function on the target contract
// 8. Validates that all tokens were successfully transferred to the contract
//
// Returns:
//   - error: Returns nil on success, or an error if any step fails including:
//   - Packet data unmarshaling errors
//   - Invalid callback data
//   - Address validation failures
//   - Contract validation failures (non-existent or no code)
//   - Token pair registration errors
//   - EVM execution errors
//   - Gas limit exceeded errors
//   - Token transfer validation failures
//
// Security Notes:
//   - Uses isolated addresses to prevent unauthorized access
//   - Validates contract existence to prevent fund loss
//   - Enforces gas limits to prevent DoS attacks
//   - Requires contracts to implement proper token transfer logic
//   - Validates final token balances to ensure successful transfers
func (i IBCMiddleware) IBCReceivePacketCallback(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	contractAddress string,
	version string,
) error {

	fmt.Println("IBCReceivePacketCallback called?!?!?!?!?!?!?!?!?")
	data, err := callbacktypes.UnmarshalPacketData(packet.GetData(), version, "")
	if err != nil {
		return err
	}

	cbData, isCbPacket, err := callbacktypes.GetCallbackData(data, version, packet.GetDestPort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().GasRemaining(), callbacktypes.DestinationCallbackKey)
	if err != nil {
		return err
	}
	if !isCbPacket {
		return nil
	}

	// `ProcessCallback` in IBC-Go overrides the infinite gas meter with a basic gas meter,
	// so we need to generate a new infinite gas meter to run the EVM executions on.
	// Skipping this causes the EVM gas estimation function to deplete all Cosmos gas.
	// We re-add the actual EVM call gas used to the original context after the call is complete
	// with the gas retrieved from the EVM message result.
	cachedCtx, writeFn := ctx.CacheContext()
	cachedCtx = evmante.BuildEvmExecutionCtx(cachedCtx).
		WithGasMeter(evmostypes.NewInfiniteGasMeterWithLimit(cbData.CommitGasLimit))

	// receiver := sdk.MustAccAddressFromBech32(data.Receiver)
	receiver, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidReceiverAddress,
			"acc addr from bech32 conversion failed for receiver address: %s", data.Receiver)
	}

	receiverHex := common.BytesToAddress(receiver.Bytes())

	// Generate secure isolated address from sender.
	isolatedAddr := GenerateIsolatedAddress(packet.GetDestChannel(), data.Sender)
	isolatedAddrHex := common.BytesToAddress(isolatedAddr.Bytes())

	// Ensure receiver address is equal to the isolated address.
	if !bytes.Equal(receiverHex.Bytes(), isolatedAddrHex.Bytes()) {
		return errorsmod.Wrapf(ErrInvalidReceiverAddress, "expected %s, got %s", isolatedAddrHex.String(), receiverHex.String())
	}

	if i.k.AccountKeeper.GetAccount(ctx, receiver) == nil {
		acc := i.k.AccountKeeper.NewAccountWithAddress(ctx, receiver)
		i.k.AccountKeeper.SetAccount(ctx, acc)
	}

	contractAddr := common.HexToAddress(contractAddress)
	contractAccount := i.k.EVMKeeper.GetAccountOrEmpty(ctx, contractAddr)

	// Check if the contract address contains code.
	// This check is required because if there is no code, the call will still pass on the EVM side,
	// but it will ignore the calldata and funds may get stuck.
	if !contractAccount.IsContract() {
		return errorsmod.Wrapf(ErrContractHasNoCode, "provided contract address is not a contract: %s", contractAddr)
	}

	// Check if the token pair exists and get the ERC20 contract address
	// for the native ERC20 or the precompile.
	// This call fails if the token does not exist or is not registered.
	coin := ibc.GetReceivedCoin(packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), data.Token.Denom.Path(), data.Token.Amount)

	tokenPairID := i.k.Erc20Keeper.GetTokenPairID(ctx, coin.Denom)
	tokenPair, found := i.k.Erc20Keeper.GetTokenPair(ctx, tokenPairID)
	if !found {
		return errorsmod.Wrapf(ErrTokenPairNotFound, "token pair for denom %s not found", data.Token.Denom.IBCDenom())
	}
	amountInt, ok := math.NewIntFromString(data.Token.Amount)
	if !ok {
		return errorsmod.Wrapf(ErrNumberOverflow, "amount overflow")
	}

	erc20 := contracts.ERC20MinterBurnerDecimalsContract

	remainingGas := math.NewIntFromUint64(cachedCtx.GasMeter().GasRemaining()).BigInt()

	// Call the EVM with the remaining gas as the maximum gas limit.
	// Up to now, the remaining gas is equal to the callback gas limit set by the user.
	// NOTE: use the cached ctx for the EVM calls.
	res, err := i.k.EVMKeeper.CallEVM(cachedCtx, erc20.ABI, receiverHex, tokenPair.GetERC20Contract(), true, "approve", contractAddr, amountInt.BigInt())
	if err != nil {
		return errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance: %v", err)
	}

	// Consume the actual used gas on the original callback context.
	ctx.GasMeter().ConsumeGas(res.GasUsed, "callback allowance")
	remainingGas = remainingGas.Sub(remainingGas, math.NewIntFromUint64(res.GasUsed).BigInt())
	if ctx.GasMeter().IsOutOfGas() || remainingGas.Cmp(big.NewInt(0)) < 0 {
		return errorsmod.Wrapf(ErrOutOfGas, "out of gas")
	}

	var approveSuccess bool
	err = erc20.ABI.UnpackIntoInterface(&approveSuccess, "approve", res.Ret)
	if err != nil {
		return errorsmod.Wrapf(ErrAllowanceFailed, "failed to unpack approve return: %v", err)
	}

	if !approveSuccess {
		return errorsmod.Wrapf(ErrAllowanceFailed, "failed to set allowance")
	}

	// NOTE: use the cached ctx for the EVM calls.
	res, err = i.k.EVMKeeper.CallEVMWithData(cachedCtx, receiverHex, &contractAddr, cbData.Calldata, true)
	if err != nil {
		return errorsmod.Wrapf(ErrEVMCallFailed, "EVM returned error: %s", err.Error())
	}

	// Consume the actual gas used on the original callback context.
	ctx.GasMeter().ConsumeGas(res.GasUsed, "callback function")
	if ctx.GasMeter().IsOutOfGas() {
		return errorsmod.Wrapf(ErrOutOfGas, "out of gas")
	}

	// Write cachedCtx events back to ctx.
	writeFn()

	// Check that the sender no longer has tokens after the callback.
	// NOTE: contracts must implement an IERC20(token).transferFrom(msg.sender, address(this), amount)
	// for the total amount, or the callback will fail.
	// This check is here to prevent funds from getting stuck in the isolated address,
	// since they would become irretrievable.
	receiverTokenBalance := i.k.Erc20Keeper.BalanceOf(ctx, erc20.ABI, tokenPair.GetERC20Contract(), receiverHex) // here,
	// we can use the original ctx and skip manually adding the gas
	if receiverTokenBalance.Cmp(big.NewInt(0)) != 0 {
		return errorsmod.Wrapf(erc20types.ErrEVMCall,
			"receiver has %d unrecoverable tokens after callback", receiverTokenBalance)
	}

	return nil
}
