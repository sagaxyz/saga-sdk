package keeper

import (
	"github.com/sagaxyz/saga-sdk/x/assetctl/types"

	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	EnabledListPrefix   = collections.NewPrefix(0x00) // Stores ChainletIDs that have enabled the registry
	AssetMetadataPrefix = collections.NewPrefix(0x01) // Stores global asset metadata keyed by Hub IBC Denom
)

type IBCInterface interface {
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeSvc     corestore.KVStoreService
	logger       log.Logger
	addressCodec address.Codec

	Schema        collections.Schema
	EnabledList   collections.KeySet[string]                     // Key: ChainletID. Value: presence means enabled.
	AssetMetadata collections.Map[string, types.RegisteredAsset] // Key: Hub IBC Denom. Value: RegisteredAsset metadata.
}

func NewKeeper(storeSvc corestore.KVStoreService, cdc codec.BinaryCodec, logger log.Logger, addressCodec address.Codec) *Keeper {
	sb := collections.NewSchemaBuilder(storeSvc)

	k := &Keeper{
		storeSvc:     storeSvc,
		cdc:          cdc,
		logger:       logger,
		addressCodec: addressCodec,
		EnabledList: collections.NewKeySet(sb,
			EnabledListPrefix,
			"enabled_chainlets",    // Tracks chainlets that opted-in
			collections.StringKey), // Key is ChainletID
		AssetMetadata: collections.NewMap(sb,
			AssetMetadataPrefix,
			"asset_metadata",      // Global asset directory
			collections.StringKey, // Key is Hub IBC Denom
			codec.CollValue[types.RegisteredAsset](cdc)),
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
	// Populate AssetMetadata from genState.Assets
	for _, asset := range genState.Assets {
		// Assuming asset.IbcDenom is the Hub IBC denom for the asset
		if asset.IbcDenom == "" {
			panic(fmt.Errorf("genesis asset has empty ibc_denom: %+v", asset)) // Or handle more gracefully
		}
		k.AssetMetadata.Set(ctx, asset.IbcDenom, asset)
	}

	// TODO: If genState needs to store enabled chainlets, populate EnabledList here.
	// For example, if GenesisState has a field like `EnabledChainletIds []string`:
	// for _, chainletId := range genState.EnabledChainletIds {
	// 	k.EnabledList.Set(ctx, chainletId)
	// }
}

// ExportGenesis returns the keeper's exported genesis state.
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	assets := []types.RegisteredAsset{}
	iter, err := k.AssetMetadata.Iterate(ctx, nil)
	if err != nil {
		// Consider logging and returning an error, or a partially valid/empty genesis
		panic(fmt.Errorf("failed to iterate AssetMetadata: %w", err))
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		asset, err := iter.Value()
		if err != nil {
			// Consider logging and returning an error
			panic(fmt.Errorf("failed to get asset value from iterator: %w", err))
		}
		assets = append(assets, asset)
	}

	// TODO: If EnabledList needs to be part of genesis, export it here.
	// enabledChainletIds := []string{}
	// enabledIter, err := k.EnabledList.Iterate(ctx, nil)
	// if err != nil { panic(err) }
	// defer enabledIter.Close()
	// for ; enabledIter.Valid(); enabledIter.Next() {
	// 	 chainletId, err := enabledIter.Key()
	// 	 if err != nil { panic(err) }
	// 	 enabledChainletIds = append(enabledChainletIds, chainletId)
	// }

	// return types.NewGenesisState(assets, enabledChainletIds) // Assuming GenesisState can take both
	return &types.GenesisState{Assets: assets} // Placeholder if GenesisState only has Assets for now
}
