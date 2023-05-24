package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamStoreKeyEnable = []byte("Enable")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEnable, &p.Enable, validateBool),
	}
}

// NewParams creates a new Params object
func NewParams(enable bool) Params {
	return Params{
		Enable: enable,
	}
}

// DefaultParams creates a parameter instance with default values
func DefaultParams() Params {
	return Params{
		Enable: false,
	}
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	return nil
}
