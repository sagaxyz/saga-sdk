package keeper_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sagaxyz/ethermint/crypto/ethsecp256k1"
	"github.com/sagaxyz/ethermint/tests"
	feemarkettypes "github.com/sagaxyz/ethermint/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evm "github.com/sagaxyz/ethermint/x/evm/types"
	"github.com/sagaxyz/sagaevm/v8/app"
	"github.com/sagaxyz/sagaevm/v8/testutil"
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app            *app.SagaEvm
	queryClient    types.QueryClient
	queryClientEvm evm.QueryClient
	address        common.Address
	signer         keyring.Signer
	ethSigner      ethtypes.Signer
	validator      stakingtypes.Validator

	adminAddress sdk.AccAddress
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9001-1",
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

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.DacKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	params := types.DefaultParams()
	params.Enable = true
	suite.app.DacKeeper.SetParams(suite.ctx, params)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	validators := s.app.StakingKeeper.GetValidators(s.ctx, 1)
	suite.validator = validators[0]

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	// add admin
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.adminAddress = sdk.AccAddress(priv.PubKey().Address().Bytes())
	//suite.AdminSigner = tests.NewSigner(priv)
	suite.app.DacKeeper.SetAdmin(suite.ctx, suite.adminAddress)
}

func (suite *KeeperTestSuite) SetupTestWithEscrow() {
	suite.SetupTest()
	params := suite.app.ClaimsKeeper.GetParams(suite.ctx)

	coins := sdk.NewCoins(sdk.NewCoin(params.ClaimsDenom, sdk.NewInt(10000000)))
	err := testutil.FundModuleAccount(suite.app.BankKeeper, suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: header.Height})
	_ = suite.app.Commit()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.DacKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)
}
