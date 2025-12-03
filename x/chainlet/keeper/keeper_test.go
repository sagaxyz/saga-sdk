package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/chainlet"
	"github.com/sagaxyz/saga-sdk/x/chainlet/keeper"
	chainlettestutil "github.com/sagaxyz/saga-sdk/x/chainlet/testutil"
	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

type TestSuite struct {
	suite.Suite

	chainletKeeper   keeper.Keeper
	ctx              sdk.Context
	msgServer        types.MsgServer
	consumerKeeper   *chainlettestutil.MockConsumerKeeper
	clientKeeper     *chainlettestutil.MockClientKeeper
	channelKeeper    *chainlettestutil.MockChannelKeeper
	connectionKeeper *chainlettestutil.MockConnectionKeeper
	upgradeKeeper    *chainlettestutil.MockUpgradeKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(chainlet.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	paramsKey := storetypes.NewKVStoreKey(paramstypes.StoreKey)
	paramsTKey := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{
			types.StoreKey:       key,
			paramstypes.StoreKey: paramsKey,
		},
		map[string]*storetypes.TransientStoreKey{
			paramstypes.TStoreKey: paramsTKey,
		},
		nil,
	)
	s.ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})

	ctrl := gomock.NewController(s.T())
	s.upgradeKeeper = chainlettestutil.NewMockUpgradeKeeper(ctrl)
	s.clientKeeper = chainlettestutil.NewMockClientKeeper(ctrl)
	s.channelKeeper = chainlettestutil.NewMockChannelKeeper(ctrl)
	s.connectionKeeper = chainlettestutil.NewMockConnectionKeeper(ctrl)
	s.consumerKeeper = chainlettestutil.NewMockConsumerKeeper(ctrl)

	s.chainletKeeper = keeper.New(
		encCfg.Codec,
		key,
		s.upgradeKeeper,
		s.channelKeeper,
		s.consumerKeeper,
		s.clientKeeper,
		s.connectionKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	s.msgServer = keeper.NewMsgServerImpl(s.chainletKeeper)

	err := s.chainletKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}
