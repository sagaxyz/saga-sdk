package types

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{},
	}
}

// NewGenesisState returns a new GenesisState instance.
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{Params: params}
}

// Validate performs basic genesis state validation.
func (gs GenesisState) Validate() error {
	return nil
}
