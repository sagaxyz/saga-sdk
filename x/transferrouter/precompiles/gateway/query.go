// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

// Owner executes an owner query through the Gateway precompile.
func (p Precompile) Owner(
	ctx sdk.Context,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// The owner method takes no arguments, so we don't need to parse any
	// Validate that no arguments were passed
	if len(args) != 0 {
		return nil, fmt.Errorf("owner method takes no arguments, got %d", len(args))
	}

	// Execute the owner logic here
	// This is where you would call your keeper methods to get the owner information
	owner, err := p.getOwner(ctx)
	if err != nil {
		return nil, err
	}

	// Return the owner response
	response := OwnerResponse{
		Owner: owner,
	}

	return method.Outputs.Pack(response.Owner)
}

// getOwner gets the current owner address
func (p Precompile) getOwner(ctx sdk.Context) (common.Address, error) {
	// TODO: Implement actual owner checking logic
	// This would typically involve calling keeper methods to get the current owner
	// For now, we'll return a hardcoded owner address (same as in the Solidity contract)
	return common.HexToAddress("0x5A6acd4e5766f1dC889a7f7736190323B5685a6a"), nil
}
