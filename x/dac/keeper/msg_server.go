package keeper

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/dac/types"
)

var _ types.MsgServer = &Keeper{}

var ErrNotAuthorized = errors.New("not authorized")

func (k Keeper) AddAllowed(goCtx context.Context, msg *types.MsgAddAllowed) (resp *types.MsgAddAllowedResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	for _, addr := range msg.Allowed {
		store.Set(addr.Bytes(), []byte{byte(addr.Format)})
	}

	resp = &types.MsgAddAllowedResponse{}
	return
}
func (k Keeper) RemoveAllowed(goCtx context.Context, msg *types.MsgRemoveAllowed) (resp *types.MsgRemoveAllowedResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	for _, addr := range msg.Allowed {
		store.Delete(addr.Bytes())
	}
	resp = &types.MsgRemoveAllowedResponse{}
	return
}

func (k Keeper) AddAdmins(goCtx context.Context, msg *types.MsgAddAdmins) (resp *types.MsgAddAdminsResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	for _, addr := range msg.Admins {
		store.Set(addr.Bytes(), []byte{byte(addr.Format)})
	}

	resp = &types.MsgAddAdminsResponse{}
	return
}
func (k Keeper) RemoveAdmins(goCtx context.Context, msg *types.MsgRemoveAdmins) (resp *types.MsgRemoveAdminsResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	for _, addr := range msg.Admins {
		store.Delete(addr.Bytes())
	}

	resp = &types.MsgRemoveAdminsResponse{}
	return
}

func (k Keeper) Enable(goCtx context.Context, msg *types.MsgEnable) (resp *types.MsgEnableResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	k.paramSpace.Set(ctx, types.ParamStoreKeyEnable, true)

	resp = &types.MsgEnableResponse{}
	return
}
func (k Keeper) Disable(goCtx context.Context, msg *types.MsgDisable) (resp *types.MsgDisableResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) {
		err = ErrNotAuthorized
		return
	}

	k.paramSpace.Set(ctx, types.ParamStoreKeyEnable, false)

	resp = &types.MsgDisableResponse{}
	return
}