package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
}

//func (suite *GenesisTestSuite) SetupTest() {
//}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			name:     "default",
			genState: DefaultGenesis(),
			expPass:  true,
		},
		{
			name: "valid genesis - disabled",
			genState: &GenesisState{
				Params: DefaultParams(),
			},
			expPass: true,
		},
		{
			name: "valid genesis - with admins",
			genState: &GenesisState{
				Params: Params{
					Enable: true,
				},
				Admins: []string{"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc"},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with allowed",
			genState: &GenesisState{
				Params: Params{
					Enable: true,
				},
				Allowed: []string{"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc"},
			},
			expPass: true,
		},
		{
			name: "valid genesis - with allowed and admins",
			genState: &GenesisState{
				Params: Params{
					Enable: true,
				},
				Allowed: []string{"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc"},
				Admins:  []string{"cosmos1qql8ag4cluz6r4dz28p3w00dnc9w8ueulg2gmc"},
			},
			expPass: true,
		},
		{
			name: "invalid genesis - enabled and no admin or allowed",
			genState: &GenesisState{
				Params: Params{
					Enable: true,
				},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - bad admin address",
			genState: &GenesisState{
				Params: DefaultParams(),
				Admins: []string{"abcd"},
			},
			expPass: false,
		},
		{
			name: "invalid genesis - bad allowed address",
			genState: &GenesisState{
				Params:  DefaultParams(),
				Allowed: []string{"abcd"},
			},
			expPass: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}
