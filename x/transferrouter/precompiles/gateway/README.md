# Gateway Precompile

This directory contains the Gateway precompile for the `transferrouter` module. It exposes a minimal EVM interface used by contracts and off-chain callers to process queued IBC packets and their callbacks on Saga EVM.

## Address

- **Precompile address**: `0x5A6A8Ce46E34c2cd998129d013fA0253d3892345`

## What it does

- **execute()**: Processes the next queued IBC transfer packet. Depending on the packet:
  - Performs an ERC20 transfer to the intended recipient; or
  - Executes a destination-side callback against a target contract after setting a temporary allowance from an isolated address.
- **executeSrcCallback()**: Processes the next queued source-side callback (acknowledgement or timeout) by calling the originating contract’s callback function.
- Emits an `Executed` event for every attempt (success or failure) and forwards any EVM logs produced by the underlying calls.

## Solidity interface

The ABI is provided in `abi.json`. The Solidity interface lives in `GatewayI.sol` and can be used directly from contracts:

```solidity
pragma solidity ^0.8.20;

interface IGateway {
    function execute() external;
    function executeSrcCallback() external;
    event Executed(
        uint256 sequence,
        bool success,
        bytes txhash,
        bool isCallback,
        bool isSourceCallback,
        bytes ret
    );
}
```

## Event semantics

`Executed(uint256 sequence, bool success, bytes txhash, bool isCallback, bool isSourceCallback, bytes ret)`

- **sequence**: IBC packet sequence processed by this call.
- **success**: Whether the inner EVM call(s) completed successfully.
- **txhash**: Original transaction hash that enqueued the packet (as raw bytes).
- **isCallback**: Whether this processed a destination-side callback.
- **isSourceCallback**: Whether this processed a source-side callback (ack/timeout).
- **ret**: Raw return bytes from the last inner EVM call (if any).

## Using from Solidity

```solidity
pragma solidity ^0.8.20;

interface IGateway {
    function execute() external;
    function executeSrcCallback() external;
}

contract UseGateway {
    address constant GATEWAY = 0x5A6A8Ce46E34c2cd998129d013fA0253d3892345;

    function processNextPacket() external {
        IGateway(GATEWAY).execute();
    }

    function processNextSourceCallback() external {
        IGateway(GATEWAY).executeSrcCallback();
    }
}
```

Notes:
- Both functions are transactions (state-changing). They will process at most one queued item per call.
- On success, logs from the inner execution are forwarded to the caller’s receipt and an `Executed` event is emitted by the precompile.

## File layout

- `gateway.go` – Precompile wiring, dispatch, gas handling, and address.
- `tx.go` – Core logic for packet execution and callbacks.
- `events.go` – Emission of the `Executed` event and log forwarding.
- `calldata.go` – Helpers to build EVM calldata from IBC transfer packets.
- `types.go` – Method names, helper structs, and constants.
- `errors.go` – Error values and messages.
- `abi.json` – ABI for the precompile.
- `GatewayI.sol` – Solidity interface.

## Behavior details

- For destination callbacks, the precompile creates a short-lived allowance from an isolated address, invokes the target contract with the provided calldata, and ensures no funds remain in the isolated address after execution.
- For plain ERC20 transfers, the precompile calls the token contract’s `transfer(address,uint256)` on behalf of the gateway.
- In all cases, an `Executed` event is emitted even when the inner execution fails, so explorers and off-chain indexers can observe outcomes.

## Related

- `x/transferrouter` module overview: `../../README.md`

## Status

The precompile currently exposes only `execute()` and `executeSrcCallback()`. Functionality such as pause/unpause, ownership, notes, or per-target approvals is not part of this precompile.