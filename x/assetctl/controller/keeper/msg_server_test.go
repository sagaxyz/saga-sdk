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
	"github.com/sagaxyz/saga-sdk/x/assetctl"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
	assetctltypes "github.com/sagaxyz/saga-sdk/x/assetctl/types"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (sdk.Context, types.MsgServer) {
	keys := storetypes.NewKVStoreKeys(
		assetctltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(assetctl.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	storeService := runtime.NewKVStoreService(storetypes.NewKVStoreKey("test"))

	ctx := sdk.NewContext(cms, tmproto.Header{}, true, logger)

	var addressCodec address.Codec = nil // Use nil or a mock if not available

	k := keeper.NewKeeper(storeService, cdc, logger, addressCodec)
	k.Authority = "cosmos1test" // Set a test authority
	msgServer := keeper.NewMsgServerImpl(*k)

	return ctx, msgServer
}

func TestRegisterAssets(t *testing.T) {
	ctx, msgServer := setupTest(t)

	tests := []struct {
		name    string
		msg     *types.MsgRegisterAssets
		wantErr bool
	}{
		{
			name: "valid registration",
			msg: &types.MsgRegisterAssets{
				Authority: "cosmos1test",
				AssetsToRegister: []types.AssetDetails{
					{
						Denom:       "ibc/...",
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
				AssetsToRegister: []types.AssetDetails{
					{
						Denom: "ibc/...",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty assets",
			msg: &types.MsgRegisterAssets{
				Authority:        "cosmos1test",
				AssetsToRegister: []types.AssetDetails{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	ctx, msgServer := setupTest(t)

	tests := []struct {
		name    string
		msg     *types.MsgUnregisterAssets
		wantErr bool
	}{
		{
			name: "valid unregistration",
			msg: &types.MsgUnregisterAssets{
				Authority:             "cosmos1test",
				IbcDenomsToUnregister: []string{"ibc/..."},
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgUnregisterAssets{
				Authority:             "wrong_authority",
				IbcDenomsToUnregister: []string{"ibc/..."},
			},
			wantErr: true,
		},
		{
			name: "empty denoms",
			msg: &types.MsgUnregisterAssets{
				Authority:             "cosmos1test",
				IbcDenomsToUnregister: []string{},
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

func TestToggleChainletRegistry(t *testing.T) {
	ctx, msgServer := setupTest(t)

	tests := []struct {
		name    string
		msg     *types.MsgToggleChainletRegistry
		wantErr bool
	}{
		{
			name: "valid toggle",
			msg: &types.MsgToggleChainletRegistry{
				Authority:  "cosmos1test",
				ChainletId: "chain-1",
				Enable:     true,
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgToggleChainletRegistry{
				Authority:  "wrong_authority",
				ChainletId: "chain-1",
				Enable:     true,
			},
			wantErr: true,
		},
		{
			name: "empty chainlet id",
			msg: &types.MsgToggleChainletRegistry{
				Authority:  "cosmos1test",
				ChainletId: "",
				Enable:     true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.ToggleChainletRegistry(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSupportAsset(t *testing.T) {
	ctx, msgServer := setupTest(t)

	tests := []struct {
		name    string
		msg     *types.MsgSupportAsset
		wantErr bool
	}{
		{
			name: "valid support",
			msg: &types.MsgSupportAsset{
				Authority: "cosmos1test",
				IbcDenom:  "ibc/...",
			},
			wantErr: false,
		},
		{
			name: "unauthorized",
			msg: &types.MsgSupportAsset{
				Authority: "wrong_authority",
				IbcDenom:  "ibc/...",
			},
			wantErr: true,
		},
		{
			name: "empty denom",
			msg: &types.MsgSupportAsset{
				Authority: "cosmos1test",
				IbcDenom:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.SupportAsset(ctx, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateParams(t *testing.T) {
	ctx, msgServer := setupTest(t)

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
