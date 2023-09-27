package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	//defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.TransferFees(ctx)
}

func (k *Keeper) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) {
}
