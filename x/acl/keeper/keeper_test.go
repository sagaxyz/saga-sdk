package keeper_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	evm "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/acl/keeper"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context

	keeper         keeper.Keeper
	queryClient    types.QueryClient
	queryClientEvm evm.QueryClient
	address        common.Address

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

	params := types.DefaultParams()
	params.Enable = true
	suite.keeper.SetParams(suite.ctx, params)

	// add admin
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.adminAddress = sdk.AccAddress(priv.PubKey().Address().Bytes())
	//suite.AdminSigner = tests.NewSigner(priv)
	suite.keeper.SetAdmin(suite.ctx, suite.adminAddress)
}