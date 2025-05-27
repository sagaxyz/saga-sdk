package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	_ sdk.Msg = &MsgSetMetadata{}
)

const (
	TypeMsgSetMetadata        = "set_metadata"
	TypeMsgEnableSetMetadata  = "enable_set_metadata"
	TypeMsgDisableSetMetadata = "disable_set_metadata"
)

// NewMsgSetMetadata creates a new instance of MsgSetMetadata
func NewMsgSetMetadata(sender string, metadata banktypes.Metadata) *MsgSetMetadata { // nolint: interfacer
	return &MsgSetMetadata{
		Authority: sender,
		Metadata:  &metadata,
	}
}

// Route should return the name of the module
func (msg MsgSetMetadata) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetMetadata) Type() string { return TypeMsgSetMetadata }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}
	if msg.Metadata == nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "metadata cannot be nil")
	}
	if err := msg.Metadata.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

// NewMsgEnableSetMetadata creates a new instance of MsgEnableSetMetadata
func NewMsgEnableSetMetadata(authority string) *MsgEnableSetMetadata {
	return &MsgEnableSetMetadata{
		Authority: authority,
	}
}

// Route should return the name of the module
func (msg MsgEnableSetMetadata) Route() string { return RouterKey }

// Type should return the action
func (msg MsgEnableSetMetadata) Type() string { return TypeMsgEnableSetMetadata }

// ValidateBasic runs stateless checks on the message
func (msg MsgEnableSetMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}
	return nil
}

// NewMsgDisableSetMetadata creates a new instance of MsgDisableSetMetadata
func NewMsgDisableSetMetadata(authority string) *MsgDisableSetMetadata {
	return &MsgDisableSetMetadata{
		Authority: authority,
	}
}

// Route should return the name of the module
func (msg MsgDisableSetMetadata) Route() string { return RouterKey }

// Type should return the action
func (msg MsgDisableSetMetadata) Type() string { return TypeMsgDisableSetMetadata }

// ValidateBasic runs stateless checks on the message
func (msg MsgDisableSetMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}
	return nil
}
