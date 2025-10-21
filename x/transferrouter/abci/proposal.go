package abci

import (
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

var CallMaxGas = uint64(10000000) // arbitrary value

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		chainId, err := utils.ParseChainID(ctx.ChainID())
		if err != nil {
			h.keeper.Logger(ctx).Error("failed to parse chain id", "error", err)
			return nil, errors.New("failed to parse chain id")
		}

		logger := h.keeper.Logger(ctx)

		var maxBlockGas uint64
		consParams := ctx.ConsensusParams()
		if consParams.Block != nil {
			maxBlockGas = uint64(consParams.Block.MaxGas)
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
			return nil, errors.New("known signer private key is empty")
		}
		privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
		if err != nil {
			return nil, errors.New("failed to parse private key")
		}

		knownSignerBz := crypto.PubkeyToAddress(privKey.PublicKey).Bytes()
		nextNonce, err := h.keeper.AccountKeeper.GetSequence(ctx, sdk.AccAddress(knownSignerBz))
		if err != nil {
			nextNonce = 0
		}
		gatewayAddress := common.HexToAddress(params.GatewayContractAddress)

		// Add the source callback queue
		nextNonce, err = h.AddSrcCallbackTxs(ctx, req, nextNonce, chainId, gatewayAddress, privKey, maxBlockGas)
		if err != nil {
			logger.Error("Error during src callback queue walk", "error", err)
			return nil, err
		}

		// TODO: possible issue here, if there are many IBC txs being sent in, they might block
		// other normal txs. We should add a % limit of space IBC txs can take in the proposal.
		err = h.AddPacketTxs(ctx, req, nextNonce, chainId, gatewayAddress, privKey, maxBlockGas)
		if err != nil {
			logger.Error("Error during packet queue walk", "error", err)
			return nil, err
		}

		// 2. Add the rest of the transactions in the incoming request
		if h.txVerifier == nil {
			return nil, errors.New("tx verifier is nil")
		}

		err = h.AddIncomingTxs(ctx, req, maxBlockGas)
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
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}

// AddSrcCallbackTxs adds the source callback transactions to the proposal
func (h *ProposalHandler) AddSrcCallbackTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, nextNonce uint64, chainId *big.Int, gatewayAddress common.Address, privKey *ecdsa.PrivateKey, maxBlockGas uint64) (uint64, error) {
	// Add the source callback queue
	err := h.keeper.SrcCallbackQueue.Walk(ctx, nil, func(key uint64, _ types.PacketQueueItem) (stop bool, err error) {
		// Calldata is a simple call to the gateway executeSrcCallback function
		calldata, err := precompilesgateway.ABI.Pack("executeSrcCallback")
		if err != nil {
			return true, err
		}

		cosmosTx, txBytes, err := h.calldataToSignedTx(ctx, calldata, nextNonce, chainId, &gatewayAddress, privKey)
		if err != nil {
			return true, err
		}

		stop = h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, cosmosTx, txBytes)
		if stop {
			return true, nil
		}

		nextNonce = nextNonce + 1
		return false, nil
	})

	return nextNonce, err
}

// AddPacketTxs adds the packet transactions to the proposal
func (h *ProposalHandler) AddPacketTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, nextNonce uint64, chainId *big.Int, gatewayAddress common.Address, privKey *ecdsa.PrivateKey, maxBlockGas uint64) error {
	err := h.keeper.PacketQueue.Walk(ctx, nil, func(key uint64, _ types.PacketQueueItem) (stop bool, err error) {
		// Calldata is a simple call to the gateway execute function
		calldata, err := precompilesgateway.ABI.Pack("execute")
		if err != nil {
			return true, err
		}

		cosmosTx, txBytes, err := h.calldataToSignedTx(ctx, calldata, nextNonce, chainId, &gatewayAddress, privKey)
		if err != nil {
			return true, err
		}

		if h.txSelector == nil {
			return true, errors.New("tx selector is nil")
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

func (h *ProposalHandler) AddIncomingTxs(ctx sdk.Context, req *abci.RequestPrepareProposal, maxBlockGas uint64) error {
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

		stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
		if stop {
			break
		}
	}

	return nil
}

func (h *ProposalHandler) calldataToSignedTx(ctx sdk.Context, calldata []byte, nonce uint64, chainID *big.Int, contract *common.Address, privKey *ecdsa.PrivateKey) (sdk.Tx, []byte, error) {
	txArgs := &evmtypes.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  CallMaxGas,
		Input:     calldata,
		GasFeeCap: big.NewInt(0),
		GasPrice:  big.NewInt(0),
		ChainID:   chainID,
		Amount:    big.NewInt(0), // No value transfer for contract calls
		GasTipCap: big.NewInt(0),
		To:        contract,
		Accesses:  nil, // No access list for now
	}

	tx := evmtypes.NewTx(txArgs)

	if h.signer == nil {
		return nil, nil, errors.New("signer is nil")
	}

	ethtx := tx.AsTransaction()
	ethtx.ChainId().Set(chainID)

	if ethtx == nil {
		return nil, nil, errors.New("as transaction returned nil")
	}

	signedTx, err := ethcoretypes.SignTx(ethtx, h.signer, privKey)
	if err != nil {
		return nil, nil, err
	}

	tx = &evmtypes.MsgEthereumTx{}
	tx.FromEthereumTx(signedTx)

	if err := tx.ValidateBasic(); err != nil {
		return nil, nil, err
	}

	cosmosTx, err := tx.BuildTx(h.txConfig.NewTxBuilder(), "saga") // TODO: get denom from params
	if err != nil {
		return nil, nil, err
	}

	// Encode transaction by default Tx encoder
	txBytes, err := h.txConfig.TxEncoder()(cosmosTx)
	if err != nil {
		return nil, nil, err
	}

	return cosmosTx, txBytes, nil
}
