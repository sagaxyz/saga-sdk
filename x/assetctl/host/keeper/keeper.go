package keeper

import (
	"context"

	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

var (
	ParamsPrefix           = collections.NewPrefix(0) // Stores params
	ICAOnHubPrefix         = collections.NewPrefix(1) // Stores the ICA on the hub
	InFlightRequestsPrefix = collections.NewPrefix(2) // Stores the in-flight requests
)

type ACLKeeper interface {
	Admin(ctx sdk.Context, addr sdk.AccAddress) bool
}

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type ICAControllerKeeper interface {
	GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
	IsActiveChannel(ctx sdk.Context, connectionID, portID string) bool
}

type ERC20Keeper interface {
	RegisterERC20Extension(ctx sdk.Context, denom string) (*erc20types.TokenPair, error)
}

type BankKeeper interface {
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	HasDenomMetaData(ctx context.Context, denom string) bool
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeSvc     corestore.KVStoreService
	logger       log.Logger
	addressCodec address.Codec

	router              baseapp.MessageRouter
	aclKeeper           ACLKeeper
	accountKeeper       AccountKeeper
	icaControllerKeeper ICAControllerKeeper
	Erc20Keeper         ERC20Keeper
	BankKeeper          BankKeeper

	Authority string

	Schema           collections.Schema
	Params           collections.Item[types.Params]
	ICAData          collections.Item[types.ICAOnHub]
	InFlightRequests collections.Map[uint64, string] // sequence -> message type url
}

func NewKeeper(
	storeSvc corestore.KVStoreService,
	cdc codec.BinaryCodec,
	logger log.Logger,
	addressCodec address.Codec,
	router baseapp.MessageRouter,
	aclKeeper ACLKeeper,
	accountKeeper AccountKeeper,
	icaControllerKeeper ICAControllerKeeper,
	erc20Keeper ERC20Keeper,
	bankKeeper BankKeeper,
	authority string,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeSvc)

	k := &Keeper{
		storeSvc:            storeSvc,
		cdc:                 cdc,
		logger:              logger,
		addressCodec:        addressCodec,
		router:              router,
		aclKeeper:           aclKeeper,
		accountKeeper:       accountKeeper,
		icaControllerKeeper: icaControllerKeeper,
		Erc20Keeper:         erc20Keeper,
		BankKeeper:          bankKeeper,
		Authority:           authority,
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
		InFlightRequests: collections.NewMap(sb,
			InFlightRequestsPrefix,
			"in_flight_requests",
			collections.Uint64Key,
			collections.StringValue,
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
