package abci_test

import (
	"errors"
	"math/big"
	"testing"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/sagaxyz/saga-sdk/x/transferrouter"
	transferrouter_abci "github.com/sagaxyz/saga-sdk/x/transferrouter/abci"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

const (
	validPrivateKey     = "f6dba52e479cf5d7ad58bc11177c105ac7b89a02be1d432e77e113fc53377978"
	validGatewayAddress = "0x5A6A8Ce46E34c2cd998129d013fA0253d3892345"
	testChainID         = "saga_12345-1"
	testChainIDNumeric  = int64(12345)
	testAuthority       = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	defaultMaxTxBytes   = int64(1000000)
	defaultMaxBlockGas  = int64(10000000)
)

type ProposalHandlerTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	keeper         keeper.Keeper
	handler        *transferrouter_abci.ProposalHandler
	mockTxSelector *MockTxSelector
	mockTxVerifier *MockTxVerifier
	mockTxConfig   *MockTxConfig
	signer         ethtypes.Signer
}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

func (s *ProposalHandlerTestSuite) SetupTest() {
	// Setup store and context
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{types.StoreKey: key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	ctx = ctx.WithChainID(testChainID)
	ctx = ctx.WithConsensusParams(tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxGas: defaultMaxBlockGas,
		},
	})
	s.ctx = ctx

	// Setup encoding
	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	cdc := enc.Codec

	// Setup mock account keeper
	mockAccountKeeper := &MockAccountKeeper{}
	mockAccountKeeper.On("GetSequence", mock.Anything, mock.Anything).Return(uint64(0), nil)

	// Create keeper
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(key),
		nil, // erc20 keeper
		nil, // ics4 wrapper
		nil, // channel keeper
		nil, // transfer keeper
		nil, // bank keeper
		mockAccountKeeper,
		nil, // evm keeper
		testAuthority,
	)

	// Set valid params
	err := k.Params.Set(ctx, types.Params{
		Enabled:                true,
		KnownSignerPrivateKey:  validPrivateKey,
		GatewayContractAddress: validGatewayAddress,
	})
	s.Require().NoError(err)

	s.keeper = k

	// Setup mocks
	s.mockTxSelector = new(MockTxSelector)
	s.mockTxVerifier = new(MockTxVerifier)
	s.mockTxConfig = new(MockTxConfig)
	s.signer = ethtypes.LatestSignerForChainID(big.NewInt(testChainIDNumeric))

	// Create handler
	s.handler = transferrouter_abci.NewProposalHandler(transferrouter_abci.ProposalHandlerOptions{
		Keeper:     k,
		TxSelector: s.mockTxSelector,
		Signer:     s.signer,
		TxVerifier: s.mockTxVerifier,
		TxConfig:   s.mockTxConfig,
	})
}

func (s *ProposalHandlerTestSuite) TearDownTest() {
	// Verify all mock expectations were met
	s.mockTxSelector.AssertExpectations(s.T())
	s.mockTxVerifier.AssertExpectations(s.T())
	s.mockTxConfig.AssertExpectations(s.T())
}

// TestPrepareProposal_EmptyQueues tests the basic happy path with no packets in queue
func (s *ProposalHandlerTestSuite) TestPrepareProposal_EmptyQueues() {
	s.mockTxSelector.On("Clear").Return()
	s.mockTxSelector.On("SelectedTxs", s.ctx).Return([][]byte{})

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Empty(resp.Txs)
}

// TestPrepareProposal_WithIncomingTxs tests adding incoming txs from mempool
func (s *ProposalHandlerTestSuite) TestPrepareProposal_WithIncomingTxs() {
	incomingTx := []byte("incoming-tx")
	mockTx := &evmtypes.MsgEthereumTx{}

	s.mockTxSelector.On("Clear").Return()
	s.mockTxVerifier.On("TxDecode", incomingTx).Return(mockTx, nil)
	s.mockTxSelector.On("SelectTxForProposal", s.ctx, uint64(defaultMaxTxBytes), uint64(defaultMaxBlockGas), mockTx, incomingTx).Return(false)
	s.mockTxSelector.On("SelectedTxs", s.ctx).Return([][]byte{incomingTx})

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{incomingTx},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Txs, 1)
}

