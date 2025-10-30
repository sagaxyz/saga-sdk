package gateway

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// ExecuteMethod defines the ABI method name for the Gateway Execute
	// transaction.
	ExecuteMethod = "execute"
	// ExecuteSrcCallbackMethod defines the ABI method name for the Gateway ExecuteSrcCallback
	// transaction.
	ExecuteSrcCallbackMethod = "executeSrcCallback"
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
	Sequence *big.Int
}

// emitNote is a struct used to parse the EmitNote parameter
// used as input in the emitNote method
type emitNote struct {
	Ref  [32]byte
	Data []byte
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
