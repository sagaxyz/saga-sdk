# x/transferrouter

# Overview

The transferrouter module is responsible for routing native internal transfers, including IBC transfers, to the EVM. This works by intercepting specific incoming messages, overriding the default behavior, storing the call in a queue, and executing it on the next block as a MsgEthereumTx.

## Flow

1. A new IBC transfer is received, and in OnRecvPacket, the transfer is stored in the call queue as a new transfer along with the original packet. The default behavior is overridden and the tokens being transferred end up in a temporary escrow account.
2. On the next block (H+1), during PrepareProposal, the block proposer will add at the top of the block as many MsgEthereumTx as there are calls in the queue (TBD: if we have a limit to avoid blocking other normal txs). The signer of the message will be a publicly known private key.
3. During ProcessProposal, validators will check the contents of the block against the call queue. If the calls do not precisely match the contents of the call queue, the block is rejected. This is in order to avoid a malicious block proposer to add bad calls to the block. Also, any other transaction that uses the escrow account will be rejected.
4. During FinalizeBlock, the MsgEthereumTxs will be executed as usual (TBD: details regarding EVM hooks, like the intermediate contract that executes the actual call in the EVM). Each MsgEthereumTx will call a posthandler that will write the IBC acknowledgement if needed, and also remove the call from the queue.