package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankKeeper defines the expected interface needed to set denom metadata
type BankKeeper interface {
	SetDenomMetaData(ctx context.Context, metadata banktypes.Metadata)
}

type AclKeeper interface {
	Admin(ctx context.Context, address sdk.AccAddress) bool
}
