package types

const (
	// ModuleName string name of module
	ModuleName = "filter"

	// StoreKey key for base fee.
	StoreKey = ModuleName

	// RouterKey uses module name for routing
	RouterKey = ModuleName

	// TransientKey is the key to access the FeeMarket transient store, that is reset
	// during the Commit phase.
	TransientKey = "transient_" + ModuleName
)

// prefix bytes for the filter persistent store
const ()

const ()

// KVStore key prefixes
var ()

// Transient Store key prefixes
var ()