// TestPrepareProposal_InvalidChainID tests error handling for invalid chain ID
func (s *ProposalHandlerTestSuite) TestPrepareProposal_InvalidChainID() {
	ctx := s.ctx.WithChainID("invalid-chain-id")

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(ctx, req)
	s.Require().Error(err)
	s.Require().Nil(resp)
	s.Require().Contains(err.Error(), "failed to parse chain id")
}

// TestPrepareProposal_MissingParams tests error handling when params are not set
func (s *ProposalHandlerTestSuite) TestPrepareProposal_MissingParams() {
	// Create new keeper without params
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{types.StoreKey: key},
		map[string]*storetypes.TransientStoreKey{},
		nil,
	)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	ctx = ctx.WithChainID(testChainID)

	enc := moduletestutil.MakeTestEncodingConfig(transferrouter.AppModuleBasic{})
	k := keeper.NewKeeper(
		enc.Codec,
		runtime.NewKVStoreService(key),
		nil, nil, nil, nil, nil, nil, nil,
		testAuthority,
	)

	mockTxSelector := new(MockTxSelector)
	mockTxSelector.On("Clear").Return()

	handler := transferrouter_abci.NewProposalHandler(transferrouter_abci.ProposalHandlerOptions{
		Keeper:     k,
		TxSelector: mockTxSelector,
		Signer:     s.signer,
		TxVerifier: nil,
		TxConfig:   nil,
	})

	prepareHandler := handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(ctx, req)
	s.Require().Error(err)
	s.Require().Nil(resp)
	s.Require().Contains(err.Error(), "failed to get params")

	mockTxSelector.AssertExpectations(s.T())
}

// TestPrepareProposal_EmptyPrivateKey tests error handling for empty private key
func (s *ProposalHandlerTestSuite) TestPrepareProposal_EmptyPrivateKey() {
	err := s.keeper.Params.Set(s.ctx, types.Params{
		Enabled:                true,
		KnownSignerPrivateKey:  "",
		GatewayContractAddress: validGatewayAddress,
	})
	s.Require().NoError(err)

	s.mockTxSelector.On("Clear").Return()

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(resp)
	s.Require().Contains(err.Error(), "known signer private key is empty")
}

// TestPrepareProposal_InvalidPrivateKey tests error handling for malformed private key
func (s *ProposalHandlerTestSuite) TestPrepareProposal_InvalidPrivateKey() {
	err := s.keeper.Params.Set(s.ctx, types.Params{
		Enabled:                true,
		KnownSignerPrivateKey:  "not-a-valid-hex-key",
		GatewayContractAddress: validGatewayAddress,
	})
	s.Require().NoError(err)

	s.mockTxSelector.On("Clear").Return()

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(resp)
	s.Require().Contains(err.Error(), "failed to parse private key")
}

// TestPrepareProposal_NilIncomingTx tests handling of nil transaction in incoming txs
func (s *ProposalHandlerTestSuite) TestPrepareProposal_NilIncomingTx() {
	s.mockTxSelector.On("Clear").Return()
	s.mockTxSelector.On("SelectedTxs", s.ctx).Return([][]byte{})

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{nil, []byte("valid-tx"), nil},
		MaxTxBytes: defaultMaxTxBytes,
	}

	// Should skip nil txs without error
	mockTx := &evmtypes.MsgEthereumTx{}
	s.mockTxVerifier.On("TxDecode", []byte("valid-tx")).Return(mockTx, nil)
	s.mockTxSelector.On("SelectTxForProposal", s.ctx, uint64(defaultMaxTxBytes), uint64(defaultMaxBlockGas), mockTx, []byte("valid-tx")).Return(false)

	resp, err := prepareHandler(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}

