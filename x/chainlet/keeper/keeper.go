package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	upgradeKeeper    types.UpgradeKeeper
	channelKeeper    types.ChannelKeeper
	consumerKeeper   types.ConsumerKeeper
	clientKeeper     types.ClientKeeper
	connectionKeeper types.ConnectionKeeper
}

func New(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, upgradeKeeper types.UpgradeKeeper, channelKeeper types.ChannelKeeper, consumerKeeper types.ConsumerKeeper, clientKeeper types.ClientKeeper, connectionKeeper types.ConnectionKeeper, authority string) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		authority:        authority,
		upgradeKeeper:    upgradeKeeper,
		channelKeeper:    channelKeeper,
		consumerKeeper:   consumerKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ----------------------------------------------------------------------------
// IBC Keeper Logic
// ----------------------------------------------------------------------------

// GetPort returns the portID for the IBC app module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the IBC app module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}
