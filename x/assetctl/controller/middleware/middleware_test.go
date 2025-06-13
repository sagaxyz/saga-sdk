package middleware

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"

	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/keeper"
	"github.com/stretchr/testify/require"
)

func TestAnteHandler(t *testing.T) {
	// Setup
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCdc := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCdc, tx.DefaultSignModes)

	logger := log.NewNopLogger()
	addressCodec := address.NewBech32Codec("saga")
	storeKey := storetypes.NewKVStoreKey("assetctl")
	storeSvc := runtime.NewKVStoreService(storeKey)
	keeper := keeper.NewKeeper(storeSvc, protoCdc, logger, addressCodec)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(cms, types.Header{}, false, logger)

	// Test msg with no supported assets
	antehandler := NewAssetControlAnteHandler(keeper)
	msg := &ibctransfertypes.MsgTransfer{
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
		Sender:        "cosmos1test",
		Receiver:      "cosmos1test",
		Token:         sdk.NewCoin("ibc/abcdef", math.NewInt(100)),
	}

	builder := txConfig.NewTxBuilder()
	builder.SetMsgs(msg)

	_, err = antehandler.AnteHandle(ctx, builder.GetTx(), false, nil)
	require.Error(t, err)

	// Now we add the asset to the supported assets and test again
	err = keeper.SupportedAssets.Set(ctx, collections.Join(
		msg.SourceChannel,
		msg.Token.Denom,
	))
	require.NoError(t, err)

	emptyAnteHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	}

	_, err = antehandler.AnteHandle(ctx, builder.GetTx(), false, emptyAnteHandler)
	require.NoError(t, err)

	// Recursive authz (2 levels of nesting)
	msg1 := &ibctransfertypes.MsgTransfer{
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
		Sender:        "cosmos1test",
		Receiver:      "cosmos1test",
		Token:         sdk.NewCoin("ibc/abcdefNOTREGISTERED", math.NewInt(100)),
	}

	anyMsg1, err := codectypes.NewAnyWithValue(msg1)
	require.NoError(t, err)

	msg2 := &authztypes.MsgExec{
		Grantee: "cosmos1test",
		Msgs:    []*codectypes.Any{anyMsg1},
	}

	anyMsg2, err := codectypes.NewAnyWithValue(msg2)
	require.NoError(t, err)

	authzMsg := &authztypes.MsgExec{
		Grantee: "cosmos1test",
		Msgs:    []*codectypes.Any{anyMsg2},
	}

	builder = txConfig.NewTxBuilder()
	builder.SetMsgs(authzMsg)

	_, err = antehandler.AnteHandle(ctx, builder.GetTx(), false, emptyAnteHandler)
	require.Error(t, err)
}
