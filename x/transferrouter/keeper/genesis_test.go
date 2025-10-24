package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
	"github.com/stretchr/testify/require"
)

// buildKeeper composes a minimal keeper instance backed by an in-memory KVStoreService and no external deps used by Init/Export.
func buildKeeper(t *testing.T) (sdk.Context, keeper.Keeper) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{
			types.StoreKey: key,
		},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})

	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	// Minimal keeper with only store service and codec required for params collections.
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, // erc20 keeper
		nil, // ics4 wrapper
		nil, // channel keeper
		nil, // transfer keeper
		nil, // bank keeper
		nil, // account keeper
		nil, // evm keeper
		"",
	)

	return ctx, k
}

func TestInitExportGenesis(t *testing.T) {
	ctx, k := buildKeeper(t)

	// default genesis
	gs := types.GenesisState{Params: types.DefaultGenesisState().Params}

	// module InitGenesis/ExportGenesis
	_ = transferrouter.InitGenesis(ctx, k, gs)
	exported := transferrouter.ExportGenesis(ctx, k)

	require.Equal(t, gs.Params, exported.Params)

	// mutate and re-init
	gs.Params.Enabled = false
	_ = transferrouter.InitGenesis(ctx, k, gs)
	exported2 := transferrouter.ExportGenesis(ctx, k)
	require.Equal(t, gs.Params, exported2.Params)
}
