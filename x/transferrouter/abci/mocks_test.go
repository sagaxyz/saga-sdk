package abci_test

import (
	"context"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/mock"
)

// MockTxSelector is a mock implementation of baseapp.TxSelector
type MockTxSelector struct {
	mock.Mock
}

func (m *MockTxSelector) SelectedTxs(ctx context.Context) [][]byte {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([][]byte)
}

func (m *MockTxSelector) SelectTxForProposal(ctx context.Context, maxTxBytes, maxBlockGas uint64, tx sdk.Tx, txBz []byte) bool {
	args := m.Called(ctx, maxTxBytes, maxBlockGas, tx, txBz)
	return args.Bool(0)
}

func (m *MockTxSelector) Clear() {
	m.Called()
}

// MockTxVerifier is a mock implementation of baseapp.ProposalTxVerifier
type MockTxVerifier struct {
	mock.Mock
}

func (m *MockTxVerifier) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockTxVerifier) ProcessProposalVerifyTx(txBz []byte) (sdk.Tx, error) {
	args := m.Called(txBz)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sdk.Tx), args.Error(1)
}

func (m *MockTxVerifier) TxDecode(txBz []byte) (sdk.Tx, error) {
	args := m.Called(txBz)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(sdk.Tx), args.Error(1)
}

func (m *MockTxVerifier) TxEncode(tx sdk.Tx) ([]byte, error) {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockTxConfig is a mock implementation of client.TxConfig
type MockTxConfig struct {
	mock.Mock
}

func (m *MockTxConfig) NewTxBuilder() client.TxBuilder {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(client.TxBuilder)
}

func (m *MockTxConfig) TxEncoder() sdk.TxEncoder {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.TxEncoder)
}

func (m *MockTxConfig) TxDecoder() sdk.TxDecoder {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.TxDecoder)
}

func (m *MockTxConfig) TxJSONEncoder() sdk.TxEncoder {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.TxEncoder)
}

func (m *MockTxConfig) TxJSONDecoder() sdk.TxDecoder {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.TxDecoder)
}

func (m *MockTxConfig) SignModeHandler() *txsigning.HandlerMap {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*txsigning.HandlerMap)
}

func (m *MockTxConfig) MarshalSignatureJSON(sigs []signing.SignatureV2) ([]byte, error) {
	args := m.Called(sigs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockTxConfig) UnmarshalSignatureJSON(bz []byte) ([]signing.SignatureV2, error) {
	args := m.Called(bz)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]signing.SignatureV2), args.Error(1)
}

func (m *MockTxConfig) SigningContext() *txsigning.Context {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*txsigning.Context)
}

func (m *MockTxConfig) WrapTxBuilder(tx sdk.Tx) (client.TxBuilder, error) {
	args := m.Called(tx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.TxBuilder), args.Error(1)
}

// MockTxBuilder is a mock implementation of client.TxBuilder
type MockTxBuilder struct {
	mock.Mock
}

func (m *MockTxBuilder) GetTx() sdk.Tx {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.Tx)
}

func (m *MockTxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func (m *MockTxBuilder) SetSignatures(signatures ...interface{}) error {
	args := m.Called(signatures)
	return args.Error(0)
}

func (m *MockTxBuilder) SetMemo(memo string) {
	m.Called(memo)
}

func (m *MockTxBuilder) SetFeeAmount(amount sdk.Coins) {
	m.Called(amount)
}

func (m *MockTxBuilder) SetGasLimit(limit uint64) {
	m.Called(limit)
}

func (m *MockTxBuilder) SetTimeoutHeight(height uint64) {
	m.Called(height)
}

func (m *MockTxBuilder) SetFeeGranter(feeGranter sdk.AccAddress) {
	m.Called(feeGranter)
}

func (m *MockTxBuilder) SetFeePayer(feePayer sdk.AccAddress) {
	m.Called(feePayer)
}

func (m *MockTxBuilder) AddAuxSignerData(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

// MockAccountKeeper is a mock implementation of AccountKeeper
type MockAccountKeeper struct {
	mock.Mock
}

func (m *MockAccountKeeper) GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	args := m.Called(ctx, addr)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	args := m.Called(ctx, addr)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.AccountI)
}

func (m *MockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	args := m.Called(ctx, addr)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(sdk.AccountI)
}

func (m *MockAccountKeeper) SetAccount(ctx context.Context, account sdk.AccountI) {
	m.Called(ctx, account)
}

func (m *MockAccountKeeper) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
	args := m.Called(ctx, moduleName)
	var mod sdk.ModuleAccountI
	if args.Get(0) != nil {
		mod = args.Get(0).(sdk.ModuleAccountI)
	}
	if args.Get(1) == nil {
		return mod, nil
	}
	return mod, args.Get(1).([]string)
}
