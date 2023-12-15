package keeper

import (
	"github.com/sagaxyz/sagaos/x/chainlet/types"
)

var _ types.QueryServer = Keeper{}
