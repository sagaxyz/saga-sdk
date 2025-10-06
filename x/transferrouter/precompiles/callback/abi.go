package callback

import (
	"embed"

	"github.com/ethereum/go-ethereum/accounts/abi"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

var ABI abi.ABI

func init() {
	var err error
	ABI, err = loadABI()
	if err != nil {
		panic(err)
	}
}

func loadABI() (abi.ABI, error) {
	newABI, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return abi.ABI{}, err
	}

	return newABI, nil
}
