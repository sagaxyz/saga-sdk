package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	ParamsKey             = []byte("Params")
	ParamStoreKeyBaseFee  = []byte("BaseFee")
	ParamStoreKeyOverride = []byte("Override")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyBaseFee, &p.BaseFee, validateBaseFee),
		paramtypes.NewParamSetPair(ParamStoreKeyOverride, &p.Override, validateBool),
	}
}

// NewParams creates a new Params instance
func NewParams(override bool, baseFee sdkmath.Int) Params {
	return Params{
		Override: override,
		BaseFee:  baseFee,
	}
}

// DefaultParams returns default evm parameters
func DefaultParams() Params {
	return Params{}
}

// Validate performs basic validation on basefee parameters.
func (p Params) Validate() error {
	return nil
}

func validateBaseFee(i interface{}) error {
	value, ok := i.(*sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value.IsNegative() {
		return fmt.Errorf("base fee cannot be negative")
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
