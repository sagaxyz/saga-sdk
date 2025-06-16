package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	testCases := []struct {
		name    string
		gs      *GenesisState
		expPass bool
	}{
		{
			name:    "default",
			gs:      DefaultGenesis(),
			expPass: true,
		},
		{
			name: "valid genesis - enabled",
			gs: &GenesisState{
				Params: DefaultParams(),
			},
			expPass: true,
		},
		{
			name: "valid genesis - disabled",
			gs: &GenesisState{
				Params: Params{
					Permissions: Permissions{
						SetMetadata: false,
					},
				},
			},
			expPass: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.gs.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
