// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	// EventTypeGatewayExecute defines the event type for the Gateway Execute transaction.
	EventTypeGatewayExecute = "GatewayExecute"

	// EventTypeGatewayNote defines the event type for the Gateway Note transaction.
	EventTypeGatewayNote = "GatewayNote"

	// ExecuteEventSignature is the signature of the Execute event
	ExecuteEventSignature = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"

	// NoteEventSignature is the signature of the Note event
	NoteEventSignature = "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b926"
)

// EmitGatewayExecuteEvent creates a new Gateway execute event emitted on an Execute transaction.
func EmitGatewayExecuteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	event abi.Event,
	precompileAddr, senderAddr common.Address,
	target common.Address,
	value *big.Int,
	data []byte,
	note []byte,
) error {
	// Prepare the event topics
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and target are indexed
	topics[1], err = cmn.MakeTopic(senderAddr)
	if err != nil {
		return err
	}
	topics[2], err = cmn.MakeTopic(target)
	if err != nil {
		return err
	}

	// Prepare the event data: value, data, note
	arguments := abi.Arguments{event.Inputs[2], event.Inputs[3], event.Inputs[4]}
	packed, err := arguments.Pack(value, data, note)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     precompileAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}

// EmitGatewayNoteEvent creates a new Gateway note event emitted on an EmitNote transaction.
func EmitGatewayNoteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	event abi.Event,
	precompileAddr, senderAddr common.Address,
	ref [32]byte,
	data []byte,
) error {
	// Prepare the event topics
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	var err error
	// sender and ref are indexed
	topics[1], err = cmn.MakeTopic(senderAddr)
	if err != nil {
		return err
	}
	topics[2], err = cmn.MakeTopic(ref)
	if err != nil {
		return err
	}

	// Prepare the event data: data
	arguments := abi.Arguments{event.Inputs[2]}
	packed, err := arguments.Pack(data)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     precompileAddr,
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()), //nolint:gosec // G115
	})

	return nil
}
