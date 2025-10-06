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
			sender:      "cosmos1k2qsqdthcn84ph8cx2nuq5nacz4f8e9hu4n5ku",
			expected:    "saga1zg3e0kwv2tc2sae3eu4ly45n5e2klk5tku9up2",
			expectedHex: "0x122397d9cC52f0A87731Cf2bF25693A6556fDa8B",
		},
	}

	for _, test := range tests {
		isolatedAddr := utils.GenerateIsolatedAddress(test.channelID, test.sender)
		sagaAddress := sdk.MustBech32ifyAddressBytes("saga", isolatedAddr.Bytes())
		assert.Equal(t, test.expected, sagaAddress)
		assert.Equal(t, test.expectedHex, common.Address(isolatedAddr).String())
	}
}
