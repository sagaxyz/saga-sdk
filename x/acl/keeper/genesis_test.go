package keeper_test

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (suite *TestSuite) TestInitGenesis() {
	addr1 := sdk.AccAddress([]byte{123})

	testCases := []struct {
		name     string
		genesis  *types.GenesisState
		malleate func()
		expPanic bool
	}{
		{
			"default genesis",
			types.DefaultGenesis(),
			func() {},
			false,
		},
		{
			"custom genesis - enabled acl",
			&types.GenesisState{
				Params: types.Params{
					Enable: true,
				},
				Admins: []string{
					addr1.String(),
				},
				Allowed: []string{
					addr1.String(),
				},
			},
			func() {},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() {
					suite.aclKeeper.InitGenesis(suite.ctx, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					suite.aclKeeper.InitGenesis(suite.ctx, tc.genesis)
				})

				params := suite.aclKeeper.GetParams(suite.ctx)
				suite.Require().Equal(params, tc.genesis.Params)

				for _, allowed := range tc.genesis.Allowed {
					addr := sdk.MustAccAddressFromBech32(allowed)
					allowed := suite.aclKeeper.Allowed(suite.ctx, addr)
					suite.Require().True(allowed)
				}
				for _, admin := range tc.genesis.Admins {
					addr := sdk.MustAccAddressFromBech32(admin)
					admin := suite.aclKeeper.IsAdmin(suite.ctx, addr)
					suite.Require().True(admin)
				}
			}
		})
	}
}

func (suite *TestSuite) TestExportGenesis() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte{234})

	genesis := suite.aclKeeper.ExportGenesis(suite.ctx)
	genesis.Admins = append(genesis.Admins, addr1.String())
	genesis.Allowed = append(genesis.Allowed, addr1.String())

	suite.aclKeeper.InitGenesis(suite.ctx, genesis)
	genesisExported := suite.aclKeeper.ExportGenesis(suite.ctx)

	sort.Slice(genesis.Admins, func(i, j int) bool { return genesis.Admins[i] < genesis.Admins[j] })
	sort.Slice(genesis.Allowed, func(i, j int) bool { return genesis.Allowed[i] < genesis.Allowed[j] })
	sort.Slice(genesisExported.Admins, func(i, j int) bool { return genesisExported.Admins[i] < genesisExported.Admins[j] })
	sort.Slice(genesisExported.Allowed, func(i, j int) bool { return genesisExported.Allowed[i] < genesisExported.Allowed[j] })

	suite.Require().Equal(genesisExported.Params, genesis.Params)
	suite.Require().Equal(genesisExported.Admins, genesis.Admins)
	suite.Require().Equal(genesisExported.Allowed, genesis.Allowed)
}
