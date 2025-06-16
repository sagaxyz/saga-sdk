package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
	"github.com/stretchr/testify/require"
)

// Ensure mockRouter is defined at the top level
type mockRouter struct{}

func setupMsgServer(t *testing.T) (sdk.Context, *keeper.Keeper, types.MsgServer) {
	keys := storetypes.NewKVStoreKeys(
		assetctltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(assetctl.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	storeService := runtime.NewKVStoreService(keys[assetctltypes.StoreKey])

	ctx := sdk.NewContext(cms, tmproto.Header{}, true, logger)

	// Use a real bech32 address codec for testing
	addressCodec := addresscodec.NewBech32Codec("cosmos")

	// Create mock keepers
	mockACLKeeper := MockACLKeeper{}
	mockAccountKeeper := MockAccountKeeper{}
	mockICAControllerKeeper := MockICAControllerKeeper{}
	mockERC20Keeper := MockERC20Keeper{}
	mockBankKeeper := MockBankKeeper{}

	// mockRouter implements baseapp.MessageRouter
	// It returns a dummy response for MsgSendTx

	msgRouter := mockRouter{}

	k := keeper.NewKeeper(
		storeService,
		cdc,
		logger,
		addressCodec,
		msgRouter,
		mockACLKeeper,
		mockAccountKeeper,
		mockICAControllerKeeper,
		mockERC20Keeper,
		mockBankKeeper,
		"cosmos1test",
	)

	// Set default Params and ICAData for tests
	params := types.Params{
		HubConnectionId: "connection-0",
		HubChannelId:    "channel-0",
	}
	err := k.Params.Set(ctx, params)
	if err != nil {
		t.Fatalf("failed to set params: %v", err)
	}
	ica := types.ICAOnHub{
		ChannelId: "channel-0",
		PortId:    "icahost",
	}
	err = k.ICAData.Set(ctx, ica)
	if err != nil {
		t.Fatalf("failed to set ICAData: %v", err)
	}

	msgServer := keeper.NewMsgServerImpl(*k)

	return ctx, k, msgServer
}

func TestMsgServer(t *testing.T) {
	ctx, k, msgServer := setupMsgServer(t)
	require.NotNil(t, k)
	require.NotNil(t, ctx)
	require.NotNil(t, msgServer)

	moduleAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)

	// Test ManageSupportedAssets
	t.Run("ManageSupportedAssets", func(t *testing.T) {
		// Test unauthorized
		msg := &types.MsgManageSupportedAssets{
			Authority:       "invalid",
			AddIbcDenoms:    []string{"ibc/denom1"},
			RemoveIbcDenoms: []string{},
		}
		_, err := msgServer.ManageSupportedAssets(ctx, msg)
		require.Error(t, err)

		// Test valid message
		msg = &types.MsgManageSupportedAssets{
			Authority:       moduleAddress.String(),
			AddIbcDenoms:    []string{"ibc/denom1", "ibc/denom2"},
			RemoveIbcDenoms: []string{},
		}
		resp, err := msgServer.ManageSupportedAssets(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	// Test ManageRegisteredAssets
	t.Run("ManageRegisteredAssets", func(t *testing.T) {
		// Test unauthorized
		msg := &types.MsgManageRegisteredAssets{
			Authority:          "wrong-authority",
			AssetsToRegister:   []string{"ibc/denom1"},
			AssetsToUnregister: []string{},
		}
		_, err := msgServer.ManageRegisteredAssets(ctx, msg)
		require.Error(t, err)

		// Test valid message
		msg = &types.MsgManageRegisteredAssets{
			Authority:          moduleAddress.String(),
			AssetsToRegister:   []string{"ibc/denom1", "ibc/denom2"},
			AssetsToUnregister: []string{},
		}
		resp, err := msgServer.ManageRegisteredAssets(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	// Test CreateICAOnHub
	t.Run("CreateICAOnHub", func(t *testing.T) {
		// reset the ICAData
		err := k.ICAData.Remove(ctx)
		require.NoError(t, err)

		// Test unauthorized
		msg := &types.MsgCreateICAOnHub{
			Authority: "invalid",
		}
		_, err = msgServer.CreateICAOnHub(ctx, msg)
		require.Error(t, err)

		// Test valid message
		msg = &types.MsgCreateICAOnHub{
			Authority: moduleAddress.String(),
		}
		resp, err := msgServer.CreateICAOnHub(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	// Test UpdateParams
	t.Run("UpdateParams", func(t *testing.T) {
		// Test unauthorized
		msg := &types.MsgUpdateParams{
			Authority: "invalid",
			Params: &types.Params{
				HubConnectionId: "connection-0",
				HubChannelId:    "channel-0",
			},
		}
		_, err := msgServer.UpdateParams(ctx, msg)
		require.Error(t, err)

		// Test valid message
		msg = &types.MsgUpdateParams{
			Authority: moduleAddress.String(),
			Params: &types.Params{
				HubConnectionId: "connection-0",
				HubChannelId:    "channel-0",
			},
		}
		resp, err := msgServer.UpdateParams(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func (m mockRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case *icacontrollertypes.MsgSendTx:
			any, err := codectypes.NewAnyWithValue(&icacontrollertypes.MsgSendTxResponse{Sequence: 1})
			if err != nil {
				return nil, err
			}
			return &sdk.Result{MsgResponses: []*codectypes.Any{any}}, nil
		case *icacontrollertypes.MsgRegisterInterchainAccount:
			any, err := codectypes.NewAnyWithValue(&icacontrollertypes.MsgRegisterInterchainAccountResponse{
				ChannelId: "channel-0",
				PortId:    "port-0",
			})
			if err != nil {
				return nil, err
			}
			return &sdk.Result{MsgResponses: []*codectypes.Any{any}}, nil
		case *types.MsgManageSupportedAssets:
			any, err := codectypes.NewAnyWithValue(&types.MsgManageSupportedAssetsResponse{Sequence: 1})
			if err != nil {
				return nil, err
			}
			return &sdk.Result{MsgResponses: []*codectypes.Any{any}}, nil
		default:
			return nil, fmt.Errorf("unhandled message type: %T", msg)
		}
	}
}

func (m mockRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	return m.Handler(nil)
}
