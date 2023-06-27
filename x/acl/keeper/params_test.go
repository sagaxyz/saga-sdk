package keeper_test

import (
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.keeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	expParams.Enable = true

	suite.Require().Equal(expParams, params)

	params.Enable = false
	suite.keeper.SetParams(suite.ctx, params)
	newParams := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
