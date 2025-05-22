package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Assets: []RegisteredAsset{},
		// PortId: "", // If port_id is re-added to GenesisState proto
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Validate each asset in the genesis state
	for _, asset := range gs.Assets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid asset in genesis state: %w", err)
		}
		// Add further validation for uniqueness of ibc_denom if needed here,
		// though keeper InitGenesis is also a good place for that.
	}

	// if gs.PortId == "" { // If port_id is re-added
	// 	return fmt.Errorf("port_id cannot be empty")
	// }
	return nil
}

// Validate performs basic validation of the RegisteredAsset fields.
func (ra RegisteredAsset) Validate() error {
	if ra.IbcDenom == "" {
		return fmt.Errorf("asset ibc_denom cannot be empty")
	}
	if ra.OriginalDenom == "" {
		return fmt.Errorf("asset original_denom cannot be empty")
	}
	if ra.DisplayName == "" {
		return fmt.Errorf("asset display_name cannot be empty")
	}
	if len(ra.DenomUnits) == 0 {
		return fmt.Errorf("asset denom_units cannot be empty for %s", ra.IbcDenom)
	}
	for _, unit := range ra.DenomUnits {
		if unit.Denom == "" {
			return fmt.Errorf("denom_unit denom cannot be empty for asset %s", ra.IbcDenom)
		}
		// Exponent 0 is valid (e.g., for the base unit itself)
	}
	return nil
}
