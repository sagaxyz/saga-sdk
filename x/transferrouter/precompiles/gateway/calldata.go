package gateway

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/contracts"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

// CreateGatewayERC20TransferExecuteCallDataFromPacket creates call data for the gateway execute function from IBC packet data
// This is a convenience function that extracts data from packet and calls CreateGatewayExecuteCallData
// Parameters:
//   - ctx: SDK context
//   - k: keeper instance
//   - packet: IBC packet containing transfer data
//   - data: transfer data from the packet
//
// Returns:
//   - []byte: encoded call data for gateway.execute function
//   - error: any error that occurred during call data creation
func CreateERC20TransferExecuteCallDataFromPacket(
	ctx sdk.Context,
	k keeper.Keeper,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
) ([]byte, error) {
	// TODO: remember to handle denoms differently if this chain was the sender
	// see ReceiverChainIsSource in transfer keeper relay.go
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + data.Denom
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)

	// Create memo with transaction hash
	txHash := tmhash.Sum(ctx.TxBytes())
	txHashHex := hex.EncodeToString(txHash)
	memo, err := json.Marshal(map[string]interface{}{
		"txHash": txHashHex,
	})
	if err != nil {
		k.Logger(ctx).Error("failed to marshal memo", "error", err)
		return nil, fmt.Errorf("failed to marshal memo: %w", err)
	}

	// Call the main function with extracted data
	return createERC20TransferCallData(ctx, k, denomTrace.IBCDenom(), data.Amount, data.Receiver, memo)
}

// createERC20TransferCallData creates call data for the gateway execute function
// This function assembles the call data needed to execute an ERC20 transfer through the gateway
// Parameters:
//   - ctx: SDK context
//   - k: keeper instance
//   - denom: the denomination to transfer (can be IBC denom or regular denom)
//   - amount: the amount to transfer as a string
//   - recipient: the recipient address as a bech32 string
//   - memo: optional memo data (can be nil)
//
// Returns:
//   - []byte: encoded call data for gateway.execute function
//   - error: any error that occurred during call data creation
func createERC20TransferCallData(
	ctx sdk.Context,
	k keeper.Keeper,
	denom string,
	amount string,
	recipient string,
	memo []byte,
) ([]byte, error) {
	// Parse the recipient address
	receiverAccAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		k.Logger(ctx).Error("failed to parse receiver address", "error", err)
		return nil, fmt.Errorf("failed to parse receiver address: %w", err)
	}
	recipientAddrHex := common.BytesToAddress(receiverAccAddr.Bytes())

	// Parse the amount
	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		k.Logger(ctx).Error("failed to parse amount", "amount", amount)
		return nil, fmt.Errorf("failed to parse amount: %s", amount)
	}

	// transfer(address recipient, uint256 amount) → bool
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	erc20CallData, err := erc20.Pack("transfer", recipientAddrHex, amountBig)
	if err != nil {
		k.Logger(ctx).Error("failed to pack ERC20 call data", "error", err)
		return nil, fmt.Errorf("failed to pack ERC20 call data: %w", err)
	}

	return erc20CallData, nil
}
