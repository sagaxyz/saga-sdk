package gateway

import (
	"embed"
	"fmt"
	"math/big"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	cmn "github.com/cosmos/evm/precompiles/common"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	transferrouterkeeper "github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
)

// PrecompileAddress of the Gateway EVM extension in hex format.
var PrecompileAddress = common.HexToAddress("0x5A6A8Ce46E34c2cd998129d013fA0253d3892345")

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
		gasCap *big.Int,
	) (*vmtypes.MsgEthereumTxResponse, error)
	CallEVM(ctx sdk.Context, abi abi.ABI, from, contract common.Address, commit bool, gasCap *big.Int, method string, args ...interface{}) (*vmtypes.MsgEthereumTxResponse, error)
	ApplyMessage(ctx sdk.Context, msg core.Message, tracer *tracing.Hooks, commit bool, internal bool) (*vmtypes.MsgEthereumTxResponse, error)
}

var _ vm.PrecompiledContract = &Precompile{}

type Precompile struct {
	cmn.Precompile
	ABI                   abi.ABI
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
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
		},
		ABI:                   ABI,
		transferKeeper:        transferKeeper,
		evmKeeper:             evmKeeper,
		packetDataUnmarshaler: packetDataUnmarshaler,
		maxCallbackGas:        maxCallbackGas,
	}

	// SetAddress defines the address of the Gateway compile contract.
	p.SetAddress(PrecompileAddress)

	return p, nil
}

// RequiredGas calculates the precompiled contract's base gas rate.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]

	method, err := p.ABI.MethodById(methodID)
	if err != nil {
		// This should never happen since this method is going to fail during Run
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

// Run executes the precompiled contract Gateway methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
		if err != nil {
			return nil, err
		}

		// This handles any out of gas errors that may occur during the execution of a precompile tx or query.
		// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
		switch method.Name {
		// Gateway transactions
		case ExecuteMethod:
			bz, err = p.Execute(ctx, evm.Origin, contract, evm.StateDB, method, args)
		case ExecuteSrcCallbackMethod:
			bz, err = p.ExecuteSrcCallback(ctx, evm.Origin, contract, evm.StateDB, method, args)
		case HandleErrorOrTimeoutMethod:
			bz, err = p.HandleErrorOrTimeout(ctx, evm.Origin, contract, evm.StateDB, method, args)
		default:
			return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
		}

		if err != nil {
			return nil, err
		}

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
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case ExecuteMethod,
		ExecuteSrcCallbackMethod,
		HandleErrorOrTimeoutMethod:
		return true
	default:
		return false
	}
}
