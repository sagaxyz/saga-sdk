package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/ethermint/crypto/ethsecp256k1"
	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

func (suite *KeeperTestSuite) TestMsgServer() {
	suite.SetupTest()

	ctx := sdk.WrapSDKContext(suite.ctx)
	suite.Run("admins", func() {
		key, _ := ethsecp256k1.GenerateKey()
		addr := sdk.AccAddress(key.PubKey().Address())
		key2, _ := ethsecp256k1.GenerateKey()
		addr2 := sdk.AccAddress(key2.PubKey().Address())

		suite.Run("add", func() {
			suite.Require().False(suite.app.DacKeeper.Admin(suite.ctx, addr))
			_, err := suite.app.DacKeeper.AddAdmins(ctx, &types.MsgAddAdmins{
				Sender: addr.String(),
				Admins: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.app.DacKeeper.AddAdmins(ctx, &types.MsgAddAdmins{
				Sender: suite.adminAddress.String(),
				Admins: []string{addr.String()},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.app.DacKeeper.Admin(suite.ctx, addr))
		})
		suite.Run("remove", func() {
			suite.Require().True(suite.app.DacKeeper.Admin(suite.ctx, addr))
			suite.Require().False(suite.app.DacKeeper.Admin(suite.ctx, addr2))

			_, err := suite.app.DacKeeper.RemoveAdmins(ctx, &types.MsgRemoveAdmins{
				Sender: addr2.String(),
				Admins: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.app.DacKeeper.RemoveAdmins(ctx, &types.MsgRemoveAdmins{
				Sender: addr.String(),
				Admins: []string{addr.String()},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.app.DacKeeper.Admin(suite.ctx, addr))
		})
	})
	suite.Run("allowed", func() {
		key, _ := ethsecp256k1.GenerateKey()
		addr := sdk.AccAddress(key.PubKey().Address())
		ethAddr := common.BytesToAddress(key.PubKey().Bytes())

		suite.Run("add", func() {
			suite.Require().False(suite.app.DacKeeper.Admin(suite.ctx, addr))
			suite.Require().False(suite.app.DacKeeper.Allowed(suite.ctx, ethAddr))

			_, err := suite.app.DacKeeper.AddAllowed(ctx, &types.MsgAddAllowed{
				Sender:  addr.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.app.DacKeeper.AddAllowed(ctx, &types.MsgAddAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []string{ethAddr.Hex()},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.app.DacKeeper.Allowed(suite.ctx, ethAddr))
		})
		suite.Run("remove", func() {
			suite.Require().False(suite.app.DacKeeper.Admin(suite.ctx, addr))
			suite.Require().True(suite.app.DacKeeper.Allowed(suite.ctx, ethAddr))

			_, err := suite.app.DacKeeper.RemoveAllowed(ctx, &types.MsgRemoveAllowed{
				Sender:  addr.String(),
				Allowed: []string{addr.String()},
			})
			suite.Require().Error(err)

			_, err = suite.app.DacKeeper.RemoveAllowed(ctx, &types.MsgRemoveAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []string{ethAddr.Hex()},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.app.DacKeeper.Allowed(suite.ctx, ethAddr))
		})
	})
}
