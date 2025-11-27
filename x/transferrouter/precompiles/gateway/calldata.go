package gateway

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

// CreateERC20TransferCallData creates call data for the gateway execute function
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
func CreateERC20TransferCallData(
	ctx sdk.Context,
	k keeper.Keeper,
	amount string,
	recipient string,
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
