package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermintevmtypes "github.com/evmos/ethermint/x/evm/types"
	ethermintfeemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

var _ ethermintevmtypes.FeeMarketKeeper = Keeper{}

func (k Keeper) AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error) {
	return k.feeMarketKeeper.AddTransientGasWanted(ctx, gasWanted)
}

func (k Keeper) GetParams(ctx sdk.Context) ethermintfeemarkettypes.Params {
	return k.feeMarketKeeper.GetParams(ctx)
}

func (k Keeper) GetBaseFee(ctx sdk.Context) *big.Int {
	params := k.GetParams2(ctx)
	if params.Override {
		return params.BaseFee.BigInt()
	}

	return k.feeMarketKeeper.GetBaseFee(ctx)
}
