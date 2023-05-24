package dac

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sagaxyz/sagaevm/v8/x/dac/keeper"
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

// NewHandler returns dac module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(errortypes.ErrUnknownRequest, errMsg)
	}
}
