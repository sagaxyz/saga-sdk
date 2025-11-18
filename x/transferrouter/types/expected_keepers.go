package types

import (
	context "context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
}

type TransferKeeper interface {
	UnescrowCoin(ctx sdk.Context, escrowAddress sdk.AccAddress, sender sdk.AccAddress, coin sdk.Coin) error
}

type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

type EVMKeeper interface {
	EnableStaticPrecompiles(ctx sdk.Context, addresses ...common.Address) error
	IsContract(ctx sdk.Context, address common.Address) bool
}

type ERC20Keeper interface {
	GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error)
	GetTokenPairID(ctx sdk.Context, token string) []byte
	GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool)
	BalanceOf(ctx sdk.Context, abi abi.ABI, contract, account common.Address) *big.Int
}

type AccountKeeper interface {
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, account sdk.AccountI)
	GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string)
}
