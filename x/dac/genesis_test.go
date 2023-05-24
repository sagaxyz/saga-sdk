package dac_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	"github.com/sagaxyz/ethermint/crypto/ethsecp256k1"
	"github.com/sagaxyz/ethermint/tests"
	feemarkettypes "github.com/sagaxyz/ethermint/x/feemarket/types"

	"github.com/sagaxyz/sagaevm/v8/app"
	"github.com/sagaxyz/sagaevm/v8/x/dac"
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app     *app.SagaEvm
	genesis types.GenesisState
}

func (suite *GenesisTestSuite) SetupTest() {
	// consensus key
	consAddress := sdk.ConsAddress(tests.GenerateAddress().Bytes())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	params := types.DefaultParams()
	suite.app.DacKeeper.SetParams(suite.ctx, params)

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
			"custom genesis - enabled dac",
			types.GenesisState{
				Params: types.Params{
					Enable: true,
				},
				Admins:  []string{addr1.String()},
				Allowed: []string{ethAddr1.Hex()},
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
					dac.InitGenesis(suite.ctx, suite.app.DacKeeper, tc.genesis)
				})
			} else {
				suite.Require().NotPanics(func() {
					dac.InitGenesis(suite.ctx, suite.app.DacKeeper, tc.genesis)
				})

				params := suite.app.DacKeeper.GetParams(suite.ctx)
				suite.Require().Equal(params, tc.genesis.Params)

				for _, allowed := range tc.genesis.Allowed {
					allowed := suite.app.DacKeeper.Allowed(suite.ctx, common.HexToAddress(allowed))
					suite.Require().True(allowed)
				}
				for _, admin := range tc.genesis.Admins {
					admin := suite.app.DacKeeper.Admin(suite.ctx, sdk.MustAccAddressFromBech32(admin))
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

	suite.genesis.Allowed = []string{ethAddr1.Hex()}
	suite.genesis.Admins = []string{addr1.String()}

	dac.InitGenesis(suite.ctx, suite.app.DacKeeper, suite.genesis)
	genesisExported := dac.ExportGenesis(suite.ctx, suite.app.DacKeeper)

	suite.Require().Equal(genesisExported.Params, suite.genesis.Params)
	suite.Require().Equal(genesisExported.Admins, suite.genesis.Admins)
	suite.Require().Equal(genesisExported.Allowed, suite.genesis.Allowed)
}
