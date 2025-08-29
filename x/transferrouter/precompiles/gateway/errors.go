// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

const (
	// ErrInvalidTarget is raised when the target address is invalid.
	ErrInvalidTarget = "invalid target address"
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
