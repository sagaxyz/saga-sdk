// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
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

	// Check if the sender is the known signer
	knownSigner, err := p.getOwner(ctx)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(sender.Bytes(), knownSigner.Bytes()) {
		return nil, fmt.Errorf("sender is not the known signer")
	}

	// Execute the call logic here
	// This is where you would call your keeper methods to perform the actual execution
	nonce, err := p.transferKeeper.AccountKeeper.GetSequence(ctx, p.Address().Bytes())
	if err != nil {
		return nil, err
	}

	msg := ethtypes.NewMessage(
		p.Address(),
		&executeArgs.Target,
		nonce,
		big.NewInt(0), // amount
		6000000,       // gasLimit
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		big.NewInt(0), // gasPrice
		executeArgs.Data,
		ethtypes.AccessList{}, // AccessList
		false,
	)

	result, err := p.evmKeeper.ApplyMessage(ctx, msg, evmtypes.NewNoOpTracer(), true)
	if err != nil {
		return nil, err
	}

	if !result.Failed() {
		for _, log := range result.Logs {
			stateDB.AddLog(log.ToEthereum())
		}
	}

	// Emit the gateway execute event
	if err := p.emitGatewayExecuteEvent(ctx, stateDB, p.Address(), p.Address(), executeArgs.Target, executeArgs.Value, executeArgs.Data, executeArgs.Note, !result.Failed(), result.Ret); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(result.Ret)
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

	if err := p.emitNoteEvent(ctx, stateDB, p.Address(), sender, noteArgs.Ref, noteArgs.Data); err != nil {
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
