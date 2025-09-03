package types

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			Enabled:               true,
			KnownSignerPrivateKey: "f6dba52e479cf5d7ad58bc11177c105ac7b89a02be1d432e77e113fc53377978", // 0x5A6acd4e5766f1dC889a7f7736190323B5685a6a
			KnownSignerAddress:    "0x5A6acd4e5766f1dC889a7f7736190323B5685a6a",
		},
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
