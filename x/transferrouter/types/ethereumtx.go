package types

import (
	fmt "fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	evmostypes "github.com/evmos/evmos/v20/x/evm/types"
)

// ToMsgEthereumTx converts the call to a MsgEthereumTx, adding the necessary signature and fields
func (c *CallQueueItem) ToMsgEthereumTx() *evmostypes.MsgEthereumTx {
	if c.Call == nil {
		return nil
	}

	// Convert bytes to common.Address for To field
	var toAddr *common.Address
	if len(c.Call.Contract) > 0 {
		addr := common.BytesToAddress(c.Call.Contract)
		toAddr = &addr
	}

	// Convert bytes to common.Address for From field (not used in txArgs but kept for future use)
	_ = func() common.Address {
		if len(c.Call.From) > 0 {
			return common.BytesToAddress(c.Call.From)
		}
		return common.Address{}
	}()

	txArgs := &evmostypes.EvmTxArgs{
		Nonce:     0,     // Will be set by the signer
		GasLimit:  21000, // Standard gas limit for simple transfers
		Input:     c.Call.Data,
		GasFeeCap: big.NewInt(0), // Will be set by the signer
		GasPrice:  big.NewInt(0), // Will be set by the signer
		ChainID:   big.NewInt(1), // Default chain ID, should be configurable
		Amount:    big.NewInt(0), // No value transfer for contract calls
		GasTipCap: big.NewInt(0), // Will be set by the signer
		To:        toAddr,
		Accesses:  nil, // No access list for now
	}

	tx := evmostypes.NewTx(txArgs)
	fmt.Println("!!!tx", tx)
	return tx
}
