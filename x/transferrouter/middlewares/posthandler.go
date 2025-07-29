package middlewares

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

var _ sdk.PostDecorator = PostHandler{}

type PostHandler struct {
	keeper keeper.Keeper
}

func NewPostHandler(k keeper.Keeper) PostHandler {
	return PostHandler{
		keeper: k,
	}
}

// PostHandle implements types.PostDecorator.
func (p PostHandler) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	// Here we need to find the corresponding call in the call queue and remove it, also we need to write the acknowledgment if needed

	// 1. Find the corresponding call in the call queue

	// 2. Remove the call from the call queue

	// 3. Write the IBC acknowledgment if needed

	return next(ctx, tx, simulate, success)
}
