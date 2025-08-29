// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/precompiles/authorization"
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

// ExecuteResponse defines the data for the execute response.
type ExecuteResponse struct {
	Success bool
	Result  []byte
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

// NewExecuteAuthorization returns a new execute authorization authz type from the given arguments.
func NewExecuteAuthorization(method *abi.Method, args []interface{}) (common.Address, *ExecuteAuthorization, error) {
	grantee, target, value, data, note, err := checkExecuteAuthzArgs(method, args)
	if err != nil {
		return common.Address{}, nil, err
	}

	executeAuthz := &ExecuteAuthorization{
		Target: target,
		Value:  value,
		Data:   data,
		Note:   note,
	}
	if err = executeAuthz.ValidateBasic(); err != nil {
		return common.Address{}, nil, err
	}

	return grantee, executeAuthz, nil
}

// ExecuteAuthorization represents the authorization for a specific execute call
type ExecuteAuthorization struct {
	Target common.Address `json:"target"`
	Value  *big.Int       `json:"value"`
	Data   []byte         `json:"data"`
	Note   []byte         `json:"note"`
}

// ValidateBasic validates the ExecuteAuthorization
func (e *ExecuteAuthorization) ValidateBasic() error {
	if e.Target == (common.Address{}) {
		return errors.New("target cannot be zero address")
	}
	if e.Value == nil || e.Value.Sign() < 0 {
		return errors.New("value must be non-negative")
	}
	return nil
}

// checkExecuteAuthzArgs checks the arguments for execute authorization
func checkExecuteAuthzArgs(method *abi.Method, args []interface{}) (common.Address, common.Address, *big.Int, []byte, []byte, error) {
	if len(args) != 5 {
		return common.Address{}, common.Address{}, nil, nil, nil, fmt.Errorf("expected 5 arguments, got %d", len(args))
	}

	grantee, ok := args[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, nil, nil, errors.New("invalid grantee address")
	}

	target, ok := args[1].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, nil, nil, nil, errors.New("invalid target address")
	}

	value, ok := args[2].(*big.Int)
	if !ok {
		return common.Address{}, common.Address{}, nil, nil, nil, errors.New("invalid value")
	}

	data, ok := args[3].([]byte)
	if !ok {
		return common.Address{}, common.Address{}, nil, nil, nil, errors.New("invalid data")
	}

	note, ok := args[4].([]byte)
	if !ok {
		return common.Address{}, common.Address{}, nil, nil, nil, errors.New("invalid note")
	}

	return grantee, target, value, data, note, nil
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

// CheckAndAcceptAuthorizationIfNeeded checks if authorization exists and accepts the grant.
// In case the origin is the caller of the address, no authorization is required.
func CheckAndAcceptAuthorizationIfNeeded(
	ctx sdk.Context,
	contract *vm.Contract,
	origin common.Address,
	authzKeeper authzkeeper.Keeper,
	msg *ExecuteMsg,
) (*authz.AcceptResponse, *time.Time, error) {
	if contract.CallerAddress == origin {
		return nil, nil, nil
	}

	auth, _, err := authorization.CheckAuthzExists(ctx, authzKeeper, contract.CallerAddress, origin, ExecuteMsgURL)
	if err != nil {
		return nil, nil, fmt.Errorf(authorization.ErrAuthzDoesNotExistOrExpired, contract.CallerAddress, origin)
	}

	resp, grantExpiration, err := AcceptGrant(ctx, contract.CallerAddress, origin, msg, auth)
	if err != nil {
		return nil, nil, err
	}

	return resp, grantExpiration, nil
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

// AcceptGrant accepts a grant for execute authorization
func AcceptGrant(ctx sdk.Context, caller, origin common.Address, msg *ExecuteMsg, auth authz.Authorization) (*authz.AcceptResponse, *time.Time, error) {
	// Implementation would go here
	return &authz.AcceptResponse{}, nil, nil
}

// UpdateGrantIfNeeded updates the grant in case the contract caller is not the origin of the message.
func UpdateGrantIfNeeded(ctx sdk.Context, contract *vm.Contract, authzKeeper authzkeeper.Keeper, origin common.Address, expiration *time.Time, resp *authz.AcceptResponse) error {
	if contract.CallerAddress != origin {
		return UpdateGrant(ctx, authzKeeper, contract.CallerAddress, origin, expiration, resp)
	}
	return nil
}

// UpdateGrant updates the grant
func UpdateGrant(ctx sdk.Context, authzKeeper authzkeeper.Keeper, caller, origin common.Address, expiration *time.Time, resp *authz.AcceptResponse) error {
	// Implementation would go here
	return nil
}
