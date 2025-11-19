package gateway

import "errors"

var (
	// ErrAllowanceFailed is raised when the allowance fails.
	ErrAllowanceFailed = errors.New("allowance failed")
	// ErrEVMCallFailed is raised when the EVM call fails.
	ErrEVMCallFailed = errors.New("evm call failed")
)
