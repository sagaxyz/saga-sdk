package keeper

import (
	"context"
	"math/big"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/corecompat"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	ParamsPrefix           = collections.NewPrefix(0) // Stores params
	PacketQueuePrefix      = collections.NewPrefix(2) // Stores the packets
	PacketResultPrefix     = collections.NewPrefix(3) // Stores the packet results
	SrcCallbackQueuePrefix = collections.NewPrefix(4) // Stores the src callback queue
)

type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
}

type TransferKeeper interface {
	DenomPathFromHash(ctx sdk.Context, denomHash string) (string, error)
	GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin
	SetTotalEscrowForDenom(sdk.Context, sdk.Coin)
}

type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type ERC20Keeper interface {
	GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error)
	GetTokenPairID(ctx sdk.Context, token string) []byte
	GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool)
	BalanceOf(ctx sdk.Context, abi abi.ABI, contract, account common.Address) *big.Int
}

type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, account sdk.AccountI)
	GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string)
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corecompat.KVStoreService
	authority    string

	Schema           collections.Schema
	Params           collections.Item[types.Params]
	PacketQueue      collections.Map[uint64, types.PacketQueueItem]
	SrcCallbackQueue collections.Map[uint64, types.PacketQueueItem]

	Erc20Keeper    ERC20Keeper
	ChannelKeeper  ChannelKeeper
	TransferKeeper TransferKeeper
	BankKeeper     BankKeeper
	AccountKeeper  AccountKeeper
	EVMKeeper      *evmkeeper.Keeper

	ics4Wrapper porttypes.ICS4Wrapper
}

// New returns a new Keeper instance.
func NewKeeper(cdc codec.BinaryCodec,
	storeSvc corecompat.KVStoreService,
	erc20Keeper ERC20Keeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper ChannelKeeper,
	transferKeeper TransferKeeper,
	bankKeeper BankKeeper,
	accountKeeper AccountKeeper,
	evmKeeper *evmkeeper.Keeper,
	authority string) Keeper {

	sb := collections.NewSchemaBuilder(storeSvc)
	k := Keeper{
		cdc:            cdc,
		storeService:   storeSvc,
		authority:      authority,
		Erc20Keeper:    erc20Keeper,
		ChannelKeeper:  channelKeeper,
		TransferKeeper: transferKeeper,
		BankKeeper:     bankKeeper,
		AccountKeeper:  accountKeeper,
		EVMKeeper:      evmKeeper,
		ics4Wrapper:    ics4Wrapper,
		Params: collections.NewItem(
			sb,
			ParamsPrefix,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		PacketQueue: collections.NewMap(
			sb,
			PacketQueuePrefix,
			"packet_queue",
			collections.Uint64Key,
			codec.CollValue[types.PacketQueueItem](cdc),
		),
		SrcCallbackQueue: collections.NewMap(
			sb,
			SrcCallbackQueuePrefix,
			"src_callback_queue",
			collections.Uint64Key,
			codec.CollValue[types.PacketQueueItem](cdc),
		),
	}

	var err error
	k.Schema, err = sb.Build()
	if err != nil {
		panic(err)
	}

	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// WriteIBCAcknowledgment writes the IBC acknowledgment for the call queue item.
// As we don't modify outgoing txs, we just pass this call to the original transferkeeper.
func (k Keeper) WriteIBCAcknowledgment(ctx sdk.Context, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, packet, ack)
}
