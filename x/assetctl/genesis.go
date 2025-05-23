package assetctl

import (
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the asset
	// for _, elem := range genState.AssetList {
	// 	k.SetAsset(ctx, elem)
	// }

	// Set asset count
	// k.SetAssetCount(ctx, genState.AssetCount)
	// this line is used by starport scaffolding # genesis/module/init
	k.InitGenesis(ctx, genState) // Defer to keeper
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// genesis := types.DefaultGenesis() // This line is removed
	// genesis.Params = k.GetParams(ctx) // TODO: uncomment if params are used

	// genesis.AssetList = k.GetAllAsset(ctx)
	// genesis.AssetCount = k.GetAssetCount(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return k.ExportGenesis(ctx) // Defer to keeper
}
