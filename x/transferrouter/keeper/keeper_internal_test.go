package keeper

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
)

// Test getDenomForThisChain covers native unwind and ibc wrapping cases.
func Test_getDenomForThisChain(t *testing.T) {
	// this chain identifiers
	port := "transfer"
	channel := "channel-0"

	// counterparty identifiers
	cPort := "transfer"
	cChannel := "channel-1"

	t.Run("unwind to base denom when counterparty prefix present and no further trace", func(t *testing.T) {
		denom := transfertypes.GetDenomPrefix(cPort, cChannel) + "uatom"
		got := getDenomForThisChain(port, channel, cPort, cChannel, denom)
		require.Equal(t, "uatom", got)
	})

	t.Run("unwind remains IBC when further trace exists", func(t *testing.T) {
		// counterparty prefixed + additional trace
		unwound := transfertypes.GetDenomPrefix("transfer", "channel-2") + "uatom"
		denom := transfertypes.GetDenomPrefix(cPort, cChannel) + unwound
		expected := transfertypes.ParseDenomTrace(unwound).IBCDenom()
		got := getDenomForThisChain(port, channel, cPort, cChannel, denom)
		require.Equal(t, expected, got)
	})

	t.Run("wrap with this chain prefix when no counterparty prefix", func(t *testing.T) {
		denom := "uosmo"
		expected := transfertypes.ParseDenomTrace(transfertypes.GetDenomPrefix(port, channel) + denom).IBCDenom()
		got := getDenomForThisChain(port, channel, cPort, cChannel, denom)
		require.Equal(t, expected, got)
	})
}
