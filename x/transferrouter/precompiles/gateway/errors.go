package gateway

import "errors"

const (
	// ErrInvalidTarget is raised when the target address is invalid.
	ErrInvalidSequence = "invalid sequence"
	// ErrInvalidSender is raised when the sender is invalid.
	ErrInvalidSender = "invalid sender: %s"
	// ErrInvalidValue is raised when the value is invalid.
	ErrInvalidValue = "invalid value: %s"
	// ErrInvalidData is raised when the data is invalid.
	ErrInvalidData = "invalid data"
	// ErrInvalidNote is raised when the note is invalid.
	ErrInvalidNote = "invalid note"
	// ErrInvalidRef is raised when the reference is invalid.
	ErrInvalidRef = "invalid reference"
	// ErrExecuteSelfCall is raised when trying to execute a call to self.
	ErrExecuteSelfCall = "cannot execute call to self"
	// ErrExecuteCallFailed is raised when the execute call fails.
	ErrExecuteCallFailed = "execute call failed"
	// ErrContractPaused is raised when the contract is paused.
	ErrContractPaused = "contract is paused"
	// ErrUnauthorized is raised when the caller is not authorized.
	ErrUnauthorized = "caller is not authorized"
)

var (
	// ErrAllowanceFailed is raised when the allowance fails.
	ErrAllowanceFailed = errors.New("allowance failed")
	// ErrEVMCallFailed is raised when the EVM call fails.
	ErrEVMCallFailed = errors.New("evm call failed")
)
