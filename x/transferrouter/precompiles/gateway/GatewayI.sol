// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title Gateway Interface
/// @notice Interface for the Gateway contract
interface IGateway {
    /// @notice Execute the related call to an IBC packet
    /// @param sequence The packet sequence to execute
    function execute(
        uint256 sequence
    ) external;

    /// @notice Emit a metadata note
    /// @param ref Reference identifier for the note
    /// @param data The note data to emit
    function emitNote(bytes32 ref, bytes calldata data) external;

    /// @notice Pause the contract
    function pause() external;

    /// @notice Unpause the contract
    function unpause() external;

    /// @notice Get the current owner address
    /// @return The address of the current owner
    function owner() external view returns (address);

    /// @notice Event emitted when a call is executed
    event Executed(
        address indexed target,
        uint256 value,
        bytes data,
        bool success,
        bytes result,
        bytes note
    );

    /// @notice Event emitted when a note is emitted
    event Note(bytes32 indexed ref, bytes data);

    /// @notice Event emitted when ownership is transferred
    event OwnershipTransferred(
        address indexed previousOwner,
        address indexed newOwner
    );

    /// @notice Event emitted when the contract is paused
    event Paused(address account);

    /// @notice Event emitted when the contract is unpaused
    event Unpaused(address account);
} 