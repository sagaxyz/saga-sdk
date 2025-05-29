package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers the necessary x/assetctl/controller interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&GenesisState{}, "saga/assetctl/controller/GenesisState", nil)
	cdc.RegisterConcrete(&Params{}, "saga/assetctl/controller/Params", nil)
}

// RegisterInterfaces registers the x/assetctl/controller interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*GenesisState)(nil))
	registry.RegisterImplementations((*Params)(nil))
}
