package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ParamsPrefix   = collections.NewPrefix(0) // Stores params
	ICAOnHubPrefix = collections.NewPrefix(1) // Stores the ICA on the hub
)

type ACLKeeper interface {
	Admin(ctx sdk.Context, addr sdk.AccAddress) bool
}

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeSvc     corestore.KVStoreService
	logger       log.Logger
	addressCodec address.Codec

	router        baseapp.MessageRouter
	aclKeeper     ACLKeeper
	accountKeeper AccountKeeper

	Authority string

	Schema  collections.Schema
	Params  collections.Item[types.Params]
	ICAData collections.Item[types.ICAOnHub]
}

func NewKeeper(storeSvc corestore.KVStoreService, cdc codec.BinaryCodec, logger log.Logger, addressCodec address.Codec) *Keeper {
	sb := collections.NewSchemaBuilder(storeSvc)

	k := &Keeper{
		storeSvc:     storeSvc,
		cdc:          cdc,
		logger:       logger,
		addressCodec: addressCodec,
		Params: collections.NewItem(sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		ICAData: collections.NewItem(sb,
			ICAOnHubPrefix,
			"ica",
			codec.CollValue[types.ICAOnHub](cdc),
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
