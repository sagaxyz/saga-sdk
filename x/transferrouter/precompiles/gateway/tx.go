// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// Execute executes a call to another contract through the Gateway precompile.
func (p Precompile) Execute(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse the execute arguments
	var executeArgs execute
	if err := method.Inputs.Copy(&executeArgs, args); err != nil {
		return nil, fmt.Errorf("error parsing execute arguments: %w", err)
	}

	// Validate the execute arguments
	if err := validateExecuteArgs(executeArgs); err != nil {
		return nil, err
	}

	// Check origin and sender
	sender, err := CheckOriginAndSender(contract, origin, origin)
	if err != nil {
		return nil, err
	}

	// Create the execute message
	msg := &ExecuteMsg{
		Target: executeArgs.Target,
		Value:  executeArgs.Value,
		Data:   executeArgs.Data,
		Note:   executeArgs.Note,
	}

	// Check and accept authorization if needed
	authzResp, expiration, err := CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	// Execute the call logic here
	// This is where you would call your keeper methods to perform the actual execution
	success, result, err := p.executeCall(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Update grant if needed
	if err := UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, authzResp); err != nil {
		return nil, err
	}

	// Emit events
	if err := p.emitExecuteEvents(ctx, stateDB, sender, executeArgs.Target, executeArgs.Value, executeArgs.Data, executeArgs.Note); err != nil {
		return nil, err
	}

	// Return success response
	response := ExecuteResponse{
		Success: success,
		Result:  result,
	}

	return method.Outputs.Pack(response.Result)
}

// validateExecuteArgs validates the execute arguments
func validateExecuteArgs(args execute) error {
	if args.Target == (common.Address{}) {
		return fmt.Errorf(ErrInvalidTarget)
	}
	if args.Value == nil || args.Value.Sign() < 0 {
		return fmt.Errorf(ErrInvalidValue, args.Value)
	}
	if args.Data == nil {
		return fmt.Errorf(ErrInvalidData)
	}
	return nil
}

// executeCall executes the actual call logic
func (p Precompile) executeCall(ctx sdk.Context, msg *ExecuteMsg) (bool, []byte, error) {
	// TODO: Implement actual call execution logic
	// This would typically involve calling keeper methods to perform the execution
	// For now, we'll just return success with empty result
	return true, []byte{}, nil
}

// emitExecuteEvents emits the execute events
func (p Precompile) emitExecuteEvents(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender common.Address,
	target common.Address,
	value *big.Int,
	data []byte,
	note []byte,
) error {
	// Get the execute event from the ABI
	executeEvent, err := p.ABI.EventByID(common.HexToHash(ExecuteEventSignature))
	if err != nil {
		return err
	}

	// Emit the gateway execute event
	if err := EmitGatewayExecuteEvent(ctx, stateDB, *executeEvent, p.Address(), sender, target, value, data, note); err != nil {
		return err
	}

	return nil
}

// EmitNote handles the emitNote method for emitting metadata notes
func (p Precompile) EmitNote(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Parse the emitNote arguments
	var noteArgs emitNote
	if err := method.Inputs.Copy(&noteArgs, args); err != nil {
		return nil, fmt.Errorf("error parsing emitNote arguments: %w", err)
	}

	// Validate the note arguments
	if err := validateNoteArgs(noteArgs); err != nil {
		return nil, err
	}

	// Check origin and sender
	sender, err := CheckOriginAndSender(contract, origin, origin)
	if err != nil {
		return nil, err
	}

	// Emit the note event
	if err := p.emitNoteEvent(ctx, stateDB, sender, noteArgs.Ref, noteArgs.Data); err != nil {
		return nil, err
	}

	return method.Outputs.Pack()
}

// validateNoteArgs validates the note arguments
func validateNoteArgs(args emitNote) error {
	if args.Ref == [32]byte{} {
		return fmt.Errorf(ErrInvalidRef)
	}
	return nil
}

// emitNoteEvent emits the note event
func (p Precompile) emitNoteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	sender common.Address,
	ref [32]byte,
	data []byte,
) error {
	// Get the note event from the ABI
	noteEvent, err := p.ABI.EventByID(common.HexToHash(NoteEventSignature))
	if err != nil {
		return err
	}

	// Emit the gateway note event
	if err := EmitGatewayNoteEvent(ctx, stateDB, *noteEvent, p.Address(), sender, ref, data); err != nil {
		return err
	}

	return nil
}

// Pause handles the pause method for pausing the contract
func (p Precompile) Pause(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// TODO: Implement pause logic
	// This would typically involve calling keeper methods to pause the contract
	return method.Outputs.Pack()
}

// Unpause handles the unpause method for unpausing the contract
func (p Precompile) Unpause(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// TODO: Implement unpause logic
	// This would typically involve calling keeper methods to unpause the contract
	return method.Outputs.Pack()
}

// Approve handles the approve method for execute authorization
func (p Precompile) Approve(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// This would implement the approve logic for execute authorization
	// Similar to the ICS20 precompile's approve method
	return nil, fmt.Errorf("approve method not yet implemented")
}

// Revoke handles the revoke method for execute authorization
func (p Precompile) Revoke(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// This would implement the revoke logic for execute authorization
	return nil, fmt.Errorf("revoke method not yet implemented")
}

// IncreaseAllowance handles the increaseAllowance method for execute authorization
func (p Precompile) IncreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// This would implement the increaseAllowance logic for execute authorization
	return nil, fmt.Errorf("increaseAllowance method not yet implemented")
}

// DecreaseAllowance handles the decreaseAllowance method for execute authorization
func (p Precompile) DecreaseAllowance(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// This would implement the decreaseAllowance logic for execute authorization
	return nil, fmt.Errorf("decreaseAllowance method not yet implemented")
}

// Allowance handles the allowance method for execute authorization
func (p Precompile) Allowance(
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// This would implement the allowance logic for execute authorization
	return nil, fmt.Errorf("allowance method not yet implemented")
}
