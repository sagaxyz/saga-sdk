// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package gateway

import (
	"embed"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	cmn "github.com/cosmos/evm/precompiles/common"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	transferrouterkeeper "github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

// PrecompileAddress of the Gateway EVM extension in hex format.
const PrecompileAddress = "0x5A6A8Ce46E34c2cd998129d013fA0253d3892345"

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var ABI abi.ABI

func init() {
	var err error
	ABI, err = cmn.LoadABI(f, "abi.json")
	if err != nil {
		panic(err)
	}
}

type EVMKeeper interface {
	CallEVMWithData(
		ctx sdk.Context,
		from common.Address,
		contract *common.Address,
		data []byte,
		commit bool,
	) (*vmtypes.MsgEthereumTxResponse, error)
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, method string, args ...interface{}) (*vmtypes.MsgEthereumTxResponse, error)
	ApplyMessage(ctx sdk.Context, msg ethtypes.Message, tracer vm.EVMLogger, commit bool) (*vmtypes.MsgEthereumTxResponse, error)
}

var _ vm.PrecompiledContract = &Precompile{}

type Precompile struct {
	cmn.Precompile
	transferKeeper        transferrouterkeeper.Keeper
	evmKeeper             EVMKeeper
	packetDataUnmarshaler porttypes.PacketDataUnmarshaler
	maxCallbackGas        uint64
}

// NewPrecompile creates a new Gateway Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferrouterkeeper.Keeper,
	evmKeeper EVMKeeper,
	packetDataUnmarshaler porttypes.PacketDataUnmarshaler,
	maxCallbackGas uint64,
) (*Precompile, error) {
	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  ABI,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		transferKeeper:        transferKeeper,
		evmKeeper:             evmKeeper,
		packetDataUnmarshaler: packetDataUnmarshaler,
		maxCallbackGas:        maxCallbackGas,
	}

	// SetAddress defines the address of the Gateway compile contract.
	p.SetAddress(common.HexToAddress(PrecompileAddress))

	return p, nil
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]

	method, err := p.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method.Name))
}

// Run executes the precompiled contract Gateway methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	p.transferKeeper.Logger(ctx).Info("RunSetup!!!!")
	// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err, stateDB, snapshot)()

	return p.RunAtomic(snapshot, stateDB, func() ([]byte, error) {

		switch method.Name {
		// Gateway transactions
		case ExecuteMethod:
			p.transferKeeper.Logger(ctx).Info("ExecuteMethod!!!!")
			bz, err = p.Execute(ctx, evm.Origin, contract, stateDB, method, args)
		case ExecuteSrcCallbackMethod:
			p.transferKeeper.Logger(ctx).Info("ExecuteSrcCallbackMethod!!!!")
			bz, err = p.ExecuteSrcCallback(ctx, evm.Origin, contract, stateDB, method, args)
		default:
			return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
		}

		if err != nil {
			p.transferKeeper.Logger(ctx).Error("error!!222", "error", err)
			return nil, err
		}

		// cost := ctx.GasMeter().GasConsumed() - initialGas
		// if !contract.UseGas(cost) {
		// 	return nil, vm.ErrOutOfGas
		// }

		if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
			p.transferKeeper.Logger(ctx).Error("error!!333", "error", err)
			return nil, err
		}

		logs := stateDB.Logs()
		p.transferKeeper.Logger(ctx).Info("logs!!!!1", "logs", logs)

		return bz, nil
	})

}

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available gateway transactions are:
//   - Execute
//   - EmitNote
//   - Pause
//   - Unpause
func (Precompile) IsTransaction(method string) bool {
	switch method {
	case ExecuteMethod,
		ExecuteSrcCallbackMethod:
		return true
	default:
		return false
	}
}
