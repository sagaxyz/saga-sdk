// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// emitGatewayExecuteEvent creates a new Gateway execute event emitted on an Execute transaction.
/*
   event Executed(
       uint256 sequence,
       bool success,
       bytes txhash,
       bool isCallback,
       bool isSourceCallback,
       bytes ret
   );
*/
func (p Precompile) emitGatewayExecuteEvent(
	ctx sdk.Context,
	stateDB vm.StateDB,
	precompileAddr common.Address,
	sequence uint64,
	success bool,
	txhash []byte,
	isCallback bool,
	isSourceCallback bool,
	ret []byte,
) error {
	event := p.ABI.Events["Executed"]

	// Prepare the event topics
	topics := make([]common.Hash, 1)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	// Prepare the event data: sequence, success, txhash, isCallback, isSourceCallback, ret
	// All parameters are non-indexed, so they go in the data field
	arguments := abi.Arguments{event.Inputs[0], event.Inputs[1], event.Inputs[2], event.Inputs[3], event.Inputs[4], event.Inputs[5]}
	seqBig := new(big.Int).SetUint64(sequence)
	packed, err := arguments.Pack(seqBig, success, txhash, isCallback, isSourceCallback, ret)
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
