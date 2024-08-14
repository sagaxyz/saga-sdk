package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgAddAllowed{}
	_ sdk.Msg = &MsgRemoveAllowed{}
	_ sdk.Msg = &MsgAddAdmins{}
	_ sdk.Msg = &MsgRemoveAdmins{}
	_ sdk.Msg = &MsgEnable{}
	_ sdk.Msg = &MsgDisable{}
)

const (
	TypeMsgAddAllowed    = "add_allowed"
	TypeMsgRemoveAllowed = "remove_allowed"
	TypeMsgAddAdmins     = "add_admins"
	TypeMsgRemoveAdmins  = "remove_admins"
	TypeMsgEnable        = "enable"
	TypeMsgDisable       = "disable"
)

// NewMsgAddAllowed creates a new instance of MsgAddAllowed
func NewMsgAddAllowed(sender string, allowed ...string) *MsgAddAllowed { // nolint: interfacer
	return &MsgAddAllowed{
		Sender:  sender,
		Allowed: allowed,
	}
}

// Route should return the name of the module
func (msg MsgAddAllowed) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAddAllowed) Type() string { return TypeMsgAddAllowed }

// ValidateBasic runs stateless checks on the message
func (msg MsgAddAllowed) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	for _, addr := range msg.Allowed {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid allowed address %s", addr)
		}
	}
	return nil
}

// NewMsgRemoveAllowed creates a new instance of MsgRemoveAllowed
func NewMsgRemoveAllowed(sender string, allowed ...string) *MsgRemoveAllowed { // nolint: interfacer
	return &MsgRemoveAllowed{
		Sender:  sender,
		Allowed: allowed,
	}
}

// Route should return the name of the module
func (msg MsgRemoveAllowed) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRemoveAllowed) Type() string { return TypeMsgRemoveAllowed }

// ValidateBasic runs stateless checks on the message
func (msg MsgRemoveAllowed) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	for _, addr := range msg.Allowed {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid allowed address %s", addr)
		}
	}
	return nil
}

// NewMsgAddAdmins creates a new instance of MsgAddAdmins
func NewMsgAddAdmins(sender string, admins ...string) *MsgAddAdmins { // nolint: interfacer
	return &MsgAddAdmins{
		Sender: sender,
		Admins: admins,
	}
}

// Route should return the name of the module
func (msg MsgAddAdmins) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAddAdmins) Type() string { return TypeMsgAddAdmins }

// ValidateBasic runs stateless checks on the message
func (msg MsgAddAdmins) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	for _, addr := range msg.Admins {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address '%s'", addr)
		}
	}
	return nil
}

// NewMsgRemoveAdmins creates a new instance of MsgRemoveAdmins
func NewMsgRemoveAdmins(sender string, admins ...string) *MsgRemoveAdmins { // nolint: interfacer
	return &MsgRemoveAdmins{
		Sender: sender,
		Admins: admins,
	}
}

// Route should return the name of the module
func (msg MsgRemoveAdmins) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRemoveAdmins) Type() string { return TypeMsgRemoveAdmins }

// ValidateBasic runs stateless checks on the message
func (msg MsgRemoveAdmins) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	for _, addr := range msg.Admins {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address '%s'", addr)
		}
	}
	return nil
}

// NewMsgEnable creates a new instance of MsgEnable
func NewMsgEnable(sender sdk.AccAddress) *MsgEnable { // nolint: interfacer
	return &MsgEnable{
		Sender: sender.String(),
	}
}

// Route should return the name of the module
func (msg MsgEnable) Route() string { return RouterKey }

// Type should return the action
func (msg MsgEnable) Type() string { return TypeMsgEnable }

// ValidateBasic runs stateless checks on the message
func (msg MsgEnable) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	return nil
}

// NewMsgDisable creates a new instance of MsgDisable
func NewMsgDisable(sender sdk.AccAddress, admins ...string) *MsgDisable { // nolint: interfacer
	return &MsgDisable{
		Sender: sender.String(),
	}
}

// Route should return the name of the module
func (msg MsgDisable) Route() string { return RouterKey }

// Type should return the action
func (msg MsgDisable) Type() string { return TypeMsgDisable }

// ValidateBasic runs stateless checks on the message
func (msg MsgDisable) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	return nil
}
