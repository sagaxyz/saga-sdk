package assetctl

import (
	"context"
	"encoding/json"
	"fmt"

	// "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	// "cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	"github.com/sagaxyz/saga-sdk/x/assetctl/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
	// _ module.HasConsensusVersion  = AppModule{}

	// _ appmodule.HasServices = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface that defines the
// independent methods a Cosmos SDK module needs to implement.
type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the name of the module as a string.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the amino codec for the module, which is used
// to marshal and unmarshal structs to/from []byte in order to persist them in the module's KVStore.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage.
// The default GenesisState need to be defined by the module developer and is primarily used for testing.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis used to validate the GenesisState, given in json.RawMessage.
// This method is called at server initialization for genesis validation.
// Ideally, it allows parallel execution with other modules genesis states.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this
// command are returned by default (ifcobra.CommandContexthor Viper !== nil).
func (ab AppModuleBasic) GetTxCmd() *cobra.Command {
	// return cli.GetTxCmd()
	return nil // Placeholder
}

// GetQueryCmd returns the root query command for the module. The subcommands of this
// command are returned by default (if cobra.CommandContexthor Viper !== nil).
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	// return cli.GetQueryCmd(types.StoreKey)
	return nil // Placeholder
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
	// accountKeeper types.AccountKeeper
	// bankKeeper    types.BankKeeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper /*accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper*/) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
		// accountKeeper:  accountKeeper,
		// bankKeeper:     bankKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))

	// m := keeper.NewMigrator(am.keeper)
	// if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
	// 	panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", types.ModuleName, err))
	// }
}

// InitGenesis performs the module's genesis initialization.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	am.keeper.InitGenesis(ctx, genState)
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock contains the logic that is automatically executed at the beginning of each block.
// The BeginBlock implementation logic varies between modules and some modules may not need to implement it.
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_ = sdkCtx
	return nil
}

// EndBlock contains the logic that is automatically executed at the end of each block.
// The EndBlock implementation logic varies between modules and some modules may not need to implement it.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	_ = sdkCtx
	return nil
}

// IsOnePerModuleType implements the OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	// simulation.RandomizedGenState(simState)
}

// // RegisterStoreDecoder registers a decoder for the module's types.
// func (am AppModule) RegisterStoreDecoder(sdr registry.StoreDecoderRegistry) {
// 	// sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
// }

// // WeightedOperations returns the all the module operations with their respective weights.
// func (am AppModule) WeightedOperations(simState module.SimulationState) []simulation.WeightedOperation {
// 	// return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.accountKeeper, am.bankKeeper, am.keeper)
// 	return nil
// }
