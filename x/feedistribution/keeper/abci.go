package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) BeginBlock(ctx context.Context) error {
	k.TransferFees(sdk.UnwrapSDKContext(ctx))

	return nil
}

func (k *Keeper) EndBlock(ctx context.Context) error {
	return nil
}
