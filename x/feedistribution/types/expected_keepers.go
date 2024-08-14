package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
