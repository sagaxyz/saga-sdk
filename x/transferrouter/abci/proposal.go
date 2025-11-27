package abci

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	precompilesgateway "github.com/sagaxyz/saga-sdk/x/transferrouter/precompiles/gateway"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
)

type ProposalHandler struct {
	keeper     keeper.Keeper
	txSelector baseapp.TxSelector
	signer     ethtypes.Signer
	txVerifier baseapp.ProposalTxVerifier
	txConfig   client.TxConfig
}

type ProposalHandlerOptions struct {
	Keeper     keeper.Keeper
	TxSelector baseapp.TxSelector
	Signer     ethtypes.Signer
	TxVerifier baseapp.ProposalTxVerifier
	TxConfig   client.TxConfig
}

func NewProposalHandler(opts ProposalHandlerOptions) *ProposalHandler {
	return &ProposalHandler{
		keeper:     opts.Keeper,
		txSelector: opts.TxSelector,
		signer:     opts.Signer,
		txVerifier: opts.TxVerifier,
		txConfig:   opts.TxConfig,
	}
}

var CallMaxGas = uint64(1000000) // arbitrary value

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		logger := h.keeper.Logger(ctx)

		// 1. Add the source callback queue
		chainId, err := utils.ParseChainID(ctx.ChainID())
		if err != nil {
			logger.Error("failed to parse chain id", "error", err)
			return nil, errors.New("failed to parse chain id")
		}

		var maxBlockGas uint64
		consParams := ctx.ConsensusParams()
		if consParams.Block != nil {
			maxBlockGas = uint64(consParams.Block.MaxGas)
		}

		if h.txSelector == nil {
			logger.Error("tx selector is nil")
			return nil, errors.New("tx selector is nil")
		}

		defer h.txSelector.Clear()

		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			logger.Error("Failed to get params", "error", err)
			return nil, errors.New("failed to get params")
		}

		// Parse the configured private key (in hex format) and derive the corresponding
		// Ethereum address of the known signer.
		if params.KnownSignerPrivateKey == "" {
			logger.Error("known signer private key is empty")
			return nil, errors.New("known signer private key is empty")
		}
		privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
		if err != nil {
			logger.Error("failed to parse private key", "error", err)
			return nil, errors.New("failed to parse private key")
		}

		knownSignerBz := crypto.PubkeyToAddress(privKey.PublicKey).Bytes()
		nextNonce, err := h.keeper.AccountKeeper.GetSequence(ctx, sdk.AccAddress(knownSignerBz))
		if err != nil {
			return nil, err
		}

		// Add all the transactions in our internal queues
		nextNonce, err = h.AddSrcCallbackTxs(ctx, req, nextNonce, chainId, privKey, maxBlockGas)
		if err != nil {
			logger.Error("Error during src callback queue walk", "error", err)
			return nil, err
		}

		err = h.AddPacketTxs(ctx, req, nextNonce, chainId, privKey, maxBlockGas)
		if err != nil {
			logger.Error("Error during packet queue walk", "error", err)
			return nil, err
		}

		err = h.AddErrorOrTimeoutTxs(ctx, req, nextNonce, chainId, privKey, maxBlockGas)
		if err != nil {
			logger.Error("Error during error or timeout queue walk", "error", err)
			return nil, err
		}

		// 2. Add the rest of the transactions in the incoming request
		if h.txVerifier == nil {
			logger.Error("tx verifier is nil")
			return nil, errors.New("tx verifier is nil")
		}

		err = h.AddIncomingTxs(ctx, req, maxBlockGas, knownSignerBz)
		if err != nil {
			logger.Error("Error while adding incoming txs", "error", err)
			return nil, err
		}

		selectedTxs := h.txSelector.SelectedTxs(ctx)

		return &abci.ResponsePrepareProposal{
			Txs: selectedTxs,
		}, nil
	}
}

// ProcessProposalHandler has no checks, it just accepts the block. This is due to the fact that the injected message
// can't be manipulated by the proposer, as the actual calldata is get during execution.
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {

		// reject any block that contains repeated txs from the known signer // TODO: implement this
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}

// AddSrcCallbackTxs adds the source callback transactions to the proposal
func (h *ProposalHandler) AddSrcCallbackTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, nextNonce uint64, chainId *big.Int, privKey *ecdsa.PrivateKey, maxBlockGas uint64) (uint64, error) {
	logger := h.keeper.Logger(ctx)
	// Add the source callback queue
	err := h.keeper.SrcCallbackQueue.Walk(ctx, nil, func(key uint64, _ types.PacketQueueItem) (stop bool, err error) {
		// Calldata is a simple call to the gateway executeSrcCallback function
		calldata, err := precompilesgateway.ABI.Pack("executeSrcCallback")
		if err != nil {
			logger.Error("failed to pack calldata", "error", err)
			return true, err
		}

		cosmosTx, txBytes, err := h.calldataToSignedTx(ctx, calldata, nextNonce, chainId, privKey)
		if err != nil {
			logger.Error("failed to convert calldata to signed tx", "error", err)
			return true, err
		}

		stop = h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, cosmosTx, txBytes)
		if stop {
			logger.Error("tx selector stopped")
			return true, nil
		}

		nextNonce = nextNonce + 1
		return false, nil
	})

	return nextNonce, err
}

