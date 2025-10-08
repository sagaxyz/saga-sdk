package types

import (
	"errors"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	ParamStoreKeyTimeoutHeight = []byte("TimeoutHeight")
	ParamStoreKeyTimeoutTime   = []byte("TimeoutTime")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(timeoutHeight uint64, timeoutTime time.Duration) Params {
	return Params{
		TimeoutHeight: timeoutHeight,
		TimeoutTime:   timeoutTime,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		1000, 24*time.Hour,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyTimeoutHeight, &p.TimeoutHeight, validateUint64),
		paramtypes.NewParamSetPair(ParamStoreKeyTimeoutTime, &p.TimeoutTime, validateDuration),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

func validateUint64(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return errors.New("param not uint64")
	}
	return nil
}

func validateDuration(v interface{}) error {
	vv, ok := v.(time.Duration)
	if !ok {
		return errors.New("param not duration")
	}
	if vv < 0 {
		return errors.New("duration negative")
	}
	return nil
}
