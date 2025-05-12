package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)

	for _, addr := range data.Admins {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		k.SetAdmin(ctx, accAddr)
	}
	for _, addr := range data.Allowed {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}
		k.SetAllowed(ctx, accAddr)
	}
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:  k.GetParams(ctx),
		Admins:  k.ExportAdmins(ctx),
		Allowed: k.ExportAllowed(ctx),
	}
}
