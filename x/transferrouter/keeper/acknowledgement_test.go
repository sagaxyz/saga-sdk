package keeper_test

import (
	"fmt"
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

func setupKeeperWithMocks(t *testing.T) (sdk.Context, keeper.Keeper, *MockTransferKeeper, *MockBankKeeper, *MockICS4Wrapper) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{types.StoreKey: key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})

	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	mockTransferKeeper := new(MockTransferKeeper)
	mockBankKeeper := new(MockBankKeeper)
	mockICS4 := new(MockICS4Wrapper)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, // erc20 keeper
		mockICS4,
		nil, // channel keeper
		mockTransferKeeper,
		mockBankKeeper,
		nil, // account keeper
		nil, // evm keeper
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	// Set default params
	err := k.Params.Set(ctx, types.Params{
		Enabled:                true,
		KnownSignerPrivateKey:  "f6dba52e479cf5d7ad58bc11177c105ac7b89a02be1d432e77e113fc53377978",
		GatewayContractAddress: "0x5A6A8Ce46E34c2cd998129d013fA0253d3892345",
	})
	require.NoError(t, err)

	return ctx, k, mockTransferKeeper, mockBankKeeper, mockICS4
}

func TestWriteIBCAcknowledgment_SuccessAck(t *testing.T) {
	ctx, k, _, _, mockICS4 := setupKeeperWithMocks(t)

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "1000",
		Sender:   "saga1sender",
		Receiver: "osmo1receiver",
	}

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&data),
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	mockICS4.On("WriteAcknowledgement", ctx, mock.Anything, ack).Return(nil)

	err := k.WriteIBCAcknowledgment(ctx, packet, ack)
	require.NoError(t, err)

	mockICS4.AssertExpectations(t)
}

// Generic error acknowledgement case
func TestWriteIBCAcknowledgment_ErrorAck(t *testing.T) {
	ctx, k, _, _, mockICS4 := setupKeeperWithMocks(t)

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "1000",
		Sender:   "cosmos1sender",
		Receiver: "cosmos1receiver",
	}

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&data),
	}

	ack := channeltypes.NewErrorAcknowledgement(fmt.Errorf("test error"))

	mockICS4.On("WriteAcknowledgement", ctx, mock.Anything, ack).Return(nil)

	err := k.WriteIBCAcknowledgment(ctx, packet, ack)
	require.NoError(t, err)

	mockICS4.AssertExpectations(t)
}
