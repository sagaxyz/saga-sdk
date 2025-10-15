package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

func setupKeeperForTest(t *testing.T) (sdk.Context, keeper.Keeper, *MockICS4Wrapper) {
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

	mockICS4 := new(MockICS4Wrapper)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, // erc20 keeper
		mockICS4,
		nil, // channel keeper
		nil, // transfer keeper
		nil, // bank keeper
		nil, // account keeper
		nil, // evm keeper
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	return ctx, k, mockICS4
}

func TestNewKeeper(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		authority,
	)

	// Verify keeper fields are initialized
	require.NotNil(t, k)
	require.NotNil(t, k.Schema)
	require.NotNil(t, k.Params)
	require.NotNil(t, k.PacketQueue)
	require.NotNil(t, k.SrcCallbackQueue)
}

func TestKeeper_Logger(t *testing.T) {
	ctx, k, _ := setupKeeperForTest(t)

	logger := k.Logger(ctx)
	require.NotNil(t, logger)
}

func TestKeeper_WriteIBCAcknowledgment_Success(t *testing.T) {
	ctx, k, mockICS4 := setupKeeperForTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               []byte("test"),
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	mockICS4.On("WriteAcknowledgement", ctx, &packet, ack).Return(nil)

	err := k.WriteIBCAcknowledgment(ctx, &packet, ack)
	require.NoError(t, err)

	mockICS4.AssertExpectations(t)
}

func TestKeeper_WriteIBCAcknowledgment_Error(t *testing.T) {
	ctx, k, mockICS4 := setupKeeperForTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               []byte("test"),
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{1})

	expectedErr := channeltypes.ErrInvalidChannelState
	mockICS4.On("WriteAcknowledgement", ctx, &packet, ack).Return(expectedErr)

	err := k.WriteIBCAcknowledgment(ctx, &packet, ack)
	require.Error(t, err)
	require.Equal(t, expectedErr, err)

	mockICS4.AssertExpectations(t)
}
