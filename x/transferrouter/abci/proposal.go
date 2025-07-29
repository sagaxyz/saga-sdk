package abci

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

type ProposalHandler struct {
	keeper     keeper.Keeper
	txSelector baseapp.TxSelector
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
		// 1. Validate the proposal by checking the contents of the block against the call queue
		return &abci.ResponseProcessProposal{}, nil
	}
}
