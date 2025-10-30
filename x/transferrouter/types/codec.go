package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterInterfaces is intentionally left blank for now as the module does not
// define any concrete Msg or interface types yet.
func RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

// RegisterLegacyAminoCodec is a no-op stub required by the AppModuleBasic.
func RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}
