package keeper

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/genesis/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

type ICAHostKeeper interface {
	GetAllInterchainAccounts(ctx sdk.Context) []icahosttypes.RegisteredInterchainAccount
}

type IBCChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (channel channeltypes.Channel, ok bool)
}

type IBCTransferKeeper interface {
	GetDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctransfertypes.DenomTrace, bool)
}
