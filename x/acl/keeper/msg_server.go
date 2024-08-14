package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

var _ types.MsgServer = &Keeper{}

var ErrNotAuthorized = errors.New("not authorized")

func (k Keeper) AddAllowed(goCtx context.Context, msg *types.MsgAddAllowed) (resp *types.MsgAddAllowedResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return
	}
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	for _, addr := range msg.Allowed {
		var accAddr sdk.AccAddress
		accAddr, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return
		}
		store.Set(accAddr, []byte{0}) //TODO
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
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAllowed)
	for _, addr := range msg.Allowed {
		var accAddr sdk.AccAddress
		accAddr, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return
		}
		store.Delete(accAddr)
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
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	for _, addr := range msg.Admins {
		var accAddr sdk.AccAddress
		accAddr, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return
		}
		store.Set(accAddr, []byte{0}) //TODO
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
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
		err = ErrNotAuthorized
		return
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAdmins)
	for _, addr := range msg.Admins {
		var accAddr sdk.AccAddress
		accAddr, err = sdk.AccAddressFromBech32(addr)
		if err != nil {
			return
		}
		store.Delete(accAddr)
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
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
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
	if !k.Admin(ctx, sender) && k.GetAuthority() != msg.GetSigners()[0].String() {
		err = ErrNotAuthorized
		return
	}

	k.paramSpace.Set(ctx, types.ParamStoreKeyEnable, false)

	resp = &types.MsgDisableResponse{}
	return
}
