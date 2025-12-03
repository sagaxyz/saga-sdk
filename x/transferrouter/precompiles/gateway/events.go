package gateway

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/contracts"
	evmcmn "github.com/cosmos/evm/precompiles/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// emitGatewayExecuteEvent creates a new Gateway execute event emitted on an Execute transaction.
/*
   event Executed(
       uint256 sequence,
	   string channelId,
       string portId,
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
	channelId string,
	portId string,
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
	arguments := abi.Arguments{event.Inputs[0], event.Inputs[1], event.Inputs[2], event.Inputs[3], event.Inputs[4], event.Inputs[5], event.Inputs[6], event.Inputs[7]}
	seqBig := new(big.Int).SetUint64(sequence)
	packed, err := arguments.Pack(seqBig, channelId, portId, success, txhash, isCallback, isSourceCallback, ret)
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

// emitErrorOrTimeoutHandledEvent creates a new ErrorOrTimeoutHandled event emitted on an HandleErrorOrTimeout transaction.
/*
   event ErrorOrTimeoutHandled(
       uint256 sequence,
       string channelId,
       string portId,
       bytes txhash,
       bytes data
   );
*/
func (p Precompile) emitErrorOrTimeoutHandledEvent(ctx sdk.Context, stateDB vm.StateDB, precompileAddr common.Address, sequence uint64, channelId string, portId string, txhash []byte, data []byte, errorMsg string) error {
	event := p.ABI.Events["ErrorOrTimeoutHandled"]

	// Prepare the event topics
	topics := make([]common.Hash, 1)

	// The first topic is always the signature of the event.
	topics[0] = event.ID

	// Prepare the event data: sequence, txhash, data
	// All parameters are non-indexed, so they go in the data field
	arguments := abi.Arguments{event.Inputs[0], event.Inputs[1], event.Inputs[2], event.Inputs[3], event.Inputs[4], event.Inputs[5]}
	seqBig := new(big.Int).SetUint64(sequence)
	packed, err := arguments.Pack(seqBig, channelId, portId, txhash, data, errorMsg)
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

// EmitTransferEvent creates a new Transfer event emitted (ERC20). In order to show the IBC transfer in the block explorer.
func (p Precompile) EmitTransferEvent(ctx sdk.Context, stateDB vm.StateDB, precompileAddr, from, to common.Address, value *big.Int) error {
	// Prepare the event topics
	topics := make([]common.Hash, 3)

	// The first topic is always the signature of the event.
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI
	topics[0] = erc20.Events["Transfer"].ID

	var err error
	topics[1], err = evmcmn.MakeTopic(from)
	if err != nil {
		return err
	}

	topics[2], err = evmcmn.MakeTopic(to)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{
		{
			Name:    "value",
			Type:    abi.Type{T: abi.IntTy, Size: 256},
			Indexed: false,
		},
	}
	packed, err := arguments.Pack(value)
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
