package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	cstore := sdkCtx.KVStore(k.storeKey)
	val := cstore.Get([]byte("test-key"))
	if val == nil {
		fmt.Printf("XXX BeginBlock: nil\n") 
	} else {
		fmt.Printf("XXX BeginBlock: %s\n", string(val))
	}
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
