package middleware

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
)

type AssetControlAnteHandler struct {
	k *keeper.Keeper
}

func NewAssetControlAnteHandler(k *keeper.Keeper) AssetControlAnteHandler {
	return AssetControlAnteHandler{
		k: k,
	}
}

var _ sdk.AnteDecorator = AssetControlAnteHandler{}

// Rejects tx if there are MsgTransfer messages and the asset is not supported on the target chainlet.
func (ah AssetControlAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	err = ah.checkMsgs(ctx, tx.GetMsgs(), 0)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// checkMsgs is recursive in case there are nested authz messages, with a hard limit of 3 levels.
func (ah AssetControlAnteHandler) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, level int) error {
	if level >= 3 {
		return errors.New("nested authz messages too deep")
	}

	ibcTransferMsgType := sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{})
	authzMsgType := sdk.MsgTypeURL(&authztypes.MsgExec{})

	for _, msg := range msgs {
		msgType := sdk.MsgTypeURL(msg)

		if msgType == ibcTransferMsgType {
			msg := msg.(*ibctransfertypes.MsgTransfer)

			supported, err := ah.k.SupportedAssets.Has(ctx, collections.Join(
				msg.SourceChannel,
				msg.Token.Denom,
			))

			if err != nil {
				return err
			}

			if !supported {
				return errors.New("asset not supported on the target chainlet")
			}
		} else if msgType == authzMsgType {
			msg := msg.(*authztypes.MsgExec)
			authzMsgs, err := msg.GetMessages()
			if err != nil {
				return err
			}

			err = ah.checkMsgs(ctx, authzMsgs, level+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
