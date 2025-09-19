package keeper

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	corestore "cosmossdk.io/core/store"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

var (
	ParamsPrefix       = collections.NewPrefix(0) // Stores params
	PacketQueuePrefix  = collections.NewPrefix(2) // Stores the packets
	PacketResultPrefix = collections.NewPrefix(3) // Stores the packet results
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
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins // TODO: remove this, just for debugging
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
}

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService corestore.KVStoreService
	authority    string

	Schema      collections.Schema
	Params      collections.Item[types.Params]
	PacketQueue collections.Map[uint64, channeltypes.Packet]

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
	storeSvc corestore.KVStoreService,
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
			codec.CollValue[channeltypes.Packet](cdc),
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
// TODO: modify the escrow account to be the known signer address
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

	// TODO: implement?
	nonrefundable := false

	// for forwarded packets, the funds were moved into an escrow account if the denom originated on this chain.
	// On an ack error or timeout on a forwarded packet, the funds in the escrow account
	// should be moved to the other escrow account on the other side or burned.
	if !ack.Success() {
		// If this packet is non-refundable due to some action that took place between the initial ibc transfer and the forward
		// we write a successful ack containing details on what happened regardless of ack error or timeout
		if nonrefundable {
			// we are not allowed to refund back to the source chain.
			// attempt to move funds to user recoverable account on this chain.
			// TODO: re-ADD
			// if err := k.moveFundsToUserRecoverableAccount(ctx, packet, data, inFlightPacket); err != nil {
			// 	return err
			// }

			ackResult := fmt.Sprintf("packet forward failed after point of no return: %s", ack.GetError())
			newAck := channeltypes.NewResultAcknowledgement([]byte(ackResult))

			return k.WriteIBCAcknowledgment(ctx, chanCap, channeltypes.Packet{
				Data:               packet.Data,
				Sequence:           packet.Sequence,
				SourcePort:         packet.SourcePort,
				SourceChannel:      packet.SourceChannel,
				DestinationPort:    packet.DestinationPort,
				DestinationChannel: packet.DestinationChannel,
				TimeoutHeight:      packet.TimeoutHeight,
				TimeoutTimestamp:   packet.TimeoutTimestamp,
			}, newAck)
		}

		//fullDenomPath := data.Denom
		fullDenomPath := getDenomForThisChain(
			packet.DestinationPort, packet.DestinationChannel,
			packet.SourcePort, packet.SourceChannel,
			data.Denom,
		)

		var err error

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
		fmt.Println("denomTrace", denomTrace, "fullDenomPath", fullDenomPath, "denomTrace.IBCDenom()", denomTrace.IBCDenom())
		coin := sdk.NewCoin(denomTrace.IBCDenom(), amount)

		// TODO: @facu during callbacks the escrow address is not the gateway contract address, it's the isolated address
		// let's make sure we send the funds back to the correct address

		// escrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)
		refundEscrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)

		// Override the receiver address to the gateway contract address
		gatewayAddr := common.HexToAddress("0x5A6A8Ce46E34c2cd998129d013fA0253d3892345") // TODO: make this configurable
		escrowAddress := sdk.AccAddress(gatewayAddr.Bytes())

		newToken := sdk.NewCoins(coin)

		k.Logger(ctx).Info("Escrow address!!!!", "escrowAddress", escrowAddress.String(), "coins", newToken)

		// Sender chain is source
		if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
			// funds were moved to escrow account for transfer, so they need to either:
			// - move to the other escrow account, in the case of native denom
			// - burn
			if transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
				// transfer funds from escrow account for forwarded packet to escrow account going back for refund.
				balances := k.BankKeeper.GetAllBalances(ctx, escrowAddress)
				k.Logger(ctx).Info("Balances of escrow!", "balances", balances)
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

// moveFundsToUserRecoverableAccount will move the funds from the escrow account to the user recoverable account
// this is only used when the maximum timeouts have been reached or there is an acknowledgement error and the packet is nonrefundable,
// i.e. an operation has occurred to make the original packet funds inaccessible to the user, e.g. a swap.
// We cannot refund the funds back to the original chain, so we move them to an account on this chain that the user can access.
func (k Keeper) moveFundsToUserRecoverableAccount(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	inFlightPacket *types.InFlightPacket,
) error {
	fullDenomPath := data.Denom

	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("failed to parse amount from packet data for forward recovery: %s", data.Amount)
	}
	denomTrace := transfertypes.ParseDenomTrace(fullDenomPath)
	token := sdk.NewCoin(denomTrace.IBCDenom(), amount)

	userAccount, err := userRecoverableAccount(inFlightPacket)
	if err != nil {
		return fmt.Errorf("failed to get user recoverable account: %w", err)
	}

	if !transfertypes.SenderChainIsSource(packet.SourcePort, packet.SourceChannel, fullDenomPath) {
		// mint vouchers back to sender
		if err := k.BankKeeper.MintCoins(
			ctx, transfertypes.ModuleName, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, transfertypes.ModuleName, userAccount, sdk.NewCoins(token)); err != nil {
			panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
		}
		return nil
	}

	escrowAddress := transfertypes.GetEscrowAddress(packet.SourcePort, packet.SourceChannel)

	if err := k.BankKeeper.SendCoins(
		ctx, escrowAddress, userAccount, sdk.NewCoins(token),
	); err != nil {
		return fmt.Errorf("failed to send coins from escrow account to user recoverable account: %w", err)
	}

	// update the total escrow amount for the denom.
	k.unescrowToken(ctx, token)

	return nil
}

// userRecoverableAccount finds an account on this chain that the original sender of the packet can recover funds from.
// If the destination receiver of the original packet is a valid bech32 address for this chain, we use that address.
// Otherwise, if the sender of the original packet is a valid bech32 address for another chain, we translate that address to this chain.
// Note that for the fallback, the coin type of the source chain sender account must be compatible with this chain.
func userRecoverableAccount(inFlightPacket *types.InFlightPacket) (sdk.AccAddress, error) {
	var originalData transfertypes.FungibleTokenPacketData
	err := transfertypes.ModuleCdc.UnmarshalJSON(inFlightPacket.PacketData, &originalData)
	if err == nil {
		sender, err := sdk.AccAddressFromBech32(originalData.Receiver)
		if err == nil {
			return sender, nil
		}
	}

	_, sender, fallbackErr := bech32.DecodeAndConvert(inFlightPacket.OriginalSenderAddress)
	if fallbackErr == nil {
		return sender, nil
	}

	return nil, fmt.Errorf("failed to decode bech32 addresses: %w", errors.Join(err, fallbackErr))
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
