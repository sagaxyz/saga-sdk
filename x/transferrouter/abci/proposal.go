package abci

import (
	"errors"
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/ethereum/go-ethereum/common"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	evmostypes "github.com/evmos/evmos/v20/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/abi"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

type ProposalHandler struct {
	keeper     keeper.Keeper
	txSelector baseapp.TxSelector
	signer     ethtypes.Signer
	txVerifier baseapp.ProposalTxVerifier
	txConfig   client.TxConfig
}

func NewProposalHandler(keeper keeper.Keeper, txSelector baseapp.TxSelector, signer ethtypes.Signer, txVerifier baseapp.ProposalTxVerifier, txConfig client.TxConfig) *ProposalHandler {
	return &ProposalHandler{
		keeper:     keeper,
		txSelector: txSelector,
		signer:     signer,
		txVerifier: txVerifier,
		txConfig:   txConfig,
	}
}

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		chainId, err := evmostypes.ParseChainID(ctx.ChainID())
		if err != nil {
			h.keeper.Logger(ctx).Error("failed to parse chain id", "error", err)
			return nil, errors.New("failed to parse chain id")
		}

		logger := h.keeper.Logger(ctx)
		logger.Info("PrepareProposalHandler start", "height", ctx.BlockHeight(), "txs_in_request", len(req.Txs), "maxTxBytes", req.MaxTxBytes)

		// 1. Add the calls to the proposal
		var maxBlockGas uint64
		consParams := ctx.ConsensusParams()
		if consParams.Block != nil {
			maxBlockGas = uint64(consParams.Block.MaxGas)
			logger.Info("Block max gas loaded", "maxBlockGas", maxBlockGas)
		} else {
			logger.Error("Consensus params Block is nil")
		}

		defer h.txSelector.Clear()

		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			logger.Error("failed to get params", "error", err)
			return nil, errors.New("failed to get params")
		}
		logger.Info("Params retrieved", "knownSignerPrivateKey_len", len(params.KnownSignerPrivateKey))

		// Parse the configured private key (in hex format) and derive the corresponding
		// Ethereum address of the known signer.
		if params.KnownSignerPrivateKey == "" {
			logger.Error("KnownSignerPrivateKey is empty")
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
			nextNonce = 0
			logger.Error("failed to get sequence", "error", err)
		}

		// TODO: possible issue here, if there are many IBC txs being sent in, they might block
		// other normal txs. We should add a % limit of space IBC txs can take in the proposal.
		err = h.keeper.PacketQueue.Walk(ctx, nil, func(key uint64, value channeltypes.Packet) (stop bool, err error) {
			logger.Info("Processing call queue item", "key", key, "value", value)

			// Calldata is a simple call to the gateway execute function with the sequence
			calldata, err := abi.GatewayABI.Pack("execute", big.NewInt(int64(key)))
			if err != nil {
				logger.Error("Failed to pack calldata", "error", err, "key", key)
				return true, err
			}

			// TODO: tmp for tests
			logger.Info("GatewayContractAddress", "GatewayContractAddress", params.GatewayContractAddress)
			if params.GatewayContractAddress == "" {
				logger.Error("GatewayContractAddress is empty")
				params.GatewayContractAddress = "0x5A6A8Ce46E34c2cd998129d013fA0253d3892345"
			}
			gatewayAddress := common.HexToAddress(params.GatewayContractAddress)

			logger.Info("About to call ToMsgEthereumTx", "key", key, "nonce", nextNonce)
			msgEthTx := calldataToMsgEthereumTx(nextNonce, chainId, &gatewayAddress, calldata)

			logger.Info("ToMsgEthereumTx completed", "key", key, "ethTx", msgEthTx)
			if msgEthTx == nil {
				logger.Error("ToMsgEthereumTx returned nil", "key", key)
				return true, errors.New("failed to convert to ethereum tx")
			}

			if h.signer == nil {
				logger.Error("Signer is nil")
				return true, errors.New("signer is nil")
			}

			logger.Info("About to call AsTransaction", "key", key)
			ethtx := msgEthTx.AsTransaction()
			ethtx.ChainId().Set(chainId)

			logger.Info("AsTransaction completed", "key", key, "ethtx", ethtx)
			if ethtx == nil {
				logger.Error("AsTransaction returned nil", "key", key)
				return true, errors.New("as transaction returned nil")
			}

			logger.Info("About to sign transaction", "key", key)
			signedTx, err := ethcoretypes.SignTx(ethtx, h.signer, privKey)
			if err != nil {
				logger.Error("Failed to sign transaction", "error", err, "key", key)
				return true, err
			}
			logger.Info("Transaction signed successfully", "key", key)

			// TODO: might not be the right way to do it, let's circle back later
			logger.Info("About to create MsgEthereumTx", "key", key)
			msgEthTx = &evmtypes.MsgEthereumTx{}
			logger.Info("About to call FromEthereumTx", "key", key)
			err = msgEthTx.FromEthereumTx(signedTx)
			if err != nil {
				logger.Error("Failed to convert from ethereum tx", "error", err, "key", key)
				return true, err
			}
			logger.Info("FromEthereumTx completed", "key", key, "hash", msgEthTx.Hash)

			if err := msgEthTx.ValidateBasic(); err != nil {
				logger.Error("tx failed basic validation", "error", err.Error(), "key", key)
				return true, err
			}

			logger.Error("About to build cosmos tx", "contract address", signedTx.To().Hex())

			cosmosTx, err := msgEthTx.BuildTx(h.txConfig.NewTxBuilder(), "saga") //"res.Params.EvmDenom")
			if err != nil {
				logger.Error("failed to build cosmos tx", "error", err.Error(), "key", key)
				return true, err
			}

			// Encode transaction by default Tx encoder
			txBytes, err := h.txConfig.TxEncoder()(cosmosTx)
			if err != nil {
				logger.Error("failed to encode eth tx using default encoder", "error", err.Error(), "key", key)
				return true, err
			}

			if h.txSelector == nil {
				logger.Error("TxSelector is nil")
				return true, errors.New("tx selector is nil")
			}

			stop = h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, cosmosTx, txBytes)
			// If the transaction is not added, we stop the walk, because we don't want to execute queued calls out of order
			if stop {
				logger.Info("No more txs to add 1")
				return true, nil
			}

			logger.Info("Transaction added to proposal", "key", key)
			nextNonce = nextNonce + 1
			return false, nil
		})

		if err != nil {
			logger.Error("Error during call queue walk", "error", err)
			return nil, err
		}

		// 2. Add the rest of the transactions in the incoming request
		if h.txVerifier == nil {
			logger.Error("TxVerifier is nil")
			return nil, errors.New("tx verifier is nil")
		}

		for i, txBz := range req.Txs {
			logger.Info("Processing incoming transaction", "index", i, "txBz_len", len(txBz))

			if txBz == nil {
				logger.Error("Transaction bytes is nil", "index", i)
				continue
			}

			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {
				logger.Error("Failed to decode transaction", "error", err, "index", i)
				return nil, err
			}

			if tx == nil {
				logger.Error("Decoded transaction is nil", "index", i)
				continue
			}

			stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
			if stop {
				break
			}
			logger.Info("Transaction added to proposal", "index", i)
		}

		if h.txSelector == nil {
			logger.Error("TxSelector is nil when getting selected txs")
			return nil, errors.New("tx selector is nil")
		}

		selectedTxs := h.txSelector.SelectedTxs(ctx)

		if selectedTxs == nil {
			selectedTxs = [][]byte{} // Return empty slice instead of nil
		}

		logger.Info("PrepareProposalHandler completed", "selected_txs_count", len(selectedTxs))
		return &abci.ResponsePrepareProposal{
			Txs: selectedTxs,
		}, nil
	}
}

// ProcessProposalHandler checks if the transaction added by the proposer is derived from a call in the queue.
// This is to prevent the proposer from adding arbitrary transactions, which is a security risk, and could be considered
// malicious behavior. TODO: add a slashing mechanism for this (might be difficult as this is outside the state machine).
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}

func calldataToMsgEthereumTx(nonce uint64, chainID *big.Int, contract *common.Address, callData []byte) *evmtypes.MsgEthereumTx {
	txArgs := &evmtypes.EvmTxArgs{
		Nonce:     nonce,    // Will be set by the signer
		GasLimit:  16100000, // Standard gas limit for simple transfers // TODO: figure out how to set this
		Input:     callData,
		GasFeeCap: big.NewInt(0), // Will be set by the signer
		GasPrice:  big.NewInt(0), // Will be set by the signer
		ChainID:   chainID,       // Default chain ID, should be configurable
		Amount:    big.NewInt(0), // No value transfer for contract calls
		GasTipCap: big.NewInt(0), // Will be set by the signer
		To:        contract,
		Accesses:  nil, // No access list for now
	}

	tx := evmtypes.NewTx(txArgs)
	return tx
}
