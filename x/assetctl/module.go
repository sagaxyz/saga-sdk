package assetctl

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	controllercli "github.com/sagaxyz/saga-sdk/x/assetctl/controller/client/cli"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	controllertypes "github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
	hostcli "github.com/sagaxyz/saga-sdk/x/assetctl/host/client/cli"
	hostkeeper "github.com/sagaxyz/saga-sdk/x/assetctl/host/keeper"
	hosttypes "github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
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
	return "assetctl"
}

// RegisterLegacyAminoCodec registers the amino codec for the module, which is used
// to marshal and unmarshal structs to/from []byte in order to persist them in the module's KVStore.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register host types
	hosttypes.RegisterLegacyAminoCodec(cdc)
	// Register controller types
	controllertypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register host types
	hosttypes.RegisterInterfaces(registry)
	// Register controller types
	controllertypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage.
// The default GenesisState need to be defined by the module developer and is primarily used for testing.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Create default genesis states for both host and controller
	hostGenState := hosttypes.GenesisState{
		Params: hosttypes.Params{},
	}
	controllerGenState := controllertypes.GenesisState{
		Params: controllertypes.Params{},
	}

	// Combine both genesis states
	combinedGenState := struct {
		Host       hosttypes.GenesisState       `json:"host"`
		Controller controllertypes.GenesisState `json:"controller"`
	}{
		Host:       hostGenState,
		Controller: controllerGenState,
	}

	// Marshal the combined state to JSON
	bz, err := json.Marshal(&combinedGenState)
	if err != nil {
		panic(fmt.Errorf("failed to marshal default genesis state: %w", err))
	}
	return bz
}

// ValidateGenesis used to validate the GenesisState, given in json.RawMessage.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState struct {
		Host       hosttypes.GenesisState       `json:"host"`
		Controller controllertypes.GenesisState `json:"controller"`
	}

	if err := json.Unmarshal(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal assetctl genesis state: %w", err)
	}

	// Validate host genesis state
	if err := genState.Host.Validate(); err != nil {
		return fmt.Errorf("invalid host genesis state: %w", err)
	}

	// Validate controller genesis state
	if err := genState.Controller.Validate(); err != nil {
		return fmt.Errorf("invalid controller genesis state: %w", err)
	}

	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	hosttypes.RegisterQueryHandlerClient(context.Background(), mux, hosttypes.NewQueryClient(clientCtx))
	controllertypes.RegisterQueryHandlerClient(context.Background(), mux, controllertypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this
// command are returned by default (if cobra.CommandContexthor Viper !== nil).
func (ab AppModuleBasic) GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assetctl",
		Short:                      "Asset control transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		controllercli.GetTxCmd(),
		hostcli.GetTxCmd(),
	)

	return cmd
}

// GetQueryCmd returns the root query command for the module. The subcommands of this
// command are returned by default (if cobra.CommandContexthor Viper !== nil).
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assetctl",
		Short:                      "Asset control query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		controllercli.GetQueryCmd(),
		hostcli.GetQueryCmd(),
	)

	return cmd
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic

	keeper     keeper.Keeper
	hostKeeper hostkeeper.Keeper
	// accountKeeper types.AccountKeeper
	// bankKeeper    types.BankKeeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, hostKeeper hostkeeper.Keeper /*accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper*/) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
		hostKeeper:     hostKeeper,
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
	hosttypes.RegisterMsgServer(cfg.MsgServer(), hostkeeper.NewMsgServerImpl(am.hostKeeper))
	hosttypes.RegisterQueryServer(cfg.QueryServer(), hostkeeper.NewQueryServerImpl(am.hostKeeper))

	controllertypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	controllertypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis performs the module's genesis initialization.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState struct {
		Host       hosttypes.GenesisState       `json:"host"`
		Controller controllertypes.GenesisState `json:"controller"`
	}

	if err := json.Unmarshal(gs, &genState); err != nil {
		panic(fmt.Errorf("failed to unmarshal assetctl genesis state: %w", err))
	}

	// Initialize host genesis state
	am.hostKeeper.InitGenesis(ctx, genState.Host)

	// Initialize controller genesis state
	am.keeper.InitGenesis(ctx, genState.Controller)
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	hostGenState := am.hostKeeper.ExportGenesis(ctx)
	controllerGenState := am.keeper.ExportGenesis(ctx)

	combinedGenState := struct {
		Host       hosttypes.GenesisState       `json:"host"`
		Controller controllertypes.GenesisState `json:"controller"`
	}{
		Host:       *hostGenState,
		Controller: *controllerGenState,
	}

	// Marshal the combined state to JSON
	bz, err := json.Marshal(&combinedGenState)
	if err != nil {
		panic(fmt.Errorf("failed to marshal genesis state: %w", err))
	}
	return bz
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock contains the logic that is automatically executed at the beginning of each block.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

// EndBlock contains the logic that is automatically executed at the end of each block.
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}

// IsOnePerModuleType implements the OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
}
