// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title Gateway Interface
/// @notice Interface for the Gateway contract
interface IGateway {
    /// @notice Execute the next packet in the queue, if any.
    function execute() external;

    // @notice Execute the next source callback in the queue, if any.
    function executeSrcCallback() external;

    /// @notice Get the current owner address
    /// @return The address of the current owner
    function owner() external view returns (address);

    /// @notice Event emitted when a call is executed
    event Executed(
        uint256 sequence,
        bool success,
        bytes txhash,
        bool isCallback,
        bool isSourceCallback,
        bytes ret
    );
} 