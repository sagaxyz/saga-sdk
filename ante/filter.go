package ante

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type FilterFn func(sdk.Context, sdk.AccAddress) bool

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
}

func BondedValidator(stakingKeeper StakingKeeper) FilterFn {
	return func(ctx sdk.Context, signer sdk.AccAddress) bool {
		valAddr := sdk.ValAddress(signer)

		var val stakingtypes.Validator
		val, err := stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return false
		}
		if val.GetStatus() != stakingtypes.Bonded {
			return false
		}

		return true
	}
}
