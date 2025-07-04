package keeper

import (
	"context"
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/admin/types"
)

var _ types.MsgServer = &Keeper{}

var ErrNotAuthorized = errors.New("not authorized")
var ErrInvalidRequest = errors.New("invalid request")

func (k Keeper) SetMetadata(
	goCtx context.Context,
	msg *types.MsgSetMetadata,
) (*types.MsgSetMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Metadata == nil {
		return nil, errorsmod.Wrap(ErrInvalidRequest, "metadata is nil")
	}

	params := k.GetParams(ctx)
	accAddr, _ := sdk.AccAddressFromBech32(msg.Authority)
	isACLAdmin := params.Permissions.SetMetadata &&
		k.aclKeeper != nil &&
		k.aclKeeper.Enabled(ctx) &&
		k.aclKeeper.IsAdmin(ctx, accAddr)
	isModuleAuth := msg.Authority == k.GetAuthority()

	if !isACLAdmin && !isModuleAuth {
		return nil, errorsmod.Wrap(ErrNotAuthorized, "authority not permitted")
	}

	k.bankKeeper.SetDenomMetaData(ctx, *msg.Metadata)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetMetadata,
			sdk.NewAttribute(types.AttributeKeyDenom, msg.Metadata.Base),
		),
	)

	return &types.MsgSetMetadataResponse{}, nil
}

func (k Keeper) EnableSetMetadata(goCtx context.Context, msg *types.MsgEnableSetMetadata) (resp *types.MsgEnableSetMetadataResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetAuthority() != msg.Authority {
		return nil, ErrNotAuthorized
	}

	p := k.GetParams(ctx)
	p.Permissions.SetMetadata = true
	k.SetParams(ctx, p)

	return &types.MsgEnableSetMetadataResponse{}, nil
}

func (k Keeper) DisableSetMetadata(goCtx context.Context, msg *types.MsgDisableSetMetadata) (resp *types.MsgDisableSetMetadataResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.GetAuthority() != msg.Authority {
		return nil, ErrNotAuthorized
	}

	p := k.GetParams(ctx)
	p.Permissions.SetMetadata = false
	k.SetParams(ctx, p)

	return &types.MsgDisableSetMetadataResponse{}, nil
}
