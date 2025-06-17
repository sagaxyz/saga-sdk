package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	EnabledListPrefix     = collections.NewPrefix(0x00) // Stores ChainletIDs that have enabled the registry
	AssetMetadataPrefix   = collections.NewPrefix(0x01) // Stores global asset metadata keyed by Hub IBC Denom
	ParamsPrefix          = collections.NewPrefix(0x02) // Stores controller module parameters
	SupportedAssetsPrefix = collections.NewPrefix(0x03) // Stores supported assets keyed by ChainletID and Hub IBC Denom
)

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeSvc     corestore.KVStoreService
	logger       log.Logger
	addressCodec address.Codec

	ICAHostKeeper     ICAHostKeeper
	IBCChannelKeeper  IBCChannelKeeper
	IBCTransferKeeper IBCTransferKeeper
	AccountKeeper     AccountKeeper
	Authority         string

	Schema          collections.Schema
	AssetMetadata   collections.Map[string, types.RegisteredAsset]       // Key: Hub IBC Denom. Value: RegisteredAsset metadata.
	SupportedAssets collections.KeySet[collections.Pair[string, string]] // Key: ChannelID, IBC Denom (on the side of the hub chainlet)
	Params          collections.Item[types.Params]
}

// 1. ICA first message, will set it as the "admin" (double-check with Brian)
// 2. When the ICA sends a SupportAsset message, we will add the asset to the supported assets list, it's key will be the channel-id and the denom on the side of the hub chainlet.

func NewKeeper(storeSvc corestore.KVStoreService, cdc codec.BinaryCodec, logger log.Logger, addressCodec address.Codec) *Keeper {
	sb := collections.NewSchemaBuilder(storeSvc)

	k := &Keeper{
		storeSvc:     storeSvc,
		cdc:          cdc,
		logger:       logger,
		addressCodec: addressCodec,
		AssetMetadata: collections.NewMap(sb,
			AssetMetadataPrefix,
			"asset_metadata",      // Global asset directory
			collections.StringKey, // Key is Hub IBC Denom
			codec.CollValue[types.RegisteredAsset](cdc)),
		SupportedAssets: collections.NewKeySet(sb,
			SupportedAssetsPrefix,
			"supported_assets", // Tracks supported assets keyed by ChainletID and Hub IBC Denom
			collections.PairKeyCodec(collections.StringKey, collections.StringKey)),
		Params: collections.NewItem(sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc)),
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
	// TODO: figure out if we need to do anything here
}

// ExportGenesis returns the keeper's exported genesis state.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{}
}
