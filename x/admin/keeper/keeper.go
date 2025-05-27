package keeper

import (
	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/sagaxyz/saga-sdk/x/admin/types"
)

type Keeper struct {
	cdc        codec.Codec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace
	bankKeeper types.BankKeeper
	aclKeeper  types.AclKeeper
	authority  string
}

func New(cdc codec.Codec, storeKey storetypes.StoreKey, ps paramtypes.Subspace, bk types.BankKeeper, aclk types.AclKeeper, authority string) Keeper {
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: ps,
		bankKeeper: bk,
		aclKeeper:  aclk,
		authority:  authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}
func (k *Keeper) GetAuthority() string {
	return k.authority
}
