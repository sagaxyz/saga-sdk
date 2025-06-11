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

func setupQueryServer(t *testing.T) (*Querier, sdk.Context) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(key)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{
			"test": key,
		},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)

	addressCodec := addresscodec.NewBech32Codec("cosmos")
	logger := log.NewNopLogger()

	keeper := NewKeeper(storeService, cdc, logger, addressCodec)
	keeper.aclKeeper = mockACLKeeper{adminAddr: sdk.AccAddress([]byte("admin"))}
	keeper.accountKeeper = mockAccountKeeper{moduleAddr: sdk.AccAddress([]byte("module"))}
	keeper.Authority = "authority"
	keeper.router = baseapp.NewMsgServiceRouter()

	return &Querier{Keeper: *keeper}, ctx
}

func TestQueryParams(t *testing.T) {
	querier, ctx := setupQueryServer(t)

	// Set params
	params := hosttypes.Params{
		HubConnectionId: "connection-0",
		HubChannelId:    "channel-0",
	}
	err := querier.Keeper.Params.Set(ctx, params)
	require.NoError(t, err)

	// Query params
	resp, err := querier.Params(ctx, &hosttypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params, resp.Params)
}

func TestQueryICAOnHub(t *testing.T) {
	querier, ctx := setupQueryServer(t)

	// Set ICA data
	icaData := hosttypes.ICAOnHub{
		ChannelId: "channel-0",
		PortId:    "port-0",
	}
	err := querier.Keeper.ICAData.Set(ctx, icaData)
	require.NoError(t, err)

	// Query ICA data
	resp, err := querier.ICAOnHub(ctx, &hosttypes.QueryICAOnHubRequest{})
	require.NoError(t, err)
	require.Equal(t, icaData, resp.IcaOnHub)

	// Test non-existent ICA
	err = querier.Keeper.ICAData.Remove(ctx)
	require.NoError(t, err)

	_, err = querier.ICAOnHub(ctx, &hosttypes.QueryICAOnHubRequest{})
	require.Error(t, err)
}
