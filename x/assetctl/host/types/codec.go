package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers the necessary x/assetctl/host interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&GenesisState{}, "saga/assetctl/host/GenesisState", nil)
	cdc.RegisterConcrete(&Params{}, "saga/assetctl/host/Params", nil)
}

// RegisterInterfaces registers the x/assetctl/host interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*GenesisState)(nil))
	registry.RegisterImplementations((*Params)(nil))
}
