package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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
		_, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return fmt.Errorf("admin address '%s' invalid: %w", admin, err)
		}
	}
	for _, allowed := range gs.Allowed {
		if !common.IsHexAddress(allowed) {
			return fmt.Errorf("allowed address '%s' is not an ethereum address", allowed)
		}
	}

	return gs.Params.Validate()
}
