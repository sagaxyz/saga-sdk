// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title Gateway Interface
/// @notice Interface for the Gateway contract
interface IGateway {
    /// @notice Execute the next packet in the queue, if any.
    function execute() external;

    // @notice Execute the next source callback in the queue, if any.
    function executeSrcCallback() external;

    // @notice Handle an error acknowledgement or timeout for an IBC transfer packet, by sending the tokens back to the sender.
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

    // @notice Event emitted when a error or timeout is handled.
    event ErrorOrTimeoutHandled(
        uint256 sequence,
        bytes txhash,
        bytes data
    );
} 