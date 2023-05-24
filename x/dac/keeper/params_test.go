package keeper_test

import (
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.DacKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()
	expParams.Enable = true

	suite.Require().Equal(expParams, params)

	params.Enable = false
	suite.app.DacKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.DacKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}
