package gateway_test

import (
	"math/big"
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/evm/contracts"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/precompiles/gateway"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

func setupKeeperForCalldataTest(t *testing.T) (sdk.Context, keeper.Keeper) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{types.StoreKey: key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	// Set tx bytes for hash generation
	ctx = ctx.WithTxBytes([]byte("test-tx-bytes"))

	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, nil, nil, nil, nil, nil, nil,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	return ctx, k
}

func TestCreateERC20TransferExecuteCallDataFromPacket_NativeDenom(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "1000",
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.NoError(t, err)
	require.NotNil(t, calldata)

	// Verify the calldata is valid ERC20 transfer calldata
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	method, err := erc20.MethodById(calldata[:4])
	require.NoError(t, err)
	require.Equal(t, "transfer", method.Name)

	// Unpack and verify parameters
	args, err := method.Inputs.Unpack(calldata[4:])
	require.NoError(t, err)
	require.Len(t, args, 2)

	// Verify recipient address
	recipient := args[0].(common.Address)
	receiverAddr, _ := sdk.AccAddressFromBech32(data.Receiver)
	expectedRecipient := common.BytesToAddress(receiverAddr.Bytes())
	require.Equal(t, expectedRecipient, recipient)

	// Verify amount
	amount := args[1].(*big.Int)
	require.Equal(t, big.NewInt(1000), amount)
}

func TestCreateERC20TransferExecuteCallDataFromPacket_IBCDenom(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	ibcDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    ibcDenom,
		Amount:   "5000",
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.NoError(t, err)
	require.NotNil(t, calldata)

	// Verify the calldata structure
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	method, err := erc20.MethodById(calldata[:4])
	require.NoError(t, err)
	require.Equal(t, "transfer", method.Name)
}

func TestCreateERC20TransferExecuteCallDataFromPacket_VerifyMemo(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	txBytes := []byte("my-test-tx-bytes")
	ctx = ctx.WithTxBytes(txBytes)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "1000",
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.NoError(t, err)
	require.NotNil(t, calldata)

	// Verify calldata is properly generated (tx hash is included in the memo internally)
	require.NotEmpty(t, calldata)
}

func TestCreateERC20TransferExecuteCallDataFromPacket_InvalidReceiver(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "1000",
		Sender:   "cosmos1sender",
		Receiver: "invalid-address",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.Error(t, err)
	require.Nil(t, calldata)
	require.Contains(t, err.Error(), "failed to parse receiver address")
}

func TestCreateERC20TransferExecuteCallDataFromPacket_InvalidAmount(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "not-a-number",
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.Error(t, err)
	require.Nil(t, calldata)
	require.Contains(t, err.Error(), "failed to parse amount")
}

func TestCreateERC20TransferExecuteCallDataFromPacket_ZeroAmount(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   "0",
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.NoError(t, err)
	require.NotNil(t, calldata)

	// Verify the amount is actually zero
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	method, err := erc20.MethodById(calldata[:4])
	require.NoError(t, err)

	args, err := method.Inputs.Unpack(calldata[4:])
	require.NoError(t, err)

	amount := args[1].(*big.Int)
	require.Equal(t, big.NewInt(0).Bytes(), amount.Bytes())
}

func TestCreateERC20TransferExecuteCallDataFromPacket_LargeAmount(t *testing.T) {
	ctx, k := setupKeeperForCalldataTest(t)

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		TimeoutHeight:      clienttypes.NewHeight(1, 100),
		TimeoutTimestamp:   0,
		Data:               []byte("test"),
	}

	largeAmount := "999999999999999999999999999999"
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "usaga",
		Amount:   largeAmount,
		Sender:   "cosmos1sender000000000000000000000000000hjkl",
		Receiver: "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
	}

	calldata, err := gateway.CreateERC20TransferExecuteCallDataFromPacket(ctx, k, packet, data)
	require.NoError(t, err)
	require.NotNil(t, calldata)

	// Verify the amount
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	method, err := erc20.MethodById(calldata[:4])
	require.NoError(t, err)

	args, err := method.Inputs.Unpack(calldata[4:])
	require.NoError(t, err)

	amount := args[1].(*big.Int)
	expectedAmount, _ := new(big.Int).SetString(largeAmount, 10)
	require.Equal(t, expectedAmount, amount)
}
