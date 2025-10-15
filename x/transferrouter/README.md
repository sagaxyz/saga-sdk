# x/transferrouter

# Overview

The transferrouter module is responsible for routing native internal transfers, including IBC transfers, to the EVM. This works by intercepting specific incoming messages, overriding the default behavior, storing the call in a queue, and executing it on the next block as a MsgEthereumTx.

## Design

This module has 4 main components:

1. IBC middleware: in charge of intercepting incoming IBC packets and storing them in the call queue.
2. ABCI++ PrepareProposal: in charge of adding the signed calls to the gateway precompile contract in the block.
3. Gateway contract: a precompile contract that will send the tokens to the receiver and emit Ethereum logs.
4. Transferrouter keeper: stores the calls to the gateway precompile contract for the incoming IBC transfers and also the source callback queue (acknowledgments and timeouts).

### ABCI++

We only use the PrepareProposal method of the ABCI++ interface.

PrepareProposal is called only once per block to the block proposer. During this call, the proposer must add the calls in the queue to the block as MsgEthereumTx. To do this, it uses the known signer's private key to sign the transactions. It will then fill up the block with other transactions that are in the mempool.

These new transactions just call the execute method of the gateway precompile contract, which doesn't take any arguments, as the precompile contract can read the call queue from the state.

ProcessProposal is not implemented, as we don't need to do any validation of the block.

### IBC middleware

In the IBC middleware the RecvPacket method is called when a new IBC packet is received. It will store the packet in the call queue as a new transfer along with the original packet. The default behavior is overridden and the tokens being transferred end up in an escrow account (the gateway contract address). Any packet that is not a transfer to the local chain will fall back to the default behavior, such as Packet Forward Middleware packets, or packets that are not transfers.

Also the OnTimeoutPacket and OnAcknowledgementPacket methods are overridden to store the source callback queue (acknowledgments and timeouts), if necessary.

### Gateway precompile contract

The gateway precompile contract has 2 methods:
- execute: executes the next call in the queue, which is an incoming IBC transfer.
- executeSrcCallback: executes the next source callback in the queue, which is an acknowledgment or a timeout.

These methods do not accept any arguments and get the necessary information directly from the state, this is in order to avoid a malicious block proposer to add bad calls to the block.

Future work: the methods can be used to make multiple calls per block, this is currently not implemented but could be added in the future by adding a param.

During these executions, any logs produced by the call will be emitted to the EVM, along with an `Executed` event that contains information about the call, including the original tx hash and the success/failure of the call.

```solidity
    event Executed(
        uint256 sequence,
        bool success,
        bytes txhash,
        bool isCallback,
        bool isSourceCallback,
        bytes ret
    );
```

## Flows

### Incoming IBC transfer

1. A new IBC transfer is received, and in OnRecvPacket, the transfer is stored in the call queue as a new transfer along with the original packet. The default behavior is overridden and the tokens being transferred end up in an escrow account (the gateway contract address).
2. On the next block (H+1), during PrepareProposal, the block proposer will add at the top of the block as many MsgEthereumTx as there are calls in the queue . The signer of the message will be a publicly known private key. Note: external actors can also add these calls to the block, as they can call the execute method of the gateway precompile contract.
3. During FinalizeBlock, the created Txs will call the Gateway contract in order to send the tokens to the end receiver. The gateway contract will also emit any necessary logs and produce an IBC acknowledgment.

### Outgoing IBC transfer

The outgoing transfer works slightly different as there is no need to create a new MsgEthereumTx, as long as the sender uses the ICS20 precompile. The transferrouter module includes a copy of the ICS20 precompile to which we added the emission of the a Transfer event.

TBD: the behavior of the native ICS 20 transfer, as we didn't add an event.

### Timeouts and error acknowledgement

When a timeout or an error acknowledgement is received, we proceed to refund the tokens to the sender.

## Callbacks

They are implemented as defined in the [ADR-008](https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-008-app-caller-cbs.md).

### Destination callbacks

The destination callbacks are handled by the gateway precompile contract. If there is a callback defined, the tokens are redirected to a generated isolated address that is derived from the sender and the destination channel.

If a callback is defined but the receiver is different from the isolated address, the tokens are refunded to the sender.

The gateway contract decides whether to execute a callback or perform a normal ERC20 transfer.

### Source callbacks

For source callbacks, when we receive a call to OnTimeoutPacket or OnAcknowledgementPacket, we store the call in the call queue as a new source callback along with the original packet. Then this is executed in the next block as a MsgEthereumTx by the gateway precompile contract.


## Security considerations and mitigations

No additional security considerations are needed, as the calls can't be externally modified, the calls are defined in the gateway precompile contract and the ICS20 precompile.

## Working with other modules/IBC middlewares

### Packet Forward Middleware

This package should not interfere with x/PFM as it will let those transactions pass through with no modifications. These transactions won't be shown on the EVM block explorer.