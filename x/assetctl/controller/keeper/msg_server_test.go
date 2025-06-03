package keeper_test

import (
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
	"github.com/stretchr/testify/require"
)

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

	// Create mock keepers with realistic test data
	mockICAHostKeeper := MockICAHostKeeper{
		Accounts: []icahosttypes.RegisteredInterchainAccount{
			{
				AccountAddress: "cosmos1test",
				ConnectionId:   "connection-0",
				PortId:         "icahost",
			},
		},
	}
	mockIBCChannelKeeper := MockIBCChannelKeeper{
		Channel: channeltypes.Channel{
			State:    channeltypes.OPEN,
			Ordering: channeltypes.ORDERED,
			Counterparty: channeltypes.Counterparty{
				PortId:    "transfer",
				ChannelId: "channel-0",
			},
			ConnectionHops: []string{"connection-0"},
			Version:        "ics20-1",
		},
		Exists: true,
	}
	mockIBCTransferKeeper := MockIBCTransferKeeper{
		DenomTrace: ibctransfertypes.DenomTrace{
			Path:      "transfer/channel-0",
			BaseDenom: "test",
		},
		Exists: true,
	}

	k := keeper.NewKeeper(storeService, cdc, logger, addressCodec)
	k.ICAHostKeeper = mockICAHostKeeper
	k.IBCChannelKeeper = mockIBCChannelKeeper
	k.IBCTransferKeeper = mockIBCTransferKeeper
	k.Authority = "cosmos1test" // Set a test authority

	return ctx, k
}

func TestRegisterAssets(t *testing.T) {
	ctx, k := setupTest(t)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name    string
		msg     *types.MsgRegisterAssets
		wantErr bool
		setup   func(*keeper.Keeper)
	}{
		{
			name: "valid registration",
			msg: &types.MsgRegisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				AssetsToRegister: []types.AssetDetails{
					{
						IbcDenom:    "ibc/1234567890abcdef",
						Denom:       "test",
						DisplayName: "Test Asset",
						Description: "Test Description",
						DenomUnits: []types.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgRegisterAssets{
				Authority: "wrong_authority",
				ChannelId: "channel-0",
				AssetsToRegister: []types.AssetDetails{
					{
						IbcDenom: "ibc/...",
						Denom:    "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty assets",
			msg: &types.MsgRegisterAssets{
				Authority:        "cosmos1test",
				ChannelId:        "channel-0",
				AssetsToRegister: []types.AssetDetails{},
			},
			wantErr: true,
		},
		{
			name: "channel not found",
			msg: &types.MsgRegisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				AssetsToRegister: []types.AssetDetails{
					{
						IbcDenom: "ibc/...",
						Denom:    "test",
					},
				},
			},
			wantErr: true,
			setup: func(k *keeper.Keeper) {
				k.IBCChannelKeeper = MockIBCChannelKeeper{
					Channel: channeltypes.Channel{},
					Exists:  false,
				}
			},
		},
		{
			name: "denom trace not found",
			msg: &types.MsgRegisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				AssetsToRegister: []types.AssetDetails{
					{
						IbcDenom: "ibc/...",
						Denom:    "test",
					},
				},
			},
			wantErr: true,
			setup: func(k *keeper.Keeper) {
				k.IBCTransferKeeper = MockIBCTransferKeeper{
					DenomTrace: ibctransfertypes.DenomTrace{},
					Exists:     false,
				}
			},
		},
		{
			name: "interchain account not found",
			msg: &types.MsgRegisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				AssetsToRegister: []types.AssetDetails{
					{
						IbcDenom: "ibc/...",
						Denom:    "test",
					},
				},
			},
			wantErr: true,
			setup: func(k *keeper.Keeper) {
				k.ICAHostKeeper = MockICAHostKeeper{
					Accounts: []icahosttypes.RegisteredInterchainAccount{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(k)
			}
			_, err := msgServer.RegisterAssets(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUnregisterAssets(t *testing.T) {
	ctx, k := setupTest(t)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name    string
		msg     *types.MsgUnregisterAssets
		wantErr bool
	}{
		{
			name: "valid unregistration",
			msg: &types.MsgUnregisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				IbcDenoms: []string{"ibc/..."},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgUnregisterAssets{
				Authority: "wrong_authority",
				ChannelId: "channel-0",
				IbcDenoms: []string{"ibc/..."},
			},
			wantErr: true,
		},
		{
			name: "empty denoms",
			msg: &types.MsgUnregisterAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				IbcDenoms: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.UnregisterAssets(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSupportAsset(t *testing.T) {
	ctx, k := setupTest(t)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name    string
		msg     *types.MsgSupportAssets
		wantErr bool
	}{
		{
			name: "valid support",
			msg: &types.MsgSupportAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				IbcDenoms: []string{"ibc/..."},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgSupportAssets{
				Authority: "wrong_authority",
				ChannelId: "channel-0",
				IbcDenoms: []string{"ibc/..."},
			},
			wantErr: true,
		},
		{
			name: "empty denoms",
			msg: &types.MsgSupportAssets{
				Authority: "cosmos1test",
				ChannelId: "channel-0",
				IbcDenoms: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.SupportAssets(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateParams(t *testing.T) {
	ctx, k := setupTest(t)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name    string
		msg     *types.MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid update",
			msg: &types.MsgUpdateParams{
				Authority: "cosmos1test",
				Params:    &types.Params{},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgUpdateParams{
				Authority: "wrong_authority",
				Params:    &types.Params{},
			},
			wantErr: true,
		},
		{
			name: "nil params",
			msg: &types.MsgUpdateParams{
				Authority: "cosmos1test",
				Params:    nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.UpdateParams(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
