package abi

import (
	"embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
	// evmos precompiles common
	"github.com/evmos/evmos/v20/precompiles/common"
)

// Embed abi json files to the executable binary. Needed when importing as dependency.
//
//go:embed erc20.json gateway.json
var f embed.FS

var (
	ERC20ABI   abi.ABI
	GatewayABI abi.ABI
)

func init() {
	erc20ABI, err := common.LoadABI(f, "erc20.json")
	if err != nil {
		panic(err)
	}
	gatewayABI, err := common.LoadABI(f, "gateway.json")
	if err != nil {
		panic(err)
	}
	ERC20ABI = erc20ABI
	GatewayABI = gatewayABI
}
