package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const (
	// Amino names
	addAdminsName     = "saga/MsgAddAdmins"
	addAllowedName    = "saga/MsgAddAllowed"
	removeAdminsName  = "saga/MsgRemoveAdmins"
	removeAllowedName = "saga/MsgRemoveAllowed"
	enableName        = "saga/MsgEnable"
	disableName       = "saga/MsgDisable"
)

// RegisterInterfaces register implementations
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgAddAllowed{},
		&MsgAddAdmins{},
		&MsgRemoveAllowed{},
		&MsgRemoveAdmins{},
		&MsgEnable{},
		&MsgDisable{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/acl interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization and EIP-712 compatibility.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddAdmins{}, addAdminsName, nil)
	cdc.RegisterConcrete(&MsgAddAllowed{}, addAllowedName, nil)
	cdc.RegisterConcrete(&MsgRemoveAdmins{}, removeAdminsName, nil)
	cdc.RegisterConcrete(&MsgRemoveAllowed{}, removeAllowedName, nil)
	cdc.RegisterConcrete(&MsgEnable{}, enableName, nil)
	cdc.RegisterConcrete(&MsgDisable{}, disableName, nil)
}
