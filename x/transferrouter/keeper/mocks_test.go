package keeper_test

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"

	erc20types "github.com/cosmos/evm/x/erc20/types"
)

// MockChannelKeeper is a mock implementation of ChannelKeeper
type MockChannelKeeper struct {
	mock.Mock
}

func (m *MockChannelKeeper) GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool) {
	args := m.Called(ctx, srcPort, srcChan)
	return args.Get(0).(channeltypes.Channel), args.Bool(1)
}

func (m *MockChannelKeeper) GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte {
	args := m.Called(ctx, portID, channelID, sequence)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]byte)
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	args := m.Called(ctx, portID, channelID)
	return args.Get(0).(uint64), args.Bool(1)
}

func (m *MockChannelKeeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	args := m.Called(ctx, portID, channelID)
	var cap *capabilitytypes.Capability
	if args.Get(1) != nil {
		cap = args.Get(1).(*capabilitytypes.Capability)
	}
	return args.String(0), cap, args.Error(2)
}

// MockTransferKeeper is a mock implementation of TransferKeeper
type MockTransferKeeper struct {
	mock.Mock
}

func (m *MockTransferKeeper) DenomPathFromHash(ctx sdk.Context, denomHash string) (string, error) {
	args := m.Called(ctx, denomHash)
	return args.String(0), args.Error(1)
}

func (m *MockTransferKeeper) GetTotalEscrowForDenom(ctx sdk.Context, denom string) sdk.Coin {
	args := m.Called(ctx, denom)
	return args.Get(0).(sdk.Coin)
}

func (m *MockTransferKeeper) SetTotalEscrowForDenom(ctx sdk.Context, coin sdk.Coin) {
	m.Called(ctx, coin)
}

// MockBankKeeper is a mock implementation of BankKeeper
type MockBankKeeper struct {
	mock.Mock
}

func (m *MockBankKeeper) SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	args := m.Called(ctx, fromAddr, toAddr, amt)
	return args.Error(0)
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	args := m.Called(ctx, senderAddr, recipientModule, amt)
	return args.Error(0)
}

func (m *MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	args := m.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	args := m.Called(ctx, moduleName, amt)
	return args.Error(0)
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	args := m.Called(ctx, senderModule, recipientAddr, amt)
	return args.Error(0)
}

// MockERC20Keeper is a mock implementation of ERC20Keeper
type MockERC20Keeper struct {
	mock.Mock
}

func (m *MockERC20Keeper) GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error) {
	args := m.Called(ctx, denom)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *MockERC20Keeper) GetTokenPairID(ctx sdk.Context, token string) []byte {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]byte)
}

func (m *MockERC20Keeper) GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool) {
	args := m.Called(ctx, id)
	return args.Get(0).(erc20types.TokenPair), args.Bool(1)
}

func (m *MockERC20Keeper) BalanceOf(ctx sdk.Context, abi abi.ABI, contract, account common.Address) *big.Int {
	args := m.Called(ctx, abi, contract, account)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*big.Int)
}

// MockAccountKeeper is a mock implementation of AccountKeeper
type MockAccountKeeper struct {
	mock.Mock
}

func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	args := m.Called(ctx, addr)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.AccountI)
}

func (m *MockAccountKeeper) GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	args := m.Called(ctx, addr)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	args := m.Called(ctx, addr)
	return args.Get(0).(sdk.AccountI)
}

func (m *MockAccountKeeper) SetAccount(ctx context.Context, account sdk.AccountI) {
	m.Called(ctx, account)
}

func (m *MockAccountKeeper) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
	args := m.Called(ctx, moduleName)
	var modAcc sdk.ModuleAccountI
	if args.Get(0) != nil {
		modAcc = args.Get(0).(sdk.ModuleAccountI)
	}
	return modAcc, args.Get(1).([]string)
}

// MockICS4Wrapper is a mock implementation of ICS4Wrapper
type MockICS4Wrapper struct {
	mock.Mock
}

func (m *MockICS4Wrapper) WriteAcknowledgement(ctx sdk.Context, packet ibcexported.PacketI, ack ibcexported.Acknowledgement) error {
	args := m.Called(ctx, packet, ack)
	return args.Error(0)
}

func (m *MockICS4Wrapper) SendPacket(ctx sdk.Context, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error) {
	args := m.Called(ctx, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	args := m.Called(ctx, portID, channelID)
	return args.String(0), args.Bool(1)
}
