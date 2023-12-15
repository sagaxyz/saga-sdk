package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

    keepertest "github.com/sagaxyz/sagaos/testutil/keeper"
    "github.com/sagaxyz/sagaos/x/chainlet/types"
    "github.com/sagaxyz/sagaos/x/chainlet/keeper"
)

func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, sdk.Context) {
	k, ctx := keepertest.ChainletKeeper(t)
	return k, keeper.NewMsgServerImpl(k), ctx
}

func TestMsgServer(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
	require.NotEmpty(t, k)
}