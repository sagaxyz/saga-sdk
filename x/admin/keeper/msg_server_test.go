package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sagaxyz/saga-sdk/x/acl/types"
)

func (suite *TestSuite) TestMsgServer() {
	suite.SetupTest()

	suite.Run("admins", func() {
		addr1 := sdk.AccAddress([]byte{111})
		addr2 := sdk.AccAddress([]byte{222})

		suite.Run("add", func() {
			suite.Require().False(suite.aclKeeper.Admin(suite.ctx, addr1))
			_, err := suite.aclKeeper.AddAdmins(suite.ctx, &types.MsgAddAdmins{
				Sender: addr1.String(),
				Admins: []string{addr1.String()},
			})
			suite.Require().Error(err)

			_, err = suite.aclKeeper.AddAdmins(suite.ctx, &types.MsgAddAdmins{
				Sender: suite.adminAddress.String(),
				Admins: []string{addr1.String()},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.aclKeeper.Admin(suite.ctx, addr1))
		})
		suite.Run("remove", func() {
			suite.Require().True(suite.aclKeeper.Admin(suite.ctx, addr1))
			suite.Require().False(suite.aclKeeper.Admin(suite.ctx, addr2))

			_, err := suite.aclKeeper.RemoveAdmins(suite.ctx, &types.MsgRemoveAdmins{
				Sender: addr2.String(),
				Admins: []string{addr1.String()},
			})
			suite.Require().Error(err)

			_, err = suite.aclKeeper.RemoveAdmins(suite.ctx, &types.MsgRemoveAdmins{
				Sender: addr1.String(),
				Admins: []string{addr1.String()},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.aclKeeper.Admin(suite.ctx, addr1))
		})
	})
	suite.Run("allowed", func() {
		addr := sdk.AccAddress([]byte{111})

		suite.Run("add", func() {
			suite.Require().False(suite.aclKeeper.Admin(suite.ctx, addr))
			suite.Require().False(suite.aclKeeper.Allowed(suite.ctx, addr))

			_, err := suite.aclKeeper.AddAllowed(suite.ctx, &types.MsgAddAllowed{
				Sender:  addr.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.aclKeeper.AddAllowed(suite.ctx, &types.MsgAddAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.aclKeeper.Allowed(suite.ctx, addr))
		})
		suite.Run("remove", func() {
			suite.Require().False(suite.aclKeeper.Admin(suite.ctx, addr))
			suite.Require().True(suite.aclKeeper.Allowed(suite.ctx, addr))

			_, err := suite.aclKeeper.RemoveAllowed(suite.ctx, &types.MsgRemoveAllowed{
				Sender:  addr.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.aclKeeper.RemoveAllowed(suite.ctx, &types.MsgRemoveAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.aclKeeper.Allowed(suite.ctx, addr))
		})
	})
}
