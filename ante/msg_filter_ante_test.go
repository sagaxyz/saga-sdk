package ante_test

import (
	"testing"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	"github.com/stretchr/testify/require"

	"github.com/sagaxyz/saga-sdk/ante"
)

var signers = []sdk.AccAddress{
	sdk.AccAddress([]byte("addr1")),
	sdk.AccAddress([]byte("addr2")),
	sdk.AccAddress([]byte("addr3")),
}

func noOpAnteDecorator() sdk.AnteHandler {
	return func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
}

func TestMsgFilterDecorator(t *testing.T) {
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	chainCodec := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(chainCodec, tx.DefaultSignModes)

	testCases := []struct {
		name     string
		msgs     []sdk.Msg
		filter   ante.FilterFn
		prefixes []string
		expErr   bool
	}{
		{
			"allowed filtered msg",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
			},
			func(sdk.Context, sdk.AccAddress) bool {
				return true
			},
			[]string{"/ibc", "/xxx"},
			false,
		},
		{
			"allowed filtered msg, multiple signers and msgs",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[1].String(),
				},
			},
			func(sdk.Context, sdk.AccAddress) bool {
				return true
			},
			[]string{"/ibc", "/xxx"},
			false,
		},
		{
			"allowed non-filtered msg",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
			},
			func(sdk.Context, sdk.AccAddress) bool {
				return false
			},
			[]string{"/xxx"},
			false,
		},
		{
			"disallowed filtered msg",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
			},
			func(sdk.Context, sdk.AccAddress) bool {
				return false
			},
			[]string{"/ibc", "/xxx"},
			true,
		},
		{
			"disallowed filtered msg, mixed msgs",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
				&banktypes.MsgSend{
					FromAddress: signers[0].String(),
				},
			},
			func(sdk.Context, sdk.AccAddress) bool {
				return false
			},
			[]string{"/ibc", "/xxx"},
			true,
		},
		{
			"disallowed filtered msg, mixed signers",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[1].String(),
				},
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
			},
			func(_ sdk.Context, addr sdk.AccAddress) bool {
				return addr.Equals(signers[0])
			},
			[]string{"/ibc", "/xxx"},
			true,
		},
		{
			"disallowed filtered msg, mixed msgs, mixed signers",
			[]sdk.Msg{
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[0].String(),
				},
				&banktypes.MsgSend{
					FromAddress: signers[0].String(),
				},
				&ibcclienttypes.MsgUpdateClient{
					Signer: signers[1].String(),
				},
				&banktypes.MsgSend{
					FromAddress: signers[1].String(),
				},
			},
			func(_ sdk.Context, addr sdk.AccAddress) bool {
				return addr.Equals(signers[0])
			},
			[]string{"/ibc", "/xxx"},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := ante.NewMsgFilterDecorator(tc.filter, tc.prefixes...)

			txBuilder := txCfg.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msgs...))

			_, err := handler.AnteHandle(sdk.Context{}, txBuilder.GetTx(), false, noOpAnteDecorator())
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
