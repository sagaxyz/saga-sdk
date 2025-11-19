package gateway_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

// MockAuthzKeeper is a mock implementation of authz keeper
type MockAuthzKeeper struct {
	mock.Mock
}

// MockEVMKeeper is a mock implementation of EVM keeper for gateway
type MockEVMKeeper struct {
	mock.Mock
}

func (m *MockEVMKeeper) CallEVMWithData(
	ctx sdk.Context,
	from common.Address,
	contract *common.Address,
	data []byte,
	commit bool,
) (*vmtypes.MsgEthereumTxResponse, error) {
	args := m.Called(ctx, from, contract, data, commit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*vmtypes.MsgEthereumTxResponse), args.Error(1)
}

func (m *MockEVMKeeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from, contract common.Address,
	commit bool,
	method string,
	args ...interface{},
) (*vmtypes.MsgEthereumTxResponse, error) {
	callArgs := m.Called(ctx, abi, from, contract, commit, method, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*vmtypes.MsgEthereumTxResponse), callArgs.Error(1)
}

// ApplyMessage is not used in current tests - commented out due to API changes
// func (m *MockEVMKeeper) ApplyMessage(
// 	ctx sdk.Context,
// 	msg ethtypes.Message,
// 	tracer interface{},
// 	commit bool,
// ) (*vmtypes.MsgEthereumTxResponse, error) {
// 	args := m.Called(ctx, msg, tracer, commit)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*vmtypes.MsgEthereumTxResponse), args.Error(1)
// }

func (m *MockEVMKeeper) GetAccountOrEmpty(ctx sdk.Context, addr common.Address) interface{} {
	args := m.Called(ctx, addr)
	return args.Get(0)
}

// MockPacketDataUnmarshaler is a mock implementation of packet data unmarshaler
type MockPacketDataUnmarshaler struct {
	mock.Mock
}

func (m *MockPacketDataUnmarshaler) UnmarshalPacketData(bz []byte) (interface{}, error) {
	args := m.Called(bz)
	return args.Get(0), args.Error(1)
}

// MockAccount is a mock EVM account
type MockAccount struct {
	isContract bool
}

func (m *MockAccount) IsContract() bool {
	return m.isContract
}
