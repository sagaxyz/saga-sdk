package abci

import (
	"bytes"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

type ProposalHandler struct {
	keeper     keeper.Keeper
	txSelector baseapp.TxSelector
	signer     ethtypes.Signer
	txVerifier baseapp.ProposalTxVerifier
}

func NewProposalHandler(keeper keeper.Keeper, txSelector baseapp.TxSelector, signer ethtypes.Signer, txVerifier baseapp.ProposalTxVerifier) *ProposalHandler {
	return &ProposalHandler{
		keeper:     keeper,
		txSelector: txSelector,
		signer:     signer,
		txVerifier: txVerifier,
	}
}

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// 1. Add the calls to the proposal
		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		defer h.txSelector.Clear()

		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			return nil, nil // TODO: handle error
		}

		// Parse the configured private key (in hex format) and derive the corresponding
		// Ethereum address of the known signer.
		privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
		if err != nil {
			return nil, nil // TODO: handle error
		}

		// TODO: possible issue here, if there are many IBC txs being sent in, they might block
		// other normal txs. We should add a % limit of space IBC txs can take in the proposal.
		err = h.keeper.CallQueue.Walk(ctx, nil, func(key uint64, value types.CallQueueItem) (stop bool, err error) {
			ethTx := value.ToMsgEthereumTx()
			signedTx, err := ethcoretypes.SignTx(ethTx.AsTransaction(), h.signer, privKey)
			if err != nil {
				return true, err
			}

			// TODO: might not be the right way to do it, let's circle back later
			msgEthTx := &evmtypes.MsgEthereumTx{}
			err = msgEthTx.FromEthereumTx(signedTx)
			if err != nil {
				return true, err
			}

			msgEthTxBz, err := msgEthTx.Marshal()
			if err != nil {
				return true, err
			}

			added := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, msgEthTx, msgEthTxBz)
			// If the transaction is not added, we stop the walk, because we don't want to execute queued calls out of order
			if !added {
				return true, nil
			}

			return false, nil
		})

		if err != nil {
			return nil, err
		}

		// 2. Add the rest of the transactions in the incoming request
		for _, txBz := range req.Txs {
			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {
				return nil, err
			}

			added := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
			if !added {
				break
			}
		}

		selectedTxs := h.txSelector.SelectedTxs(ctx)

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
		params, err := h.keeper.Params.Get(ctx)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Parse the configured private key (in hex format) and derive the corresponding
		// Ethereum address of the known signer.
		privKey, err := crypto.HexToECDSA(params.KnownSignerPrivateKey)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// TODO: also check that the transaction has only been added once

		knownSignerBz := crypto.PubkeyToAddress(privKey.PublicKey).Bytes()

		for _, tx := range req.Txs {
			msg := evmostypes.MsgEthereumTx{}
			err = msg.UnmarshalBinary(tx)

			// TODO: should we just crash here?
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Check if the signer is the known signer
			ethtx := msg.AsTransaction()

			sender, err := h.signer.Sender(ethtx)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			if !bytes.Equal(sender.Bytes(), knownSignerBz) {
				h.keeper.Logger(ctx).Error("transaction not signed by known signer, proposer might be malicious", "hash", msg.Hash)
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Verify if the transaction comes from the call queue, if it doesn't, return a rejection
			_, callQueueItem, found := h.keeper.GetCallQueueItemByHash(ctx, msg.Hash)
			if !found {
				h.keeper.Logger(ctx).Error("transaction not found in call queue, proposer might be malicious", "hash", msg.Hash)
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Let's also compare the transaction's bytes, might be overkill, let's revisit later if needed
			callQTxBz, err := callQueueItem.ToMsgEthereumTx().AsTransaction().MarshalBinary()
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			blockTxBz, err := msg.AsTransaction().MarshalBinary()
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			if !bytes.Equal(callQTxBz, blockTxBz) {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

		}

		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}
