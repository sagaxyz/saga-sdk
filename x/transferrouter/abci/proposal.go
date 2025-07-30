package abci

import (
	"bytes"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

type ProposalHandler struct {
	keeper      keeper.Keeper
	txSelector  baseapp.TxSelector
	knownSigner []byte
}

func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// 1. Add the calls to the proposal
		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		defer h.txSelector.Clear()

		err := h.keeper.CallQueue.Walk(ctx, nil, func(key uint64, value types.CallQueueItem) (stop bool, err error) {
			// Add the call to the proposal

			var newTx sdk.Tx

			added := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, newTx, []byte{})

			// If the transaction is not added, we stop the walk, because we don't want to execute txs out of order
			if !added {
				return true, nil
			}

			// req.Txs = append(req.Txs, value.Call.ToMsgEthereumTx())
			return false, nil
		})

		if err != nil {
			return nil, err
		}

		// 2. Add the rest of the transactions in the incoming request, but making sure we don't exceed the max block size

		return &abci.ResponsePrepareProposal{}, nil
	}
}

func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		for _, tx := range req.Txs {
			msg := evmostypes.MsgEthereumTx{}
			err := msg.UnmarshalBinary(tx)
			// Check if the signer is the
			if err == nil {
				signer := msg.GetFrom() // TODO: or GetSender()?
				if bytes.Equal(signer.Bytes(), h.knownSigner) {
					// Verify if the transaction comes from the call queue, if it doesn't, return a rejection
					callQueueItem, found := h.keeper.GetCallQueueItemByHash(ctx, msg.Hash)
					if !found {
						return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
					}

					// Let's also compare the transaction's bytes, might be overkill, let's revisit later
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
			}
		}

		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_ACCEPT,
		}, nil
	}
}
