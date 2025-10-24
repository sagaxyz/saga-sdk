package utils_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/utils"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenerateIsolatedAddress(t *testing.T) {
	tests := []struct {
		channelID   string
		sender      string
		expected    string
		expectedHex string
	}{
		{
			channelID:   "channel-0",
			sender:      "cosmos1srkyl542x05en2vwwazhce6pachseq4e209ycs",
			expected:    "saga1y0xwpzngnwz0yms5jdyz4lgj3sssrzcxj89hr2",
			expectedHex: "0x23ccE08a689b84f26e1493482aFD128c21018B06",
		},
	}

	for _, test := range tests {
		isolatedAddr := utils.GenerateIsolatedAddress(test.channelID, test.sender)
		sagaAddress := sdk.MustBech32ifyAddressBytes("saga", isolatedAddr.Bytes())
		assert.Equal(t, test.expected, sagaAddress)
		assert.Equal(t, test.expectedHex, common.Address(isolatedAddr).String())
	}
}
