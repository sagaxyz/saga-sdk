package keeper

import (
	"context"
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
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	hosttypes "github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
	"github.com/stretchr/testify/require"
)

var _ icacontrollertypes.MsgServer = &mockICAControllerMsgServer{}

type mockICAControllerMsgServer struct {
}

func (m *mockICAControllerMsgServer) RegisterInterchainAccount(ctx context.Context, msg *icacontrollertypes.MsgRegisterInterchainAccount) (*icacontrollertypes.MsgRegisterInterchainAccountResponse, error) {
	return &icacontrollertypes.MsgRegisterInterchainAccountResponse{
		ChannelId: "channel-0",
		PortId:    "icacontroller-0",
	}, nil
}

func (m *mockICAControllerMsgServer) SendTx(ctx context.Context, msg *icacontrollertypes.MsgSendTx) (*icacontrollertypes.MsgSendTxResponse, error) {
	return &icacontrollertypes.MsgSendTxResponse{
		Sequence: 1,
	}, nil
}

func (m *mockICAControllerMsgServer) UpdateParams(ctx context.Context, msg *icacontrollertypes.MsgUpdateParams) (*icacontrollertypes.MsgUpdateParamsResponse, error) {
	return &icacontrollertypes.MsgUpdateParamsResponse{}, nil
}

func setupMsgServer(t *testing.T) (*msgServer, sdk.Context) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	icacontrollertypes.RegisterInterfaces(interfaceRegistry)
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
	adminAddr := sdk.AccAddress([]byte("admin123456789012345678901234567890"))
	moduleAddr := sdk.AccAddress([]byte("module123456789012345678901234567890"))
	keeper.aclKeeper = mockACLKeeper{adminAddr: adminAddr}
	keeper.accountKeeper = mockAccountKeeper{moduleAddr: moduleAddr}
	keeper.Authority = adminAddr.String()

	// Register ICA controller interfaces and msg server in the router
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(interfaceRegistry)
	icacontrollertypes.RegisterMsgServer(router, &mockICAControllerMsgServer{})

	keeper.router = router

	// Initialize params
	params := hosttypes.Params{
		HubConnectionId: "connection-0",
		HubChannelId:    "channel-0",
	}
	err := keeper.Params.Set(ctx, params)
	require.NoError(t, err)

	return &msgServer{Keeper: *keeper}, ctx
}

func TestRegisterDenoms(t *testing.T) {
	msgServer, ctx := setupMsgServer(t)

	// Test unauthorized
	invalidAddr := sdk.AccAddress([]byte("invalid123456789012345678901234567890"))
	msg := &hosttypes.MsgRegisterDenoms{
		Authority: invalidAddr.String(),
		IbcDenoms: []string{"ibc/denom1"},
	}
	_, err := msgServer.RegisterDenoms(ctx, msg)
	require.Error(t, err)

	// Test empty denoms
	adminAddr := sdk.AccAddress([]byte("admin123456789012345678901234567890"))
	msg = &hosttypes.MsgRegisterDenoms{
		Authority: adminAddr.String(),
		IbcDenoms: []string{},
	}
	_, err = msgServer.RegisterDenoms(ctx, msg)
	require.Error(t, err)

	// Test valid message
	msg = &hosttypes.MsgRegisterDenoms{
		Authority: adminAddr.String(),
		IbcDenoms: []string{"ibc/denom1", "ibc/denom2"},
	}
	_, err = msgServer.RegisterDenoms(ctx, msg)
	require.NoError(t, err)
}

func TestUpdateParams(t *testing.T) {
	msgServer, ctx := setupMsgServer(t)

	// Test unauthorized
	invalidAddr := sdk.AccAddress([]byte("invalid123456789012345678901234567890"))
	msg := &hosttypes.MsgUpdateParams{
		Authority: invalidAddr.String(),
		Params: &hosttypes.Params{
			HubConnectionId: "connection-0",
			HubChannelId:    "channel-0",
		},
	}
	_, err := msgServer.UpdateParams(ctx, msg)
	require.Error(t, err)

	// Test nil params
	adminAddr := sdk.AccAddress([]byte("admin123456789012345678901234567890"))
	msg = &hosttypes.MsgUpdateParams{
		Authority: adminAddr.String(),
		Params:    nil,
	}
	_, err = msgServer.UpdateParams(ctx, msg)
	require.Error(t, err)

	// Test valid message
	msg = &hosttypes.MsgUpdateParams{
		Authority: adminAddr.String(),
		Params: &hosttypes.Params{
			HubConnectionId: "connection-0",
			HubChannelId:    "channel-0",
		},
	}
	_, err = msgServer.UpdateParams(ctx, msg)
	require.NoError(t, err)
}

func TestSupportAssets(t *testing.T) {
	msgServer, ctx := setupMsgServer(t)

	// Test unauthorized
	invalidAddr := sdk.AccAddress([]byte("invalid123456789012345678901234567890"))
	msg := &hosttypes.MsgSupportAssets{
		Authority: invalidAddr.String(),
		IbcDenoms: []string{"ibc/denom1"},
	}
	_, err := msgServer.SupportAssets(ctx, msg)
	require.Error(t, err)

	// Test empty denoms
	adminAddr := sdk.AccAddress([]byte("admin123456789012345678901234567890"))
	msg = &hosttypes.MsgSupportAssets{
		Authority: adminAddr.String(),
		IbcDenoms: []string{},
	}
	_, err = msgServer.SupportAssets(ctx, msg)
	require.Error(t, err)

	// Test valid message
	msg = &hosttypes.MsgSupportAssets{
		Authority: adminAddr.String(),
		IbcDenoms: []string{"ibc/denom1", "ibc/denom2"},
	}
	_, err = msgServer.SupportAssets(ctx, msg)
	require.NoError(t, err)
}

func TestCreateICAOnHub(t *testing.T) {
	msgServer, ctx := setupMsgServer(t)

	// Test unauthorized
	invalidAddr := sdk.AccAddress([]byte("invalid123456789012345678901234567890"))
	msg := &hosttypes.MsgCreateICAOnHub{
		Authority: invalidAddr.String(),
	}
	_, err := msgServer.CreateICAOnHub(ctx, msg)
	require.Error(t, err)

	// Test valid message
	adminAddr := sdk.AccAddress([]byte("admin123456789012345678901234567890"))
	msg = &hosttypes.MsgCreateICAOnHub{
		Authority: adminAddr.String(),
	}
	_, err = msgServer.CreateICAOnHub(ctx, msg)
	require.NoError(t, err)

	// Test duplicate creation
	_, err = msgServer.CreateICAOnHub(ctx, msg)
	require.Error(t, err)
}
