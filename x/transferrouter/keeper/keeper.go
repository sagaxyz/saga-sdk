package keeper

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/corecompat"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	// corestore "cosmossdk.io/core/store"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
	callbacktypes "github.com/sagaxyz/saga-sdk/x/transferrouter/v10types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
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
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
}

type TransferKeeper interface {
	DenomPathFromHash(ctx sdk.Context, denomHash string) (string, error)
	GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin
	SetTotalEscrowForDenom(sdk.Context, sdk.Coin)
}

type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
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

// WriteIBCAcknowledgment writes the IBC acknowledgment for the call queue item
func (k Keeper) WriteIBCAcknowledgment(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// WriteAcknowledgementForPacket writes an acknowledgement for a packet (copied from PFM)
func (k Keeper) WriteAcknowledgementForPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	ack channeltypes.Acknowledgement,
) error {
	// Lookup module by channel capability
	_, chanCap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if err != nil {
		return fmt.Errorf("could not retrieve module from port-id: %w", err)
	}

	// for packets w/callbacks, the funds were moved into an escrow account if the denom originated on this chain.
	// On an ack error or timeout on a packet w/callbacks, the funds in the escrow account
	// should be moved to the other escrow account on the other side or burnt.
	if !ack.Success() {
		// Override the receiver address to the gateway contract address
		params, err := k.Params.Get(ctx)
		if err != nil {
			k.Logger(ctx).Error("failed to get params", "error", err)
			return err
		}
		gatewayAddr := common.HexToAddress(params.GatewayContractAddress)
		escrowAddress := sdk.AccAddress(gatewayAddr.Bytes())

		// If it's a callback packet, we override the escrow address to the isolated address (as that's where the funds were received)
		_, isCbPacket, err := callbacktypes.GetCallbackData(data, callbacktypes.V1, packet.GetDestPort(), ctx.GasMeter().GasRemaining(), ctx.GasMeter().Limit(), callbacktypes.DestinationCallbackKey)

		if isCbPacket {
			if err != nil {
				// if isCbPacket is true and the error != nil, we have a malformed packet
				return fmt.Errorf("failed to get callback data: %w", err)
			}
			// Generate secure isolated address from sender and override the escrow address
			isolatedAddr := utils.GenerateIsolatedAddress(packet.GetDestChannel(), data.Sender)
			escrowAddress = isolatedAddr
		}

		fullDenomPath := getDenomForThisChain(
			packet.DestinationPort, packet.DestinationChannel,
			packet.SourcePort, packet.SourceChannel,
			data.Denom,
		)

		// deconstruct the token denomination into the denomination trace info
		// to determine if the sender is the source chain
		if strings.HasPrefix(data.Denom, "ibc/") {
			fullDenomPath, err = k.TransferKeeper.DenomPathFromHash(ctx, data.Denom)
			if err != nil {
				return err
			}
		}

		amount, ok := sdkmath.NewIntFromString(data.Amount)
		if !ok {
			return fmt.Errorf("failed to parse amount from packet data for forward refund: %s", data.Amount)
		}

		denomTrace := transfertypes.ParseDenomTrace(fullDenomPath)
		coin := sdk.NewCoin(denomTrace.IBCDenom(), amount)

		refundEscrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)

		newToken := sdk.NewCoins(coin)

		// Sender chain is source
		if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
			// funds were moved to escrow account for transfer, so they need to either:
			// - move to the other escrow account, in the case of native denom
			// - burn
			if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
				// transfer funds from escrow account for forwarded packet to escrow account going back for refund.
				if err := k.BankKeeper.SendCoins(
					ctx, escrowAddress, refundEscrowAddress, newToken,
				); err != nil {
					return fmt.Errorf("failed to send coins from escrow account to refund escrow account: %w", err)
				}
			} else {
				// transfer the coins from the escrow account to the module account and burn them.
				if err := k.BankKeeper.SendCoinsFromAccountToModule(
					ctx, escrowAddress, transfertypes.ModuleName, newToken,
				); err != nil {
					return fmt.Errorf("failed to send coins from escrow to module account for burn: %w", err)
				}

				if err := k.BankKeeper.BurnCoins(
					ctx, transfertypes.ModuleName, newToken,
				); err != nil {
					// NOTE: should not happen as the module account was
					// retrieved on the step above and it has enough balance
					// to burn.
					panic(fmt.Sprintf("cannot burn coins after a successful send from escrow account to module account: %v", err))
				}

				k.unescrowToken(ctx, coin)
			}
		} else {
			// Funds in the escrow account were burned,
			// so on a timeout or acknowledgement error we need to mint the funds back to the escrow account.
			if err := k.BankKeeper.MintCoins(ctx, transfertypes.ModuleName, newToken); err != nil {
				return fmt.Errorf("cannot mint coins to the %s module account: %v", transfertypes.ModuleName, err)
			}

			if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, transfertypes.ModuleName, refundEscrowAddress, newToken); err != nil {
				return fmt.Errorf("cannot send coins from the %s module to the escrow account %s: %v", transfertypes.ModuleName, refundEscrowAddress, err)
			}

			currentTotalEscrow := k.TransferKeeper.GetTotalEscrowForDenom(ctx, coin.GetDenom())
			newTotalEscrow := currentTotalEscrow.Add(coin)
			k.TransferKeeper.SetTotalEscrowForDenom(ctx, newTotalEscrow)
		}
	}

	return k.WriteIBCAcknowledgment(ctx, chanCap, channeltypes.Packet{
		Data:               packet.Data,
		Sequence:           packet.Sequence,
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
	}, ack)
}

// unescrowToken will update the total escrow by deducting the unescrowed token
// from the current total escrow.
func (k Keeper) unescrowToken(ctx sdk.Context, token sdk.Coin) {
	currentTotalEscrow := k.TransferKeeper.GetTotalEscrowForDenom(ctx, token.GetDenom())
	newTotalEscrow := currentTotalEscrow.Sub(token)
	k.TransferKeeper.SetTotalEscrowForDenom(ctx, newTotalEscrow)
}

func getDenomForThisChain(port, channel, counterpartyPort, counterpartyChannel, denom string) string {
	counterpartyPrefix := transfertypes.GetDenomPrefix(counterpartyPort, counterpartyChannel)
	if strings.HasPrefix(denom, counterpartyPrefix) {
		// unwind denom
		unwoundDenom := denom[len(counterpartyPrefix):]
		denomTrace := transfertypes.ParseDenomTrace(unwoundDenom)
		if denomTrace.Path == "" {
			// denom is now unwound back to native denom
			return unwoundDenom
		}
		// denom is still IBC denom
		return denomTrace.IBCDenom()
	}
	// append port and channel from this chain to denom
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
