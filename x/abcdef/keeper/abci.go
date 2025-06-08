package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) BeginBlock(ctx sdk.Context) {
}

func (k *Keeper) EndBlock(ctx sdk.Context) {
	err := k.Send(ctx)
    if err != nil {
        ctx.Logger().Error("failed to send IBC packet", "error", err)
    }
}
