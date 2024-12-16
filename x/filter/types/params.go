package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	ParamsKey             = []byte("Params")
	ParamStoreKeyPrefixes = []byte("Prefixes")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyPrefixes, &p.Prefixes, validateStringSlice),
	}
}

// NewParams creates a new Params instance
func NewParams(prefixes ...string) Params {
	return Params{
		Prefixes: prefixes,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{}
}

// Validate performs basic validation on filter parameters.
func (p Params) Validate() error {
	return nil
}

func validateStringSlice(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
