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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	mockAccountKeeper := MockAccountKeeper{}

	k := keeper.NewKeeper(storeService, cdc, logger, addressCodec)
	k.ICAHostKeeper = mockICAHostKeeper
	k.IBCChannelKeeper = mockIBCChannelKeeper
	k.IBCTransferKeeper = mockIBCTransferKeeper
	k.AccountKeeper = mockAccountKeeper
	k.Authority = "cosmos1test" // Set a test authority

	return ctx, k
}

func TestRegisterAssets(t *testing.T) {
	ctx, k := setupTest(t)
	moduleAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)

	tests := []struct {
		name    string
		msg     *types.MsgManageRegisteredAssets
		wantErr bool
		setup   func(*keeper.Keeper)
	}{
		{
			name: "valid registration",
			msg: &types.MsgManageRegisteredAssets{
				Authority: moduleAddress.String(),
				ChannelId: "channel-0",
				AssetsToRegister: []banktypes.Metadata{
					{
						Description: "Test Description",
						DenomUnits: []*banktypes.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
						Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
						Display: "test",
					},
				},
			},
			wantErr: false,
			setup: func(k *keeper.Keeper) {
				k.ICAHostKeeper = MockICAHostKeeper{
					Accounts: []icahosttypes.RegisteredInterchainAccount{
						{
							AccountAddress: moduleAddress.String(),
							ConnectionId:   "connection-0",
							PortId:         "icahost",
						},
					},
				}
				k.IBCChannelKeeper = MockIBCChannelKeeper{
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
			},
		},
		{
			name: "unauthorized",
			msg: &types.MsgManageRegisteredAssets{
				Authority: "wrong_authority",
				ChannelId: "channel-0",
				AssetsToRegister: []banktypes.Metadata{
					{
						Description: "Test Description",
						DenomUnits: []*banktypes.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
						Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
						Display: "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty assets",
			msg: &types.MsgManageRegisteredAssets{
				Authority:        moduleAddress.String(),
				ChannelId:        "channel-0",
				AssetsToRegister: []banktypes.Metadata{},
			},
			wantErr: true,
		},
		{
			name: "channel not found",
			msg: &types.MsgManageRegisteredAssets{
				Authority: moduleAddress.String(),
				ChannelId: "channel-0",
				AssetsToRegister: []banktypes.Metadata{
					{
						Description: "Test Description",
						DenomUnits: []*banktypes.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
						Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
						Display: "test",
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
			msg: &types.MsgManageRegisteredAssets{
				Authority: moduleAddress.String(),
				ChannelId: "channel-0",
				AssetsToRegister: []banktypes.Metadata{
					{
						Description: "Test Description",
						DenomUnits: []*banktypes.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
						Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
						Display: "test",
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
			msg: &types.MsgManageRegisteredAssets{
				Authority: moduleAddress.String(),
				ChannelId: "channel-0",
				AssetsToRegister: []banktypes.Metadata{
					{
						Description: "Test Description",
						DenomUnits: []*banktypes.DenomUnit{
							{
								Denom:    "test",
								Exponent: 6,
							},
						},
						Base:    "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
						Display: "test",
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
			msgServer := keeper.NewMsgServerImpl(*k)
			_, err := msgServer.ManageRegisteredAssets(ctx, tt.msg)
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
	moduleAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	tests := []struct {
		name    string
		msg     *types.MsgManageRegisteredAssets
		wantErr bool
		setup   func(*keeper.Keeper)
	}{
		{
			name: "valid unregistration",
			msg: &types.MsgManageRegisteredAssets{
				Authority:          moduleAddress.String(),
				ChannelId:          "channel-0",
				AssetsToUnregister: []string{"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"},
			},
			wantErr: false,
			setup: func(k *keeper.Keeper) {
				k.ICAHostKeeper = MockICAHostKeeper{
					Accounts: []icahosttypes.RegisteredInterchainAccount{
						{
							AccountAddress: moduleAddress.String(),
							ConnectionId:   "connection-0",
							PortId:         "icahost",
						},
					},
				}
				k.IBCChannelKeeper = MockIBCChannelKeeper{
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
			},
		},
		{
			name: "unauthorized",
			msg: &types.MsgManageRegisteredAssets{
				Authority:          "wrong_authority",
				ChannelId:          "channel-0",
				AssetsToUnregister: []string{"ibc/..."},
			},
			wantErr: true,
		},
		{
			name: "empty denoms",
			msg: &types.MsgManageRegisteredAssets{
				Authority:          moduleAddress.String(),
				ChannelId:          "channel-0",
				AssetsToUnregister: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(k)
			}
			msgServer := keeper.NewMsgServerImpl(*k)
			_, err := msgServer.ManageRegisteredAssets(ctx, tt.msg)
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
	moduleAddress := k.AccountKeeper.GetModuleAddress(assetctltypes.ModuleName)
	tests := []struct {
		name    string
		msg     *types.MsgManageSupportedAssets
		wantErr bool
		setup   func(*keeper.Keeper)
	}{
		{
			name: "valid support",
			msg: &types.MsgManageSupportedAssets{
				Authority:    moduleAddress.String(),
				ChannelId:    "channel-0",
				AddIbcDenoms: []string{"ibc/..."},
			},
			wantErr: false,
			setup: func(k *keeper.Keeper) {
				k.ICAHostKeeper = MockICAHostKeeper{
					Accounts: []icahosttypes.RegisteredInterchainAccount{
						{
							AccountAddress: moduleAddress.String(),
							ConnectionId:   "connection-0",
							PortId:         "icahost",
						},
					},
				}
				k.IBCChannelKeeper = MockIBCChannelKeeper{
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
			},
		},
		{
			name: "unauthorized",
			msg: &types.MsgManageSupportedAssets{
				Authority:    "wrong_authority",
				ChannelId:    "channel-0",
				AddIbcDenoms: []string{"ibc/..."},
			},
			wantErr: true,
		},
		{
			name: "empty denoms",
			msg: &types.MsgManageSupportedAssets{
				Authority:    moduleAddress.String(),
				ChannelId:    "channel-0",
				AddIbcDenoms: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(k)
			}
			msgServer := keeper.NewMsgServerImpl(*k)
			_, err := msgServer.ManageSupportedAssets(ctx, tt.msg)
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
