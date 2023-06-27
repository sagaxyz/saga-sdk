package acl

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sagaxyz/saga-sdk/x/acl/keeper"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

// NewHandler returns acl module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(_ sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return nil, errorsmod.Wrap(errortypes.ErrUnknownRequest, errMsg)
	}
}
