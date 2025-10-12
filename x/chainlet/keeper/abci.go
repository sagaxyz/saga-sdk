package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) BeginBlock(ctx context.Context) error {
	return nil
}

func (k *Keeper) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.Send(ctx)
	if err != nil {
		k.Logger(sdkCtx).Error(fmt.Sprintf("send failed: %s", err))
		//return err
	}

	return nil
}
