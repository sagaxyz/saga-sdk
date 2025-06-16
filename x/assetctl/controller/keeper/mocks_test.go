package keeper_test

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// MockICAHostKeeper is a mock implementation of ICAHostKeeper
type MockICAHostKeeper struct {
	Accounts []icahosttypes.RegisteredInterchainAccount
}

func (m MockICAHostKeeper) GetAllInterchainAccounts(ctx sdk.Context) []icahosttypes.RegisteredInterchainAccount {
	return m.Accounts
}

// MockIBCChannelKeeper is a mock implementation of IBCChannelKeeper
type MockIBCChannelKeeper struct {
	Channel channeltypes.Channel
	Exists  bool
}

func (m MockIBCChannelKeeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return m.Channel, m.Exists
}

// MockIBCTransferKeeper is a mock implementation of IBCTransferKeeper
type MockIBCTransferKeeper struct {
	DenomTrace ibctransfertypes.DenomTrace
	Exists     bool
}

func (m MockIBCTransferKeeper) GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool) {
	return m.DenomTrace, m.Exists
}

// MockAccountKeeper is a mock implementation of AccountKeeper
// Returns a fixed address for the module name used in tests
type MockAccountKeeper struct{}

func (m MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	if name == "assetctl" {
		return sdk.AccAddress([]byte("cosmos1test"))
	}
	return nil
}
