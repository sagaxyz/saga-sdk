# x/transferrouter

# Overview

The transferrouter module is responsible for routing native internal transfers, including IBC transfers, to the EVM. This works by intercepting specific incoming messages, overriding the default behavior, storing the call in a queue, and executing it on the next block as a MsgEthereumTx.

## Design

This module has 3 main components:

1. IBC middleware: in charge of intercepting incoming IBC packets and storing them in the call queue.
2. ABCI++ Prepare and ProcessProposal: in charge of adding the signed calls to the block, and checking the contents of the block against the call queue to avoid a malicious block proposer to add bad calls to the block.
3. Posthandler: in charge of writing the IBC acknowledgement if needed, and also removing the call from the queue.
4. ??? Maybe a contract in the EVM that will execute the actual call.

### ABCI++

We only use the PrepareProposal and ProcessProposal methods of the ABCI++ interface.

PrepareProposal is called only once per block to the block proposer. During this call, the proposer must add the calls in the queue to the block as MsgEthereumTx. To do this, it uses the known signer's private key to sign the transactions. It will then fill up the block with other transactions that are in the mempool, except for any transaction that uses the known signer's key (which should also be rejected by check tx -- TODO: add this check).

ProcessProposal is called for each block to each validator. During this call, the validator will check the contents of the block against the call queue. If the calls do not precisely match the contents of the call queue, the block is rejected. This is in order to avoid a malicious block proposer to add bad calls to the block. Also, any other transaction that use the known signer will be rejected.

### IBC middleware

In the IBC middleware, only the RecvPacket method is overridden, which is called when a new IBC packet is received. It will store the packet in the call queue as a new transfer along with the original packet. The default behavior is overridden and the tokens being transferred end up in an escrow account (known signer's address). Any packet that is not a transfer to the local chain will fall back to the default behavior, such as Packet Forward Middleware packets, or packets that are not transfers.

### Posthandler

The posthandler is called after the MsgEthereumTx is executed, it will check against the list of transactions in the call queue and if the transaction is found, it will write the IBC acknowledgement if needed, and remove the call from the queue.

## Flow

1. A new IBC transfer is received, and in OnRecvPacket, the transfer is stored in the call queue as a new transfer along with the original packet. The default behavior is overridden and the tokens being transferred end up in a temporary escrow account.
2. On the next block (H+1), during PrepareProposal, the block proposer will add at the top of the block as many MsgEthereumTx as there are calls in the queue (TBD: if we have a limit to avoid blocking other normal txs). The signer of the message will be a publicly known private key.
3. During ProcessProposal, validators will check the contents of the block against the call queue. If the calls do not precisely match the contents of the call queue, the block is rejected. This is in order to avoid a malicious block proposer to add bad calls to the block. Also, any other transaction that uses the escrow account will be rejected.
4. During FinalizeBlock, the MsgEthereumTxs will be executed as usual (TBD: details regarding EVM hooks, like the intermediate contract that executes the actual call in the EVM). Each MsgEthereumTx will call a posthandler that will write the IBC acknowledgement if needed, and also remove the call from the queue.

## Security considerations and mitigations

### Malicious block proposer

A malicious block proposer can try to steal funds from the escrow account by adding a bad call to the block. This is mitigated by the ProcessProposal call, which will check the contents of the block against the call queue. If the calls do not precisely match the contents of the call queue, the block is rejected. Additionally, a social slashing mechanism could be implemented to punish the block proposer for doing this, by informing in the logs about any mismatch.

### Attempts to use the escrow account/known signer's key

Any transaction that uses the escrow account or the known signer's key will be rejected. This is mitigated by the ProcessProposal call, which will check the contents of the block against the call queue. If the calls do not match the contents of the call queue, the block is rejected.

Also an antehandler can be implemented to reject any transaction that uses the escrow account or the known signer's key, this is to avoid getting useless transactions into the mempool.