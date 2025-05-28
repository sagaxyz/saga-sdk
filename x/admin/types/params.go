package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store key
var (
	ParamPermissions = []byte("Permissions")
)

var _ paramtypes.ParamSet = &Params{}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamPermissions, &p.Permissions, validatePermissions),
	}
}

// NewParams creates a new Params object
func NewParams(permissions Permissions) Params {
	return Params{
		Permissions: permissions,
	}
}

// DefaultParams creates a parameter instance with default values
func DefaultParams() Params {
	return Params{
		Permissions: Permissions{
			SetMetadata: true,
		},
	}
}

func validatePermissions(i interface{}) error {
	_, ok := i.(Permissions)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	return nil
}
