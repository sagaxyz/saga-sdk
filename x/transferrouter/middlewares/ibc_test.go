package middlewares

import (
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

// mock underlying IBCModule which simply echoes OnRecvPacket success

var _ porttypes.IBCModule = mockApp{}
var _ porttypes.PacketDataUnmarshaler = mockApp{}

type mockApp struct{}

func (m mockApp) UnmarshalPacketData(ctx sdk.Context, channelVersion string, proto string, bz []byte) (any, string, error) {
	return bz, "", nil
}

// OnChanCloseConfirm implements types.IBCModule.
func (m mockApp) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanCloseInit implements types.IBCModule.
func (m mockApp) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenAck implements types.IBCModule.
func (m mockApp) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	panic("unimplemented")
}

// OnChanOpenConfirm implements types.IBCModule.
func (m mockApp) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanOpenInit implements types.IBCModule.
func (m mockApp) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, counterparty channeltypes.Counterparty, version string) (string, error) {
	panic("unimplemented")
}

// OnChanOpenTry implements types.IBCModule.
func (m mockApp) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	panic("unimplemented")
}

// OnTimeoutPacket implements types.IBCModule.
func (m mockApp) OnTimeoutPacket(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	panic("unimplemented")
}

func (m mockApp) OnAcknowledgementPacket(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return nil
}

func (m mockApp) OnRecvPacket(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	// pretend ICS20 app accepted the packet
	return channeltypes.NewResultAcknowledgement([]byte{1})
}

func buildMiddleware(t *testing.T) (sdk.Context, IBCMiddleware, keeper.Keeper) {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{types.StoreKey: key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	// ensure tx bytes are set for tmhash
	ctx = ctx.WithTxBytes([]byte("tx-bytes"))

	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	// keeper with only store service and params collection used
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, nil, nil, nil, nil, nil, nil,
		"",
	)

	// set params needed by OnRecvPacket
	require.NoError(t, k.Params.Set(ctx, types.Params{
		Enabled:                true,
		KnownSignerPrivateKey:  "",
		GatewayContractAddress: common.HexToAddress("0x5A6A8Ce46E34c2cd998129d013fA0253d3892345").Hex(),
	}))

	// Compose the app that satisfies IBCModuleWithUnmarshaler
	app := mockApp{}
	mw := NewIBCMiddleware(app, app, 1000000, k)
	return ctx, mw, k
}

func Test_addSrcCallbackToQueue_ack_and_timeout(t *testing.T) {
	ctx, mw, _ := buildMiddleware(t)

	// create ICS20 packet with memo containing src_callback
	memo := map[string]any{
		"src_callback": map[string]any{
			"address":   "0x0000000000000000000000000000000000000001",
			"gas_limit": "1000",
		},
	}
	memoBz, _ := json.Marshal(memo)
	data := transfertypes.FungibleTokenPacketData{Denom: "usaga", Amount: "1", Sender: "s", Receiver: "r", Memo: string(memoBz)}
	bz := transfertypes.ModuleCdc.MustMarshalJSON(&data)

	pkt := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               bz,
	}

	// ack path
	ack := channeltypes.NewResultAcknowledgement([]byte{0x01})
	err := mw.addSrcCallbackToQueue(ctx, pkt, ack.Acknowledgement(), false)
	require.NoError(t, err)

	// timeout path
	err = mw.addSrcCallbackToQueue(ctx, pkt, nil, true)
	require.NoError(t, err)
}

func Test_newErrorAcknowledgement(t *testing.T) {
	ack := newErrorAcknowledgement(assertAnError{})
	require.False(t, ack.Success())
	require.Contains(t, string(ack.Acknowledgement()), "transfer-router error:")
}

type assertAnError struct{}

func (assertAnError) Error() string { return "boom" }
