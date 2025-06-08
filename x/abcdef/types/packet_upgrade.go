package types

// ValidateBasic is used for validating the packet
func (p UpgradePacketData) ValidateBasic() error {

	// TODO: Validate the packet data

	return nil
}

// GetBytes is a helper for serialising
func (p UpgradePacketData) GetBytes() ([]byte, error) {
	var modulePacket AbcdefPacketData

	modulePacket.Packet = &AbcdefPacketData_UpgradePacket{&p}

	return modulePacket.Marshal()
}