package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	// Amino names
	setMetadataName    = "saga/MsgSetMetadata"
	EnableSetMetadata  = "saga/MsgEnableSetMetadata"
	DisableSetMetadata = "saga/MsgDisableSetMetadata"
)

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSetMetadata{},
		&MsgEnableSetMetadata{},
		&MsgDisableSetMetadata{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/admin interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetMetadata{}, setMetadataName, nil)
	cdc.RegisterConcrete(&MsgEnableSetMetadata{}, EnableSetMetadata, nil)
	cdc.RegisterConcrete(&MsgDisableSetMetadata{}, DisableSetMetadata, nil)
}
