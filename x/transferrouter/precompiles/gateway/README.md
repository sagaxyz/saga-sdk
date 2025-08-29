# Gateway Precompile

This directory contains the Gateway precompile contract for the TransferRouter module.

## Overview

The Gateway precompile provides Ethereum-compatible smart contract functionality for executing calls to other contracts and emitting metadata notes within the Saga network.

## Purpose

- Enables smart contract interactions with the TransferRouter module
- Provides a standardized interface for contract execution and metadata emission
- Allows developers to integrate gateway functionality into their dApps

## Structure

The Gateway precompile follows the standard precompile pattern used in the Saga SDK:

- `gateway.go` - Main precompile implementation with the `Precompile` struct and core methods
- `types.go` - Type definitions, constants, and helper functions
- `tx.go` - Transaction handling methods (execute, emitNote, pause, unpause)
- `query.go` - Query methods (owner)
- `events.go` - Event emission functions
- `errors.go` - Error constants
- `abi.json` - Ethereum ABI definition for the precompile contract
- `GatewayI.sol` - Solidity interface for the Gateway contract

## Methods

### Transactions
- `execute(address target, uint256 value, bytes data, bytes note)` - Execute a call to another contract
- `emitNote(bytes32 ref, bytes data)` - Emit a metadata note
- `pause()` - Pause the contract
- `unpause()` - Unpause the contract
- `approve(address grantee, address target, uint256 value, bytes data, bytes note)` - Approve execute authorization
- `revoke(address grantee, address target)` - Revoke execute authorization
- `increaseAllowance(address grantee, address target, uint256 value, bytes data, bytes note)` - Increase execute allowance
- `decreaseAllowance(address grantee, address target, uint256 value, bytes data, bytes note)` - Decrease execute allowance

### Queries
- `owner()` - Get the current owner address

### Events
- `Executed(address target, uint256 value, bytes data, bool success, bytes result, bytes note)` - Emitted when a call is executed
- `Note(bytes32 ref, bytes data)` - Emitted when a note is emitted
- `OwnershipTransferred(address previousOwner, address newOwner)` - Emitted when ownership is transferred
- `Paused(address account)` - Emitted when the contract is paused
- `Unpaused(address account)` - Emitted when the contract is unpaused

## Usage

The Gateway precompile can be called from Ethereum smart contracts deployed on the Saga network to interact with the underlying TransferRouter module.

## Related

- [TransferRouter Module](../../../README.md)
- [ICS20 Precompile](../ics20/README.md)

## Development Status

This is a boilerplate implementation. The following components need to be implemented:

- Actual call execution logic in `executeCall()`
- Owner checking logic in `getOwner()`
- Pause/unpause functionality
- Authorization methods (approve, revoke, etc.)
- Integration with TransferRouter keeper methods
- Proper error handling and validation
- Comprehensive testing 