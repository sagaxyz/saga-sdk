package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

func NewAddress(format AddressFormat, value string) *Address {
	return &Address{
		Format: format,
		Value:  value,
	}
}

func LoadAddress(format AddressFormat, data []byte) (address *Address, err error) {
	var value string
	switch format {
	case AddressFormat_ADDRESS_BECH32:
		err = sdk.VerifyAddressFormat(data)
		if err != nil {
			return
		}
		addr := sdk.AccAddress(data)
		value = addr.String()
	case AddressFormat_ADDRESS_EIP55:
		addr := common.BytesToAddress(data)
		value = addr.Hex()
	default:
		err = fmt.Errorf("invalid address format: %s", format)
		return
	}

	address = NewAddress(format, value)
	return
}

func GuessAddress(value string) (address *Address, err error) {
	if common.IsHexAddress(value) {
		address = NewAddress(AddressFormat_ADDRESS_EIP55, value)
		return
	}
	_, err = sdk.AccAddressFromBech32(value)
	if err != nil {
		return
	}

	address = NewAddress(AddressFormat_ADDRESS_BECH32, value)
	return
}

func (a *Address) Bytes() []byte {
	switch a.Format {
	case AddressFormat_ADDRESS_BECH32:
		addr, err := sdk.AccAddressFromBech32(a.Value)
		if err != nil {
			panic(fmt.Sprintf("invalid bech32 address '%s': %s", a.Value, err))
		}
		return addr.Bytes()
	case AddressFormat_ADDRESS_EIP55:
		if !common.IsHexAddress(a.Value) {
			panic(fmt.Sprintf("invalid bech32 address '%s'", a.Value))
		}
		addr := common.HexToAddress(a.Value)
		return addr.Bytes()
	default:
		panic(fmt.Sprintf("invalid address format: %d", a.Format))
	}
}

func (a *Address) Validate() error {
	switch a.Format {
	case AddressFormat_ADDRESS_BECH32:
		_, err := sdk.AccAddressFromBech32(a.Value)
		if err != nil {
			return fmt.Errorf("bech32 address '%s' invalid: %w", a.Value, err)
		}
	case AddressFormat_ADDRESS_EIP55:
		if !common.IsHexAddress(a.Value) {
			return fmt.Errorf("EIP-55 address '%s' is not a hex address", a.Value)
		}
	default:
		return fmt.Errorf("invalid address format: %d", a.Format)
	}
	return nil
}
