package keeper

import (
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	hosttypes "github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	"github.com/stretchr/testify/require"
)

type mockACLKeeper struct {
	adminAddr sdk.AccAddress
}

func (m mockACLKeeper) Admin(ctx sdk.Context, addr sdk.AccAddress) bool {
	return addr.Equals(m.adminAddr)
}

type mockAccountKeeper struct {
	moduleAddr sdk.AccAddress
}

func (m mockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return m.moduleAddr
}

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(key)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{
			"test": key,
		},
		map[string]*storetypes.TransientStoreKey{
			"transient_test": storetypes.NewTransientStoreKey("transient_test"),
		},
		nil,
	)

	addressCodec := addresscodec.NewBech32Codec("cosmos")
	logger := log.NewNopLogger()

	keeper := NewKeeper(storeService, cdc, logger, addressCodec)
	keeper.aclKeeper = mockACLKeeper{adminAddr: sdk.AccAddress([]byte("admin"))}
	keeper.accountKeeper = mockAccountKeeper{moduleAddr: sdk.AccAddress([]byte("module"))}
	keeper.Authority = "authority"
	keeper.router = baseapp.NewMsgServiceRouter()

	return keeper, ctx
}

func TestNewKeeper(t *testing.T) {
	keeper, _ := setupKeeper(t)
	require.NotNil(t, keeper)
	require.NotNil(t, keeper.Params)
	require.NotNil(t, keeper.ICAData)
}

func TestParams(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test setting params
	params := hosttypes.Params{
		HubConnectionId: "connection-0",
		HubChannelId:    "channel-0",
	}
	err := keeper.Params.Set(ctx, params)
	require.NoError(t, err)

	// Test getting params
	retrievedParams, err := keeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, params, retrievedParams)
}

func TestICAData(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Test setting ICA data
	icaData := hosttypes.ICAOnHub{
		ChannelId: "channel-0",
		PortId:    "port-0",
	}
	err := keeper.ICAData.Set(ctx, icaData)
	require.NoError(t, err)

	// Test getting ICA data
	retrievedICAData, err := keeper.ICAData.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, icaData, retrievedICAData)

	// Test checking if ICA exists
	has, err := keeper.ICAData.Has(ctx)
	require.NoError(t, err)
	require.True(t, has)
}
