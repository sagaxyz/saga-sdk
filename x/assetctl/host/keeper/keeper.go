package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	EnabledListPrefix = collections.NewPrefix(0x00) // Stores ChainletIDs that have enabled the registry
	ParamsPrefix      = collections.NewPrefix(0x01) // Stores global asset metadata keyed by Hub IBC Denom
)

type IBCInterface interface {
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeSvc     corestore.KVStoreService
	logger       log.Logger
	addressCodec address.Codec

	Authority string

	Schema      collections.Schema
	EnabledList collections.KeySet[string] // Key: ChainletID. Value: presence means enabled.
	Params      collections.Item[types.Params]
}

func NewKeeper(storeSvc corestore.KVStoreService, cdc codec.BinaryCodec, logger log.Logger, addressCodec address.Codec) *Keeper {
	sb := collections.NewSchemaBuilder(storeSvc)

	k := &Keeper{
		storeSvc:     storeSvc,
		cdc:          cdc,
		logger:       logger,
		addressCodec: addressCodec,
		EnabledList: collections.NewKeySet(
			sb,
			EnabledListPrefix,
			"enabled_chainlets", // Tracks chainlets that opted-in
			collections.StringKey,
		),
		Params: collections.NewItem(sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc),
		),
	}

	var err error
	k.Schema, err = sb.Build()
	if err != nil {
		panic(err)
	}

	return k
}

// InitGenesis initializes the keeper's state from a provided genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// TODO: Implement
}

// ExportGenesis returns the keeper's exported genesis state.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// TODO: Implement
	return &types.GenesisState{}
}
