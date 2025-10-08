package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

var _ types.QueryServer = Keeper{}
