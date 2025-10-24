package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

func TestQuerier_Params_Success(t *testing.T) {
	ctx, k, _, _, _ := setupKeeperWithMocks(t)

	querier := keeper.NewQuerier(k)

	// Params were set in setupKeeperWithMocks
	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Params.Enabled)
	require.Equal(t, "0x5A6A8Ce46E34c2cd998129d013fA0253d3892345", resp.Params.GatewayContractAddress)
}

func TestQuerier_Params_NotFound(t *testing.T) {
	ctx, k, _ := setupKeeperForTest(t)

	querier := keeper.NewQuerier(k)

	// Params not set, should return error
	resp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
}
