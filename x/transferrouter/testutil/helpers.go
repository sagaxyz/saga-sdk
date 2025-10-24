package testutil

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	callbacktypes "github.com/cosmos/ibc-go/v10/modules/apps/callbacks/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

// SetupKeeperWithMocks creates a test keeper with mocked dependencies
// This is a basic setup - callers should inject their own mocks as needed
func SetupKeeperWithMocks(t *testing.T) (sdk.Context, keeper.Keeper) {
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

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, // erc20 keeper
		nil, // ics4 wrapper
		nil, // channel keeper
		nil, // transfer keeper
		nil, // bank keeper
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

	return ctx, k
}

// CreateTestContext creates a test SDK context with block info
func CreateTestContext(t *testing.T) sdk.Context {
	t.Helper()

	key := storetypes.NewKVStoreKey("test")
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{"test": key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{
		Time:   tmtime.Now(),
		Height: 100,
	})
	ctx = ctx.WithChainID("saga_12345-1")
	ctx = ctx.WithTxBytes([]byte("test-tx-bytes"))

	return ctx
}

// CreateTestPacket creates a valid test IBC packet
func CreateTestPacket(sequence uint64, data []byte) channeltypes.Packet {
	return channeltypes.Packet{
		Sequence:           sequence,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 1000),
		TimeoutTimestamp:   0,
		Data:               data,
	}
}

// CreateTestPacketData creates valid FungibleTokenPacketData
func CreateTestPacketData(denom, amount, sender, receiver string) transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    denom,
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Memo:     "",
	}
}

// CreateTestPacketDataWithMemo creates FungibleTokenPacketData with memo
func CreateTestPacketDataWithMemo(denom, amount, sender, receiver, memo string) transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    denom,
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Memo:     memo,
	}
}

// CreateTestCallbackData creates valid callback data
func CreateTestCallbackData(address string, gasLimit uint64, calldata []byte) callbacktypes.CallbackData {
	return callbacktypes.CallbackData{
		CallbackAddress: address,
		CommitGasLimit:  gasLimit,
		Calldata:        calldata,
	}
}

// CreateTestPacketQueueItem creates a packet queue item for testing
func CreateTestPacketQueueItem(packet channeltypes.Packet, txHash []byte) types.PacketQueueItem {
	return types.PacketQueueItem{
		Packet:         &packet,
		OriginalTxHash: txHash,
	}
}

// CreateTestSrcCallbackQueueItem creates a src callback queue item for testing
func CreateTestSrcCallbackQueueItem(packet channeltypes.Packet, txHash []byte, isTimeout bool, ack []byte) types.PacketQueueItem {
	return types.PacketQueueItem{
		Packet:          &packet,
		OriginalTxHash:  txHash,
		IsTimeout:       isTimeout,
		Acknowledgement: ack,
	}
}

// MarshalPacketData marshals packet data to JSON bytes
func MarshalPacketData(t *testing.T, data transfertypes.FungibleTokenPacketData) []byte {
	t.Helper()
	bz := transfertypes.ModuleCdc.MustMarshalJSON(&data)
	return bz
}

// ValidTestAddress returns a valid bech32 cosmos address for testing
func ValidTestAddress() string {
	return "cosmos1abc123def456ghi789jkl012mno345pqr678st"
}

// ValidTestSenderAddress returns a valid sender address for testing
func ValidTestSenderAddress() string {
	return "cosmos1sender000000000000000000000000000hjkl"
}
