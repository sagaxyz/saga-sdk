package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
)

var _ types.QueryServer = Keeper{}
