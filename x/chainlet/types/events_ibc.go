package types

// IBC events
const (
	EventTypeTimeout              = "timeout"
	EventTypeCreateUpgradePacket  = "create_upgrade_packet"
	EventTypeConfirmUpgradePacket = "confirm_upgrade_packet"
	EventTypeCancelUpgradePacket  = "cancel_upgrade_packet"
	// this line is used by starport scaffolding # ibc/packet/event

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
)
