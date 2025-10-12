package types

import "errors"

// ValidateBasic is used for validating the packet
func (p ConfirmUpgradePacketData) ValidateBasic() error {
	if p.ChainId == "" {
		return errors.New("chainId cannot be empty")
	}
	if p.Plan == "" {
		return errors.New("plan cannot be empty")
	}
	if p.Height == 0 {
		return errors.New("height has to be positive")
	}

	return nil
}

// GetBytes is a helper for serialising
func (p ConfirmUpgradePacketData) GetBytes() ([]byte, error) {
	var modulePacket ChainletPacketData

	modulePacket.Packet = &ChainletPacketData_ConfirmUpgradePacket{&p}

	return modulePacket.Marshal()
}

// ValidateBasic is used for validating the packet
func (p CreateUpgradePacketData) ValidateBasic() error {
	if p.ChainId == "" {
		return errors.New("chainId cannot be empty")
	}
	if p.Name == "" {
		return errors.New("name cannot be empty")
	}
	if p.Info == "" {
		return errors.New("info cannot be empty")
	}
	if p.Height == 0 {
		return errors.New("height has to be positive")
	}
	return nil
}

// GetBytes is a helper for serialising
func (p CreateUpgradePacketData) GetBytes() ([]byte, error) {
	var modulePacket ChainletPacketData

	modulePacket.Packet = &ChainletPacketData_CreateUpgradePacket{&p}

	return modulePacket.Marshal()
}

// ValidateBasic is used for validating the packet
func (p CancelUpgradePacketData) ValidateBasic() error {
	if p.ChainId == "" {
		return errors.New("chainId cannot be empty")
	}
	if p.Plan == "" {
		return errors.New("plan cannot be empty")
	}
	return nil
}

// GetBytes is a helper for serialising
func (p CancelUpgradePacketData) GetBytes() ([]byte, error) {
	var modulePacket ChainletPacketData

	modulePacket.Packet = &ChainletPacketData_CancelUpgradePacket{&p}

	return modulePacket.Marshal()
}
