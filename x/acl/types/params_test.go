package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"success - zero value",
			Params{},
			false,
		},
		{
			"success - constructor with true",
			NewParams(true),
			false,
		},
		{
			"success - constructor with false",
			NewParams(false),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()
		if tc.expError {
			assert.Error(t, err, tc.name)
		} else {
			assert.NoError(t, err, tc.name)
		}
	}
}

func TestParamsValidateBool(t *testing.T) {
	err := validateBool(true)
	assert.NoError(t, err)
	err = validateBool(false)
	assert.NoError(t, err)
	err = validateBool("")
	assert.Error(t, err)
	err = validateBool(int64(123))
	assert.Error(t, err)
}
