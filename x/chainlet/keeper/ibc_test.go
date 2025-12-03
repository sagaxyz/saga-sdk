package keeper_test

import (
	"errors"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	chainlettypes "github.com/sagaxyz/saga-sdk/x/chainlet/types"
	sdkchainlettypes "github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

func (s *TestSuite) packetVerificationMocks(providerClientID, clientID, connectionID, channelID string) {
	gomock.InOrder(
		s.consumerKeeper.EXPECT().
			GetProviderClientID(gomock.Any()).
			Return(providerClientID, nil),
		s.channelKeeper.EXPECT().
			GetChannel(gomock.Any(), sdkchainlettypes.PortID, gomock.Eq(channelID)).
			Return(ibcchanneltypes.Channel{
				ConnectionHops: []string{connectionID},
			}, true),
		s.connectionKeeper.EXPECT().
			GetConnection(gomock.Any(), gomock.Eq(connectionID)).
			Return(ibcconnectiontypes.ConnectionEnd{
				ClientId:     clientID,
				Versions:     []*ibcconnectiontypes.Version{},
				State:        0,
				Counterparty: ibcconnectiontypes.Counterparty{},
				DelayPeriod:  0,
			}, true),
	)
}

func (s *TestSuite) TestCreateUpgradePacket() {
	tests := []struct {
		name string
		fn   func(chainID, clientID, connectionID, channelID string)
	}{
		{
			name: "recv success",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: channelID,
				}
				data := chainlettypes.CreateUpgradePacketData{
					ChainId: chainID,
					Name:    "xxx",
					Height:  123,
					Info:    "xyz",
				}

				s.packetVerificationMocks(clientID, clientID, connectionID, channelID)
				gomock.InOrder(
					s.upgradeKeeper.EXPECT().
						GetUpgradePlan(gomock.Any()).
						Return(upgradetypes.Plan{}, upgradetypes.ErrNoUpgradePlanFound),
					s.upgradeKeeper.EXPECT().
						ScheduleUpgrade(gomock.Any(), upgradetypes.Plan{
							Name:   "xxx",
							Height: 123,
							Info:   "xyz",
						}).
						Return(nil),
				)
				_, err := s.chainletKeeper.OnRecvCreateUpgradePacket(s.ctx, packet, data)
				s.Require().NoError(err)
			},
		}, {
			name: "recv failure: incorrect client ID",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: "channel-42",
				}
				data := chainlettypes.CreateUpgradePacketData{
					ChainId: chainID,
					Name:    "xxx",
					Height:  123,
					Info:    "xyz",
				}

				s.packetVerificationMocks(clientID, "bad-client", "bad-connection", "channel-42")
				_, err := s.chainletKeeper.OnRecvCreateUpgradePacket(s.ctx, packet, data)
				s.Require().Error(err)
			},
		}, {
			name: "recv failure: plan already exists",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: channelID,
				}
				data := chainlettypes.CreateUpgradePacketData{
					ChainId: chainID,
					Name:    "xxx",
					Height:  123,
					Info:    "xyz",
				}

				s.packetVerificationMocks(clientID, clientID, connectionID, channelID)
				gomock.InOrder(
					s.upgradeKeeper.EXPECT().
						GetUpgradePlan(gomock.Any()).
						Return(upgradetypes.Plan{
							Name: "something",
						}, nil),
				)
				_, err := s.chainletKeeper.OnRecvCreateUpgradePacket(s.ctx, packet, data)
				s.Require().Error(err)
			},
		},
	}
	for i, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			chainID := fmt.Sprintf("chain_%d-1", i+1)
			clientID := fmt.Sprintf("client-%d", i)
			connectionID := fmt.Sprintf("connection-%d", i)
			channelID := fmt.Sprintf("channel-%d", i)

			tt.fn(chainID, clientID, connectionID, channelID)
		})
	}
}

func (s *TestSuite) TestCancelUpgradePacket() {
	tests := []struct {
		name string
		fn   func(chainID, clientID, connectionID, channelID string)
	}{
		{
			name: "recv success",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: channelID,
				}
				data := chainlettypes.CancelUpgradePacketData{
					ChainId: chainID,
					Plan:    "1-to-2",
				}

				s.packetVerificationMocks(clientID, clientID, connectionID, channelID)
				gomock.InOrder(
					s.upgradeKeeper.EXPECT().
						GetUpgradePlan(gomock.Any()).
						Return(upgradetypes.Plan{
							Name: "1-to-2",
						}, nil),
					s.upgradeKeeper.EXPECT().
						ClearUpgradePlan(gomock.Any()).
						Return(nil),
				)
				_, err := s.chainletKeeper.OnRecvCancelUpgradePacket(s.ctx, packet, data)
				s.Require().NoError(err)
			},
		}, {
			name: "recv failure: incorrect client ID",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: "channel-42",
				}
				data := chainlettypes.CancelUpgradePacketData{
					ChainId: chainID,
					Plan:    "1-to-2",
				}

				s.packetVerificationMocks(clientID, "bad-client", "bad-connection", "channel-42")
				_, err := s.chainletKeeper.OnRecvCancelUpgradePacket(s.ctx, packet, data)
				s.Require().Error(err)
			},
		}, {
			name: "recv success: no plan to cancel",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: channelID,
				}
				data := chainlettypes.CancelUpgradePacketData{
					ChainId: chainID,
					Plan:    "1-to-2",
				}

				s.packetVerificationMocks(clientID, clientID, connectionID, channelID)
				gomock.InOrder(
					s.upgradeKeeper.EXPECT().
						GetUpgradePlan(gomock.Any()).
						Return(upgradetypes.Plan{}, upgradetypes.ErrNoUpgradePlanFound),
				)
				_, err := s.chainletKeeper.OnRecvCancelUpgradePacket(s.ctx, packet, data)
				s.Require().NoError(err)
			},
		}, {
			name: "recv failure: cancel error",
			fn: func(chainID, clientID, connectionID, channelID string) {
				packet := channeltypes.Packet{
					DestinationChannel: channelID,
				}
				data := chainlettypes.CancelUpgradePacketData{
					ChainId: chainID,
					Plan:    "1-to-2",
				}

				s.packetVerificationMocks(clientID, clientID, connectionID, channelID)
				gomock.InOrder(
					s.upgradeKeeper.EXPECT().
						GetUpgradePlan(gomock.Any()).
						Return(upgradetypes.Plan{
							Name: "1-to-2",
						}, nil),
					s.upgradeKeeper.EXPECT().
						ClearUpgradePlan(gomock.Any()).
						Return(errors.New("error")),
				)
				_, err := s.chainletKeeper.OnRecvCancelUpgradePacket(s.ctx, packet, data)
				s.Require().Error(err)
			},
		},
	}
	for i, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			chainID := fmt.Sprintf("chain_%d-1", i+1)
			clientID := fmt.Sprintf("client-%d", i)
			connectionID := fmt.Sprintf("connection-%d", i)
			channelID := fmt.Sprintf("channel-%d", i)

			tt.fn(chainID, clientID, connectionID, channelID)
		})
	}
}
