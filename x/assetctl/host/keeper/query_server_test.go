package keeper_test

import (
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/sagaxyz/saga-sdk/x/assetctl"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
	"github.com/stretchr/testify/require"
)

func setupQueryServer(t *testing.T) (sdk.Context, *keeper.Keeper, types.QueryServer) {
	keys := storetypes.NewKVStoreKeys(
		assetctltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(assetctl.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	storeService := runtime.NewKVStoreService(keys[assetctltypes.StoreKey])

	ctx := sdk.NewContext(cms, tmproto.Header{}, true, logger)

	var addressCodec address.Codec = nil // Use nil or a mock if not available

	// Create mock keepers
	mockACLKeeper := MockACLKeeper{}
	mockAccountKeeper := MockAccountKeeper{}
	mockICAControllerKeeper := MockICAControllerKeeper{}
	mockERC20Keeper := MockERC20Keeper{}
	mockBankKeeper := MockBankKeeper{}

	k := keeper.NewKeeper(
		storeService,
		cdc,
		logger,
		addressCodec,
		baseapp.NewMsgServiceRouter(),
		mockACLKeeper,
		mockAccountKeeper,
		mockICAControllerKeeper,
		mockERC20Keeper,
		mockBankKeeper,
		"cosmos1test",
	)

	queryServer := keeper.NewQueryServerImpl(*k)

	return ctx, k, queryServer
}

func TestQueryServer(t *testing.T) {
	ctx, k, queryServer := setupQueryServer(t)
	require.NotNil(t, k)
	require.NotNil(t, ctx)
	require.NotNil(t, queryServer)

	// Test cases will be added here once the query types are defined
}
