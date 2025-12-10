package ante

import (
	"fmt"
	"strings"

	errors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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

// Rejects tx if any matching message does not pass the filter fn for every signer.
func (mvfd MsgFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	// Check if at least one Msg matches any of the filtered prefixes
	var match bool
	for _, msg := range tx.GetMsgs() {
		msgType := sdk.MsgTypeURL(msg)

		for _, prefix := range mvfd.prefixes {
			if strings.HasPrefix(msgType, prefix) {
				match = true
				break
			}
		}
		if match {
			break
		}
	}

	if match {
		var signers [][]byte
		signers, err = sigTx.GetSigners()
		if err != nil {
			return
		}
		for _, signer := range signers {
			if !mvfd.filter(ctx, signer) {
				err = fmt.Errorf("address %s denied for some of the tx message(s)", sdk.AccAddress(signer).String())
				return
			}
		}
	}

	return next(ctx, tx, simulate)
}
