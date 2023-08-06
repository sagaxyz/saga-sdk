package keeper_test

import (
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/acl"
	"github.com/sagaxyz/saga-sdk/x/acl/keeper"
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func NewContextWithDB(t *testing.T, keys []storetypes.StoreKey, tkeys []storetypes.StoreKey) testutil.TestContext {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	for _, key := range keys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	}
	for _, tkey := range tkeys {
		cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	}
	err := cms.LoadLatestVersion()
	assert.NoError(t, err)

	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return testutil.TestContext{
		Ctx: ctx,
		DB:  db,
		CMS: cms,
	}
}

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
	key := sdk.NewKVStoreKey(types.StoreKey)
	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	paramsTKey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	testCtx := NewContextWithDB(suite.T(), []storetypes.StoreKey{paramsKey, key}, []storetypes.StoreKey{paramsTKey})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(acl.AppModuleBasic{})

	suite.ctx = ctx

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
	genesis.Admins = append(genesis.Admins, &types.Address{
		Format: types.AddressFormat_ADDRESS_BECH32,
		Value:  suite.adminAddress.String(),
	})
	suite.aclKeeper.InitGenesis(suite.ctx, genesis)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.aclKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient

	suite.encCfg = encCfg
}
