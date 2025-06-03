package middleware

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	for _, msg := range tx.GetMsgs() {
		msgType := sdk.MsgTypeURL(msg)

		if msgType == "/ibc.applications.transfer.v1.MsgTransfer" {
			msg := msg.(*ibctransfertypes.MsgTransfer)

			supported, err := ah.k.SupportedAssets.Has(ctx, collections.Join(
				msg.SourceChannel,
				msg.Token.Denom,
			))

			if err != nil {
				return ctx, err
			}

			if !supported {
				return ctx, errors.New("asset not supported on the target chainlet")
			}
		}
	}

	return next(ctx, tx, simulate)
}
