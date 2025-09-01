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

// emitGatewayExecuteEvent creates a new Gateway execute event emitted on an Execute transaction.
func (p Precompile) emitGatewayExecuteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	precompileAddr, senderAddr common.Address,
	target common.Address,
	value *big.Int,
	data []byte,
	note []byte,
) error {
	event := p.ABI.Events["Execute"]

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

// emitNoteEvent creates a new Gateway note event emitted on an EmitNote transaction.
func (p Precompile) emitNoteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	precompileAddr, senderAddr common.Address,
	ref [32]byte,
	data []byte,
) error {
	event := p.ABI.Events["Note"]

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
