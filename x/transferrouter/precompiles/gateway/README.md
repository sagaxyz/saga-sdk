# Gateway Precompile

This directory contains the Gateway precompile for the `transferrouter` module. It exposes a minimal EVM interface used by contracts and off-chain callers to process queued IBC packets, their callbacks, and error/timeout handling on Saga EVM.

## Address

- **Precompile address**: `0x5A6A8Ce46E34c2cd998129d013fA0253d3892345`

## What it does

The Gateway precompile provides three main functions to process queued IBC operations:

- **execute()**: Processes the next queued IBC transfer packet from the destination queue. Depending on the packet:
  - Performs an ERC20 transfer to the intended recipient; or
  - Executes a destination-side callback against a target contract after setting a temporary allowance from an isolated address.
  - On success: writes a success acknowledgement to IBC.
  - On failure: burns the received tokens and writes an error acknowledgement to IBC.
  - Emits an `Executed` event for every attempt (success or failure) and forwards any EVM logs produced by the underlying calls.

- **executeSrcCallback()**: Processes the next queued source-side callback (acknowledgement or timeout) by calling the originating contract's callback function (`onPacketAcknowledgement` or `onPacketTimeout`). Emits an `Executed` event with `isSourceCallback` set to `true`.

- **handleErrorOrTimeout()**: Handles error acknowledgements or timeouts for IBC transfer packets by refunding tokens to the original sender:
  - If the token is native to this chain (source chain), mints the tokens and transfers them to the sender.
  - If the token is from another chain, unescrows the tokens from the IBC escrow address and transfers them to the sender.
  - Emits an `ErrorOrTimeoutHandled` event and an ERC20 `Transfer` event for visibility in block explorers.

## Solidity interface

The ABI is provided in `abi.json`. The Solidity interface lives in `GatewayI.sol` and can be used directly from contracts:

```solidity
pragma solidity ^0.8.20;

interface IGateway {
    /// @notice Execute the next packet in the queue, if any.
    function execute() external;

    /// @notice Execute the next source callback in the queue, if any.
    function executeSrcCallback() external;

    /// @notice Handle an error acknowledgement or timeout for an IBC transfer packet, by sending the tokens back to the sender.
    function handleErrorOrTimeout() external;

    /// @notice Event emitted when a call is executed
    event Executed(
        uint256 sequence,
        bool success,
        bytes txhash,
        bool isCallback,
        bool isSourceCallback,
        bytes ret
    );

    /// @notice Event emitted when an error or timeout is handled.
    event ErrorOrTimeoutHandled(
        uint256 sequence,
        bytes txhash,
        bytes data
    );
}
```

## Event semantics

### Executed Event

`Executed(uint256 sequence, bool success, bytes txhash, bool isCallback, bool isSourceCallback, bytes ret)`

- **sequence**: IBC packet sequence processed by this call.
- **success**: Whether the inner EVM call(s) completed successfully.
- **txhash**: Original transaction hash that enqueued the packet (as raw bytes).
- **isCallback**: Whether this processed a destination-side callback.
- **isSourceCallback**: Whether this processed a source-side callback (ack/timeout).
- **ret**: Raw return bytes from the last inner EVM call (if any).

### ErrorOrTimeoutHandled Event

`ErrorOrTimeoutHandled(uint256 sequence, bytes txhash, bytes data)`

- **sequence**: IBC packet sequence that was handled.
- **txhash**: Original transaction hash that enqueued the packet (as raw bytes).
- **data**: Raw packet data from the IBC transfer.

## File layout

- `gateway.go` – Precompile wiring, dispatch, gas handling, and address.
- `tx.go` – Core logic for packet execution, callbacks, and error/timeout handling.
- `events.go` – Emission of `Executed` and `ErrorOrTimeoutHandled` events, log forwarding, and ERC20 Transfer event emission.
- `calldata.go` – Helpers to build EVM calldata from IBC transfer packets.
- `types.go` – Method names, helper structs, and constants.
- `errors.go` – Error values and messages.
- `abi.json` – ABI for the precompile.
- `GatewayI.sol` – Solidity interface.

## Behavior details

All three functions are transactions (state-changing) and will process at most one queued item per call. On success, logs from the inner execution are forwarded to the caller's receipt and the appropriate event is emitted by the precompile. The `execute()` function always succeeds from the EVM perspective (returns no error) to ensure the transaction appears in block explorers, even if the inner execution fails. The `success` field in the `Executed` event indicates the actual outcome.

### execute()

- **Destination callbacks**: The precompile creates a temporary allowance from an isolated address (derived from the sender and destination channel), invokes the target contract with the provided calldata, and verifies that no funds remain in the isolated address after execution. If tokens remain, the execution fails to prevent funds from becoming irretrievable.
- **Plain ERC20 transfers**: The precompile calls the token contract's `transfer(address,uint256)` on behalf of the gateway address.
- **Failure handling**: If execution fails, the received tokens are burned (sent to the transferrouter module and then burned) and an error acknowledgement is written to IBC.
- **Success handling**: If execution succeeds, a success acknowledgement is written to IBC.
- **Event emission**: An `Executed` event is always emitted, even when the inner execution fails, so explorers and off-chain indexers can observe outcomes.

### executeSrcCallback()

- Processes source-side callbacks (acknowledgements or timeouts) by calling the originating contract's callback interface.
- For acknowledgements: calls `onPacketAcknowledgement(channel, port, sequence, data, ack)`.
- For timeouts: calls `onPacketTimeout(channel, port, sequence, data)`.
- The call is made from the txrouter module account address.
- Emits an `Executed` event with `isSourceCallback` set to `true`.

### handleErrorOrTimeout()

- Handles error acknowledgements or timeouts by refunding tokens to the original sender.
- For native tokens (source chain): mints the tokens and transfers them to the sender.
- For foreign tokens: unescrows the tokens from the IBC escrow address and transfers them to the sender.
- Emits an `ErrorOrTimeoutHandled` event and an ERC20 `Transfer` event for block explorer visibility.
- Removes the packet from the error/timeout queue after processing.

## Related

- `x/transferrouter` module overview: `../../README.md`

## Status

The precompile currently exposes `execute()`, `executeSrcCallback()`, and `handleErrorOrTimeout()`. Functionality such as pause/unpause, ownership, notes, or per-target approvals is not part of this precompile.
