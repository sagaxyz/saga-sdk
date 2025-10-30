// SPDX-License-Identifier: MIT 
pragma solidity ^0.8.20;

// This contract should only be used to assemble calldata and the events in tests

contract CallbackTest {
    event PacketAcknowledgement(
        string channelId,
        string portId,
        uint64 sequence,
        bytes data,
        bytes acknowledgement
    );

    event PacketTimeout(
        string channelId,
        string portId,
        uint64 sequence,
        bytes data
    );
    /// @dev Callback function to be called on the source chain
    /// after the packet life cycle is completed and acknowledgement is processed
    /// by source chain. The contract address is passed the packet information and acknowledgmeent
    /// to execute the callback logic.
    /// @param channelId the channnel identifier of the packet
    /// @param portId the port identifier of the packet
    /// @param sequence the sequence number of the packet
    /// @param data the data of the packet
    /// @param acknowledgement the acknowledgement of the packet
    function onPacketAcknowledgement(
        string memory channelId,
        string memory portId,
        uint64 sequence,
        bytes memory data,
        bytes memory acknowledgement
    ) external {
        emit PacketAcknowledgement(channelId, portId, sequence, data, acknowledgement);
    }

    /// @dev Callback function to be called on the source chain
    /// after the packet life cycle is completed and the packet is timed out
    /// by source chain. The contract address is passed the packet information
    /// to execute the callback logic.
    /// @param channelId the channnel identifier of the packet
    /// @param portId the port identifier of the packet
    /// @param sequence the sequence number of the packet
    /// @param data the data of the packet
    function onPacketTimeout(
        string memory channelId,
        string memory portId,
        uint64 sequence,
        bytes memory data
    ) external {
        emit PacketTimeout(channelId, portId, sequence, data);
    }
}