package types

const (
	// ModuleName defines the module name
	ModuleName = "acl"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for claims
	RouterKey = ModuleName
)

// prefix bytes for the claims module's persistent store
const (
	prefixAllowed = iota + 1
	prefixAdmins
)

// KVStore key prefixes
var (
	KeyPrefixAdmins  = []byte{prefixAdmins}
	KeyPrefixAllowed = []byte{prefixAllowed}
)
