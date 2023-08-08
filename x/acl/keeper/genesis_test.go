package keeper_test

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sagaxyz/saga-sdk/crypto/ethsecp256k1"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (suite *TestSuite) TestInitGenesis() {
	key1, _ := ethsecp256k1.GenerateKey()
	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes())

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
				Admins: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_BECH32,
						Value:  addr1.String(),
					},
				},
				Allowed: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_EIP55,
						Value:  ethAddr1.Hex(),
					},
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
					allowed := suite.aclKeeper.Allowed(suite.ctx, allowed)
					suite.Require().True(allowed)
				}
				for _, admin := range tc.genesis.Admins {
					admin := suite.aclKeeper.Admin(suite.ctx, sdk.MustAccAddressFromBech32(admin.Value))
					suite.Require().True(admin)
				}
			}
		})
	}
}

func (suite *TestSuite) TestExportGenesis() {
	suite.SetupTest()

	key1, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes())

	genesis := suite.aclKeeper.ExportGenesis(suite.ctx)
	genesis.Admins = append(genesis.Admins, &types.Address{
		Format: types.AddressFormat_ADDRESS_BECH32,
		Value:  addr1.String(),
	})
	genesis.Allowed = append(genesis.Allowed, &types.Address{
		Format: types.AddressFormat_ADDRESS_EIP55,
		Value:  ethAddr1.Hex(),
	})

	suite.aclKeeper.InitGenesis(suite.ctx, genesis)
	genesisExported := suite.aclKeeper.ExportGenesis(suite.ctx)

	sort.Slice(genesis.Admins, func(i, j int) bool { return genesis.Admins[i].Value < genesis.Admins[j].Value })
	sort.Slice(genesis.Allowed, func(i, j int) bool { return genesis.Allowed[i].Value < genesis.Allowed[j].Value })
	sort.Slice(genesisExported.Admins, func(i, j int) bool { return genesisExported.Admins[i].Value < genesisExported.Admins[j].Value })
	sort.Slice(genesisExported.Allowed, func(i, j int) bool { return genesisExported.Allowed[i].Value < genesisExported.Allowed[j].Value })

	suite.Require().Equal(genesisExported.Params, genesis.Params)
	suite.Require().Equal(genesisExported.Admins, genesis.Admins)
	suite.Require().Equal(genesisExported.Allowed, genesis.Allowed)
}
