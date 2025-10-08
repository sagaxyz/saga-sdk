package types

// ValidateBasic is used for validating the packet
func (p ConfirmUpgradePacketData) ValidateBasic() error {

	// TODO: Validate the packet data

	return nil
}

// GetBytes is a helper for serialising
func (p ConfirmUpgradePacketData) GetBytes() ([]byte, error) {
	var modulePacket ChainletPacketData

	modulePacket.Packet = &ChainletPacketData_ConfirmUpgradePacket{&p}

	return modulePacket.Marshal()
}
