package transferrouter

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) []abci.ValidatorUpdate {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the current module state to a genesis file.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not get parameters at genesis"))
	}

	return &types.GenesisState{
		Params: params,
	}
}
