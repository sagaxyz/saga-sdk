package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/ethermint/crypto/ethsecp256k1"

	"github.com/sagaxyz/saga-sdk/x/dac/types"
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
			suite.Require().False(suite.keeper.Admin(suite.ctx, addr))
			_, err := suite.keeper.AddAdmins(ctx, &types.MsgAddAdmins{
				Sender: addr.String(),
				Admins: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_BECH32,
						Value:  addr.String(),
					},
				},
			})
			suite.Require().Error(err)

			_, err = suite.keeper.AddAdmins(ctx, &types.MsgAddAdmins{
				Sender: suite.adminAddress.String(),
				Admins: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_BECH32,
						Value:  addr.String(),
					},
				},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.keeper.Admin(suite.ctx, addr))
		})
		suite.Run("remove", func() {
			suite.Require().True(suite.keeper.Admin(suite.ctx, addr))
			suite.Require().False(suite.keeper.Admin(suite.ctx, addr2))

			_, err := suite.keeper.RemoveAdmins(ctx, &types.MsgRemoveAdmins{
				Sender: addr2.String(),
				Admins: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_BECH32,
						Value:  addr.String(),
					},
				},
			})
			suite.Require().Error(err)

			_, err = suite.keeper.RemoveAdmins(ctx, &types.MsgRemoveAdmins{
				Sender: addr.String(),
				Admins: []*types.Address{
					{
						Format: types.AddressFormat_ADDRESS_BECH32,
						Value:  addr.String(),
					},
				},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.keeper.Admin(suite.ctx, addr))
		})
	})
	suite.Run("allowed", func() {
		key, _ := ethsecp256k1.GenerateKey()
		addr := sdk.AccAddress(key.PubKey().Address())
		ethAddr := &types.Address{
			Format: types.AddressFormat_ADDRESS_EIP55,
			Value:  common.BytesToAddress(key.PubKey().Bytes()).Hex(),
		}

		suite.Run("add", func() {
			suite.Require().False(suite.keeper.Admin(suite.ctx, addr))
			suite.Require().False(suite.keeper.Allowed(suite.ctx, ethAddr))

			_, err := suite.keeper.AddAllowed(ctx, &types.MsgAddAllowed{
				Sender:  addr.String(),
				Allowed: []*types.Address{ethAddr},
			})
			suite.Require().Error(err)

			_, err = suite.keeper.AddAllowed(ctx, &types.MsgAddAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []*types.Address{ethAddr},
			})
			suite.Require().NoError(err)
			suite.Require().True(suite.keeper.Allowed(suite.ctx, ethAddr))
		})
		suite.Run("remove", func() {
			suite.Require().False(suite.keeper.Admin(suite.ctx, addr))
			suite.Require().True(suite.keeper.Allowed(suite.ctx, ethAddr))

			_, err := suite.keeper.RemoveAllowed(ctx, &types.MsgRemoveAllowed{
				Sender:  addr.String(),
				Allowed: []*types.Address{ethAddr},
			})
			suite.Require().Error(err)

			_, err = suite.keeper.RemoveAllowed(ctx, &types.MsgRemoveAllowed{
				Sender:  suite.adminAddress.String(),
				Allowed: []*types.Address{ethAddr},
			})
			suite.Require().NoError(err)
			suite.Require().False(suite.keeper.Allowed(suite.ctx, ethAddr))
		})
	})
}
