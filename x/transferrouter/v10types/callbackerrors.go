package v10types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrCannotUnmarshalPacketData = errorsmod.Register(ModuleName, 9, "cannot unmarshal packet data2")
	ErrNotPacketDataProvider     = errorsmod.Register(ModuleName, 10, "packet is not a PacketDataProvider2")
	ErrCallbackKeyNotFound       = errorsmod.Register(ModuleName, 11, "callback key not found in packet data2")
	ErrCallbackAddressNotFound   = errorsmod.Register(ModuleName, 12, "callback address not found in packet data2")
	ErrCallbackOutOfGas          = errorsmod.Register(ModuleName, 13, "callback out of gas2")
	ErrCallbackPanic             = errorsmod.Register(ModuleName, 14, "callback panic2")
	ErrInvalidCallbackData       = errorsmod.Register(ModuleName, 15, "invalid callback data2")
)
