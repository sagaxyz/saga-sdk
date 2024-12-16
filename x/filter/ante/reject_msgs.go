package cosmos

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sagaxyz/saga-sdk/x/filter/keeper"
)

// RejectMessagesDecorator rejects filtered msg type prefixes
type RejectMessagesDecorator struct {
	filterKeeper keeper.Keeper
}

func NewRejectMessagesDecorator(filterKeeper keeper.Keeper) RejectMessagesDecorator {
	return RejectMessagesDecorator{
		filterKeeper: filterKeeper,
	}
}

func (rmd RejectMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	prefixes := rmd.filterKeeper.GetParams(ctx).Prefixes
	for _, msg := range tx.GetMsgs() {
		msgType := sdk.MsgTypeURL(msg)

		for _, prefix := range prefixes {
			if strings.HasPrefix(msgType, prefix) {
				return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType, "Message type '%s' rejected: matching filter module prefix rule '%s'", msgType, prefix)
			}
		}
	}

	return next(ctx, tx, simulate)
}
