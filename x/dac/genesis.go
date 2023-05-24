package dac

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sagaxyz/sagaevm/v8/x/dac/keeper"
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetParams(ctx, data.Params)

	for _, admin := range data.Admins {
		addr := sdk.MustAccAddressFromBech32(admin)
		k.SetAdmin(ctx, addr)
	}
	for _, allowed := range data.Allowed {
		addr := common.HexToAddress(allowed)
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
