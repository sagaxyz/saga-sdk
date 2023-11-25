package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetRecipient(ctx sdk.Context) string {
	params := k.GetParams(ctx)
	return params.Recipient
}

func (k Keeper) TransferFees(ctx sdk.Context) {
	params := k.GetParams(ctx)
	if !params.Enabled {
		return
	}
	// Since this is called in BeginBlock, collected fees will be from the previous block
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())

	recipient := sdk.MustAccAddressFromBech32(params.Recipient)

	// Transfer collected fees to the recipient account
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, recipient, feesCollectedInt)
	if err != nil {
		panic(err)
	}
}
