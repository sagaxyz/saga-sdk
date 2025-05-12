package keeper_test

import (
	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (suite *TestSuite) TestParams() {
	params := suite.aclKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	expParams.Enable = true

	suite.Require().Equal(expParams, params)

	params.Enable = false
	suite.aclKeeper.SetParams(suite.ctx, params)
	newParams := suite.aclKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
