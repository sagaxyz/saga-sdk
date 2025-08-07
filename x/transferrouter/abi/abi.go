package abi

import (
	"embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
	// evmos precompiles common
	"github.com/evmos/evmos/v20/precompiles/common"
)

// abiPath defines the path to the ERC-20 precompile ABI JSON file.
var abiPath = "abi.json"

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var ABI abi.ABI

func init() {
	newABI, err := common.LoadABI(f, abiPath)
	if err != nil {
		panic(err)
	}
	ABI = newABI
}