// AddErrorOrTimeoutTxs adds the error or timeout transactions to the proposal
func (h *ProposalHandler) AddErrorOrTimeoutTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, nextNonce uint64, chainId *big.Int, privKey *ecdsa.PrivateKey, maxBlockGas uint64) error {
	logger := h.keeper.Logger(ctx)
	return h.keeper.ErrorOrTimeoutQueue.Walk(ctx, nil, func(_ uint64, _ types.PacketQueueItem) (stop bool, err error) {
		// Calldata is a simple call to the gateway handleErrorOrTimeout function
		calldata, err := precompilesgateway.ABI.Pack("handleErrorOrTimeout")
		if err != nil {
			logger.Error("failed to pack calldata", "error", err)
			return true, err
		}

		cosmosTx, txBytes, err := h.calldataToSignedTx(ctx, calldata, nextNonce, chainId, privKey)
		if err != nil {
			logger.Error("failed to convert calldata to signed tx", "error", err)
			return true, err
		}

		stop = h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, cosmosTx, txBytes)
		if stop {
			return true, nil
		}

		nextNonce = nextNonce + 1
		return false, nil
	})
}

// AddPacketTxs adds the packet transactions to the proposal
func (h *ProposalHandler) AddPacketTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, nextNonce uint64, chainId *big.Int, privKey *ecdsa.PrivateKey, maxBlockGas uint64) error {
	logger := h.keeper.Logger(ctx)
	err := h.keeper.PacketQueue.Walk(ctx, nil, func(_ uint64, _ types.PacketQueueItem) (stop bool, err error) {
		// Calldata is a simple call to the gateway execute function
		calldata, err := precompilesgateway.ABI.Pack("execute")
		if err != nil {
			logger.Error("failed to pack calldata", "error", err)
			return true, err
		}

		cosmosTx, txBytes, err := h.calldataToSignedTx(ctx, calldata, nextNonce, chainId, privKey)
		if err != nil {
			logger.Error("failed to convert calldata to signed tx", "error", err)
			return true, err
		}

		stop = h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, cosmosTx, txBytes)
		// If the transaction is not added, we stop the walk, because we don't want to execute queued calls out of order
		if stop {
			return true, nil
		}

		nextNonce = nextNonce + 1
		return false, nil
	})

	return err
}

func (h *ProposalHandler) AddIncomingTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, maxBlockGas uint64, knownSignerBz []byte) error {
	for _, txBz := range req.Txs {
		if txBz == nil {
			continue
		}

		tx, err := h.txVerifier.TxDecode(txBz)
		if err != nil {
			return err
		}

		if tx == nil {
			continue
		}

		// Reject any txs from the known signer
		skip := false
		for _, msg := range tx.GetMsgs() {
			ethTx, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}
			from := common.BytesToAddress(ethTx.From)

			if bytes.Equal(from.Bytes(), knownSignerBz) {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
		if stop {
			break
		}
	}

	return nil
}

func (h *ProposalHandler) calldataToSignedTx(ctx sdk.Context, calldata []byte, nonce uint64, chainID *big.Int, privKey *ecdsa.PrivateKey) (sdk.Tx, []byte, error) {
	logger := h.keeper.Logger(ctx)
	txArgs := &evmtypes.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  CallMaxGas,
		Input:     calldata,
		GasFeeCap: big.NewInt(0),
		GasPrice:  big.NewInt(0),
		ChainID:   chainID,
		Amount:    big.NewInt(0), // No value transfer for contract calls
		GasTipCap: big.NewInt(0),
		To:        &precompilesgateway.PrecompileAddress,
		Accesses:  nil, // No access list for now
	}

	ethtx := txArgs.ToTx()

	if h.signer == nil {
		logger.Error("signer is nil")
		return nil, nil, errors.New("signer is nil")
	}

	ethtx.ChainId().Set(chainID)

	if ethtx == nil {
		logger.Error("as transaction returned nil")
		return nil, nil, errors.New("as transaction returned nil")
	}

	signedTx, err := ethcoretypes.SignTx(ethtx, h.signer, privKey)
	if err != nil {
		logger.Error("sign tx failed", "error", err)
		return nil, nil, err
	}

	tx := &evmtypes.MsgEthereumTx{}
	err = tx.FromSignedEthereumTx(signedTx, h.signer)
	if err != nil {
		logger.Error("from signed ethereum tx failed", "error", err)
		return nil, nil, err
	}

	if err := tx.ValidateBasic(); err != nil {
		logger.Error("validate basic failed", "error", err)
		return nil, nil, err
	}

	cosmosTx, err := tx.BuildTx(h.txConfig.NewTxBuilder(), "saga") // TODO: get denom from params
	if err != nil {
		logger.Error("build tx failed", "error", err)
		return nil, nil, err
	}

	// Encode transaction by default Tx encoder
	txBytes, err := h.txConfig.TxEncoder()(cosmosTx)
	if err != nil {
		logger.Error("tx encoder failed", "error", err)
		return nil, nil, err
	}

	return cosmosTx, txBytes, nil
}
