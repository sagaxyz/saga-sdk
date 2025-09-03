// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"fmt"
	"math/big"

	"github.com/evmos/evmos/v20/x/evm/core/vm"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// ExecuteMethod defines the ABI method name for the Gateway Execute
	// transaction.
	ExecuteMethod = "execute"
	// EmitNoteMethod defines the ABI method name for the Gateway EmitNote
	// transaction.
	EmitNoteMethod = "emitNote"
	// PauseMethod defines the ABI method name for the Gateway Pause
	// transaction.
	PauseMethod = "pause"
	// UnpauseMethod defines the ABI method name for the Gateway Unpause
	// transaction.
	UnpauseMethod = "unpause"
	// OwnerMethod defines the ABI method name for the Gateway Owner
	// query.
	OwnerMethod = "owner"
)

// EventGatewayExecute is the event type emitted when an execute call is made.
type EventGatewayExecute struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
	Note   []byte
}

// EventGatewayNote is the event type emitted when a note is emitted.
type EventGatewayNote struct {
	Ref  [32]byte
	Data []byte
}

// OwnerResponse defines the data for the owner response.
type OwnerResponse struct {
	Owner common.Address
}

// execute is a struct used to parse the Execute parameter
// used as input in the execute method
type execute struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
	Note   []byte
}

// emitNote is a struct used to parse the EmitNote parameter
// used as input in the emitNote method
type emitNote struct {
	Ref  [32]byte
	Data []byte
}

// CheckOriginAndSender ensures the correct sender is being used.
func CheckOriginAndSender(contract *vm.Contract, origin common.Address, sender common.Address) (common.Address, error) {
	if contract.CallerAddress == sender {
		return sender, nil
	} else if origin != sender {
		return common.Address{}, fmt.Errorf("origin address %s is not the same as sender address %s", origin.String(), sender.String())
	}
	return sender, nil
}

// ExecuteMsg represents an execute message
type ExecuteMsg struct {
	Target common.Address `json:"target"`
	Value  *big.Int       `json:"value"`
	Data   []byte         `json:"data"`
	Note   []byte         `json:"note"`
}

// ExecuteMsgURL is the URL for the execute message
const ExecuteMsgURL = "/saga.transferrouter.v1.MsgExecute"