// TestPrepareProposal_IncomingTxStopsSelection tests that incoming tx can stop selection
func (s *ProposalHandlerTestSuite) TestPrepareProposal_IncomingTxStopsSelection() {
	txs := [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3")}
	mockTx1 := &evmtypes.MsgEthereumTx{}
	mockTx2 := &evmtypes.MsgEthereumTx{}

	s.mockTxSelector.On("Clear").Return()
	s.mockTxVerifier.On("TxDecode", txs[0]).Return(mockTx1, nil)
	s.mockTxSelector.On("SelectTxForProposal", s.ctx, uint64(defaultMaxTxBytes), uint64(defaultMaxBlockGas), mockTx1, txs[0]).Return(false)
	s.mockTxVerifier.On("TxDecode", txs[1]).Return(mockTx2, nil)
	s.mockTxSelector.On("SelectTxForProposal", s.ctx, uint64(defaultMaxTxBytes), uint64(defaultMaxBlockGas), mockTx2, txs[1]).Return(true) // Stop here
	s.mockTxSelector.On("SelectedTxs", s.ctx).Return([][]byte{txs[0], txs[1]})

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        txs,
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Txs, 2)
}

// TestPrepareProposal_TxDecodeFails tests handling of tx decode errors
func (s *ProposalHandlerTestSuite) TestPrepareProposal_TxDecodeFails() {
	incomingTx := []byte("malformed-tx")

	s.mockTxSelector.On("Clear").Return()
	s.mockTxVerifier.On("TxDecode", incomingTx).Return(nil, errors.New("decode error"))

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{incomingTx},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().Error(err)
	s.Require().Nil(resp)
}

// TestPrepareProposal_TxDecodeReturnsNil tests handling when decoder returns nil tx
func (s *ProposalHandlerTestSuite) TestPrepareProposal_TxDecodeReturnsNil() {
	incomingTx := []byte("nil-tx")

	s.mockTxSelector.On("Clear").Return()
	s.mockTxVerifier.On("TxDecode", incomingTx).Return(nil, nil)
	s.mockTxSelector.On("SelectedTxs", s.ctx).Return([][]byte{})

	prepareHandler := s.handler.PrepareProposalHandler()
	req := &abci.RequestPrepareProposal{
		Txs:        [][]byte{incomingTx},
		MaxTxBytes: defaultMaxTxBytes,
	}

	resp, err := prepareHandler(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}

// TestProcessProposal_AlwaysAccepts tests that ProcessProposal always accepts
func (s *ProposalHandlerTestSuite) TestProcessProposal_AlwaysAccepts() {
	processHandler := s.handler.ProcessProposalHandler()

	testCases := []struct {
		name string
		txs  [][]byte
	}{
		{
			name: "empty txs",
			txs:  [][]byte{},
		},
		{
			name: "single tx",
			txs:  [][]byte{[]byte("tx1")},
		},
		{
			name: "multiple txs",
			txs:  [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3")},
		},
		{
			name: "with nil tx",
			txs:  [][]byte{[]byte("tx1"), nil, []byte("tx3")},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := &abci.RequestProcessProposal{
				Txs: tc.txs,
			}

			resp, err := processHandler(s.ctx, req)
			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Require().Equal(abci.ResponseProcessProposal_ACCEPT, resp.Status)
		})
	}
}

// TestNewProposalHandler tests the constructor
func (s *ProposalHandlerTestSuite) TestNewProposalHandler() {
	handler := transferrouter_abci.NewProposalHandler(
		transferrouter_abci.ProposalHandlerOptions{
			Keeper:     s.keeper,
			TxSelector: s.mockTxSelector,
			Signer:     s.signer,
			TxVerifier: s.mockTxVerifier,
			TxConfig:   s.mockTxConfig,
		},
	)

	s.Require().NotNil(handler)
	s.Require().NotNil(handler.PrepareProposalHandler())
	s.Require().NotNil(handler.ProcessProposalHandler())
}
