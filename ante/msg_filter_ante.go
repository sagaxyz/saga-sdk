package ante

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgFilterDecorator struct {
	filter   FilterFn
	prefixes []string
}

func NewMsgFilterDecorator(fn FilterFn, prefixes ...string) MsgFilterDecorator {
	return MsgFilterDecorator{
		filter:   fn,
		prefixes: prefixes,
	}
}

// Denies messages where its type matches any of the provided prefixes and does not pass the filter fn for every signer.
func (mvfd MsgFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		msgType := sdk.MsgTypeURL(msg)

		for _, prefix := range mvfd.prefixes {
			if !strings.HasPrefix(msgType, prefix) {
				continue
			}

			for _, signer := range msg.GetSigners() {
				if !mvfd.filter(ctx, signer) {
					err = fmt.Errorf("address %s denied for message type %s", signer.String(), msgType)
					return
				}
			}

			// No need to check other matching prefixes as the messages passed the check already
			break
		}
	}

	return next(ctx, tx, simulate)
}
