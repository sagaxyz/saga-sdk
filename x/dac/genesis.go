package dac

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/dac/keeper"
	"github.com/sagaxyz/saga-sdk/x/dac/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetParams(ctx, data.Params)

	for _, addr := range data.Admins {
		accAddr, err := sdk.AccAddressFromBech32(addr.Value)
		if err != nil {
			panic(err)
		}
		k.SetAdmin(ctx, accAddr)
	}
	for _, addr := range data.Allowed {
		k.SetAllowed(ctx, addr)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:  k.GetParams(ctx),
		Admins:  k.ExportAdmins(ctx),
		Allowed: k.ExportAllowed(ctx),
	}
}
