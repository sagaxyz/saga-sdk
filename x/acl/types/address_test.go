package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/saga-sdk/crypto/ethsecp256k1"
	"github.com/stretchr/testify/assert"
)

func TestAddressValidate(t *testing.T) {
	key1, _ := ethsecp256k1.GenerateKey()
	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes())

	testCases := []struct {
		name     string
		format   AddressFormat
		value    string
		expError bool
	}{
		{
			"valid bech32",
			AddressFormat_ADDRESS_BECH32,
			addr1.String(),
			false,
		},
		{
			"valid hex/EIP-55",
			AddressFormat_ADDRESS_EIP55,
			ethAddr1.String(),
			false,
		},
		{
			"invalid hex/EIP-55",
			AddressFormat_ADDRESS_EIP55,
			addr1.String(),
			true,
		},
		{
			"invalid bech32",
			AddressFormat_ADDRESS_BECH32,
			ethAddr1.String(),
			true,
		},
		{
			"empty hex/EIP-55",
			AddressFormat_ADDRESS_EIP55,
			"",
			true,
		},
		{
			"empty bech32",
			AddressFormat_ADDRESS_BECH32,
			"",
			true,
		},
		{
			"missing hex/EIP-55 format",
			0,
			ethAddr1.String(),
			true,
		},
		{
			"missing bech32 format",
			0,
			addr1.String(),
			true,
		},
	}

	for _, tc := range testCases {
		addr := NewAddress(tc.format, tc.value)
		err := addr.Validate()
		if tc.expError {
			assert.Error(t, err, tc.name)
		} else {
			assert.NoError(t, err, tc.name)
		}
	}
}

func TestAddressLoad(t *testing.T) {
	key1, _ := ethsecp256k1.GenerateKey()
	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes())

	testCases := []struct {
		name     string
		format   AddressFormat
		value    []byte
		expError bool
	}{
		{
			"valid bech32",
			AddressFormat_ADDRESS_BECH32,
			addr1.Bytes(),
			false,
		},
		{
			"valid hex/EIP-55",
			AddressFormat_ADDRESS_EIP55,
			ethAddr1.Bytes(),
			false,
		},
		{
			"missing hex/EIP-55 format",
			0,
			ethAddr1.Bytes(),
			true,
		},
		{
			"missing bech32 format",
			0,
			addr1.Bytes(),
			true,
		},
	}

	for _, tc := range testCases {
		addr, err := LoadAddress(tc.format, tc.value)
		if tc.expError {
			assert.Error(t, err, tc.name)
		} else {
			assert.NoError(t, err, tc.name)

			err = addr.Validate()
			assert.NoError(t, err, tc.name)

			assert.Equal(t, addr.Bytes(), tc.value)
		}
	}
}

func TestAddressGuess(t *testing.T) {
	key1, _ := ethsecp256k1.GenerateKey()
	addr1 := sdk.AccAddress(key1.PubKey().Address())
	ethAddr1 := common.BytesToAddress(key1.PubKey().Bytes())

	testCases := []struct {
		name      string
		address   string
		expFormat AddressFormat
		expError  bool
	}{
		{
			"valid bech32",
			addr1.String(),
			AddressFormat_ADDRESS_BECH32,
			false,
		},
		{
			"valid hex/EIP-55",
			ethAddr1.String(),
			AddressFormat_ADDRESS_EIP55,
			false,
		},
		{
			"empty string",
			"",
			0,
			true,
		},
		{
			"invalid string",
			"abcde",
			0,
			true,
		},
	}

	for _, tc := range testCases {
		addr, err := GuessAddress(tc.address)
		if tc.expError {
			assert.Error(t, err, tc.name)
		} else {
			assert.NoError(t, err, tc.name)

			err = addr.Validate()
			assert.NoError(t, err, tc.name)

			assert.Equal(t, addr.Format, tc.expFormat)
		}
	}
}
