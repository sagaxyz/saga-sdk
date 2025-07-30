package types

import (
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
)

// ToMsgEthereumTx converts the call to a MsgEthereumTx, adding the necessary signature and fields
func (c *CallQueueItem) ToMsgEthereumTx() *evmostypes.MsgEthereumTx {
	// TODO: Implement this, we'll need the signer and some more stuff passed in
	return &evmostypes.MsgEthereumTx{}
}
