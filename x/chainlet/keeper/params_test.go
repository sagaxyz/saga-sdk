package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

    keepertest "github.com/sagaxyz/saga-sdk/testutil/keeper"
    "github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.ChainletKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
