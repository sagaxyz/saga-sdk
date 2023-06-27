package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// GetGenesisStateFromAppState returns GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.Enable && len(gs.Admins) == 0 && len(gs.Allowed) == 0 {
		return errors.New("no allowed or admin address")
	}

	for _, admin := range gs.Admins {
		if admin.Format != AddressFormat_ADDRESS_BECH32 {
			return fmt.Errorf("unsupported admin address format: %s", admin.Format)
		}
		err := admin.Validate()
		if err != nil {
			return fmt.Errorf("admin address invalid: %w", err)
		}
	}
	for _, allowed := range gs.Allowed {
		err := allowed.Validate()
		if err != nil {
			return fmt.Errorf("allowed address invalid: %w", err)
		}
	}

	return gs.Params.Validate()
}
