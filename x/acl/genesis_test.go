package acl_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/acl"
	"github.com/sagaxyz/saga-sdk/x/acl/keeper"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	aclKeeper keeper.Keeper
	genesis   types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	params := types.DefaultParams()

	suite.genesis = *types.DefaultGenesis()
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	key1, _ := ethsecp256k1.GenerateKey()
	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes()) //TODO does this make sense?

	testCases := []struct {
		name     string
		genesis  types.GenesisState
		malleate func()
		expPanic bool
	}{
		{
			"default genesis",
			suite.genesis,
			func() {},
			false,
		},
		{
			"custom genesis - enabled acl",
			types.GenesisState{
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
					acl.InitGenesis(suite.ctx, suite.aclKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					acl.InitGenesis(suite.ctx, suite.aclKeeper, tc.genesis)
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

func (suite *GenesisTestSuite) TestExportGenesis() {
	suite.SetupTest()

	key1, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes()) //TODO does this make sense?

	suite.genesis.Admins = []*types.Address{
		{
			Format: types.AddressFormat_ADDRESS_BECH32,
			Value:  addr1.String(),
		},
	}
	suite.genesis.Allowed = []*types.Address{
		{
			Format: types.AddressFormat_ADDRESS_EIP55,
			Value:  ethAddr1.Hex(),
		},
	}

	acl.InitGenesis(suite.ctx, suite.aclKeeper, suite.genesis)
	genesisExported := acl.ExportGenesis(suite.ctx, suite.aclKeeper)

	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
	suite.Require().Equal(genesisExported.Admins, suite.genesis.Admins)
	suite.Require().Equal(genesisExported.Allowed, suite.genesis.Allowed)
}
