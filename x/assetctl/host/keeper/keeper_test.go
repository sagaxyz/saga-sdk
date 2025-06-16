package keeper_test

import (
	"context"
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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/keeper"
	hosttypes "github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
	"github.com/stretchr/testify/require"
)

// MockACLKeeper is a mock implementation of ACLKeeper
type MockACLKeeper struct{}

func (m MockACLKeeper) Admin(ctx sdk.Context, addr sdk.AccAddress) bool {
	return true
}

// MockAccountKeeper is a mock implementation of AccountKeeper
type MockAccountKeeper struct{}

func (m MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress([]byte("cosmos1test"))
}

// MockICAControllerKeeper is a mock implementation of ICAControllerKeeper
type MockICAControllerKeeper struct{}

func (m MockICAControllerKeeper) GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool) {
	return "cosmos1test", true
}

func (m MockICAControllerKeeper) IsActiveChannel(ctx sdk.Context, connectionID, portID string) bool {
	return true
}

// MockERC20Keeper is a mock implementation of ERC20Keeper
type MockERC20Keeper struct{}

func (m MockERC20Keeper) RegisterERC20Extension(ctx sdk.Context, denom string) (*erc20types.TokenPair, error) {
	return nil, nil
}

// MockBankKeeper is a mock implementation of BankKeeper
type MockBankKeeper struct{}

func (m MockBankKeeper) GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool) {
	return banktypes.Metadata{}, true
}

func (m MockBankKeeper) HasDenomMetaData(ctx context.Context, denom string) bool {
	return true
}

func (m MockBankKeeper) SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata) {}

func setupTest(t *testing.T) (sdk.Context, *keeper.Keeper) {
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

	return ctx, k
}

func TestNewKeeper(t *testing.T) {
	ctx, k := setupTest(t)
	require.NotNil(t, k)
	require.NotNil(t, ctx)
}

func TestParams(t *testing.T) {
	ctx, k := setupTest(t)

	// Test setting params
	params := hosttypes.Params{
		HubConnectionId: "connection-0",
		HubChannelId:    "channel-0",
	}
	err := k.Params.Set(ctx, params)
	require.NoError(t, err)

	// Test getting params
	retrievedParams, err := k.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, params, retrievedParams)
}

func TestICAData(t *testing.T) {
	ctx, k := setupTest(t)

	// Test setting ICA data
	icaData := hosttypes.ICAOnHub{
		ChannelId: "channel-0",
		PortId:    "port-0",
	}
	err := k.ICAData.Set(ctx, icaData)
	require.NoError(t, err)

	// Test getting ICA data
	retrievedICAData, err := k.ICAData.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, icaData, retrievedICAData)

	// Test checking if ICA exists
	has, err := k.ICAData.Has(ctx)
	require.NoError(t, err)
	require.True(t, has)
}
