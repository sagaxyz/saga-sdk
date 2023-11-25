package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid",
			NewParams(true, "cosmos147klh7th5jkjy3aajsj2rqvhtvh9mfde37wq5g"),
			false,
		},
		{
			"empty",
			Params{},
			false,
		},
		{
			"enabled without recipient",
			NewParams(true, ""),
			true,
		},
		{
			"enabled with an invalid address",
			NewParams(true, "abcd"),
			true,
		},
		{
			"disabled with an invalid address",
			NewParams(false, "abcd"),
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

func (suite *ParamsTestSuite) TestParamsValidatePriv() {
	suite.Require().NoError(validateBool(true))
	suite.Require().Error(validateBool(2))
	suite.Require().NoError(validateRecipient("cosmos147klh7th5jkjy3aajsj2rqvhtvh9mfde37wq5g"))
	suite.Require().NoError(validateRecipient(""))
	suite.Require().Error(validateRecipient("2141231"))
}
