package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/acl"
	"github.com/sagaxyz/saga-sdk/x/acl/keeper"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

type TestSuite struct {
	suite.Suite

	ctx          sdk.Context
	aclKeeper    keeper.Keeper
	paramsKeeper paramskeeper.Keeper
	queryClient  types.QueryClient
	encCfg       moduletestutil.TestEncodingConfig

	adminAddress sdk.AccAddress
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
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
		nil)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	suite.ctx = ctx 
	encCfg := moduletestutil.MakeTestEncodingConfig(acl.AppModuleBasic{})

	suite.paramsKeeper = paramskeeper.NewKeeper(
		encCfg.Codec,
		encCfg.Amino,
		paramsKey,
		paramsTKey,
	)
	suite.paramsKeeper.Subspace(paramstypes.ModuleName).WithKeyTable(types.ParamKeyTable())

	suite.paramsKeeper.Subspace(types.ModuleName).WithKeyTable(types.ParamKeyTable())
	ss, ok := suite.paramsKeeper.GetSubspace(types.ModuleName)
	if !ok {
		panic("cannot get subspace")
	}

	suite.aclKeeper = keeper.New(
		encCfg.Codec,
		key,
		ss,
	)

	suite.adminAddress = sdk.AccAddress([]byte{123})
	genesis := types.DefaultGenesis()
	genesis.Params.Enable = true
	genesis.Admins = append(genesis.Admins, suite.adminAddress.String())
	suite.aclKeeper.InitGenesis(suite.ctx, genesis)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.aclKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient

	suite.encCfg = encCfg
}
