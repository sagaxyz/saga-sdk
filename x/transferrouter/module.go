package transferrouter

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	// No CLI commands yet

	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	_ module.AppModule       = AppModule{}
	_ module.AppModuleBasic  = AppModuleBasic{}
	_ module.HasABCIEndBlock = AppModule{}
)

// ------------------------------
// AppModuleBasic
// ------------------------------

// AppModuleBasic implements the basic methods needed for a Cosmos SDK module.
type AppModuleBasic struct{}

// Name returns the module name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers legacy amino codec. Currently a no-op.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// ConsensusVersion specifies the consensus version of the module.
func (AppModuleBasic) ConsensusVersion() uint64 { return 1 }

// DefaultGenesis returns default genesis state for the module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis validation.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return gs.Validate()
}

// RegisterRESTRoutes does not expose REST endpoints.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes does not expose gRPC Gateway endpoints for now.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// GetTxCmd returns nil since the module does not support tx commands yet.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns nil since the module does not expose queries yet.
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// RegisterInterfaces registers protobuf interfaces.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// ------------------------------
// AppModule
// ------------------------------

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func (am AppModule) IsOnePerModuleType() {}
func (am AppModule) IsAppModule()        {}

// NewAppModule creates a new AppModule instance.
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// RegisterInvariants registers module invariants (none for now).
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterServices registers module gRPC services (none for now).
func (am AppModule) RegisterServices(_ module.Configurator) {}

// InitGenesis initializes module state from genesis data.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

// ExportGenesis exports current state as genesis.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// BeginBlock is a no-op for this module.
func (am AppModule) BeginBlock(_ sdk.Context) {}

// EndBlock returns no validator updates.
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	am.keeper.Logger(sdkCtx).Info("EndBlock called!!!! ================================")
	count := 0
	am.keeper.PacketQueue.Walk(ctx, nil, func(key uint64, value channeltypes.Packet) (stop bool, err error) {
		count++
		return false, nil

	})
	if count > 2 {
		am.keeper.Logger(sdkCtx).Info("Clearing packet queue", "count", count)
		am.keeper.PacketQueue.Clear(ctx, nil)
		return nil, nil
	}
	return nil, nil
}

// GenerateGenesisState is currently a no-op.
func (AppModule) GenerateGenesisState(_ *module.SimulationState) {}

// WeightedOperations returns nil as the module has no operations for simulation.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
