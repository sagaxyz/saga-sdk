package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	ParamsKey              = []byte("Params")
	ParamStoreKeyEnabled   = []byte("Enabled")
	ParamStoreKeyRecipient = []byte("Recipient")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyRecipient, &p.Recipient, validateRecipient),
		paramtypes.NewParamSetPair(ParamStoreKeyEnabled, &p.Enabled, validateBool),
	}
}

// NewParams creates a new Params instance
func NewParams(enabled bool, recipient string) Params {
	return Params{
		Enabled:   enabled,
		Recipient: recipient,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{}
}

// Validate performs basic validation on feedistribution parameters.
func (p Params) Validate() error {
	if p.Enabled && p.Recipient == "" {
		return errors.New("cannot be enabled without a recipient address")
	}
	if p.Recipient != "" {
		_, err := sdk.AccAddressFromBech32(p.Recipient)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateRecipient(i interface{}) error {
	value, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if value != "" {
		_, err := sdk.AccAddressFromBech32(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
