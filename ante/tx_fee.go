package ante

import (
	"math"
	"strings"

	errors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// Returns a function mirroring authante.checkTxFeeWithValidatorMinGasPrices but with the ability
// to make the provided transaction URI prefixes feeless for signers passing the freeFn.
func CheckTxFeeWithValidatorMinGasPrices(freeFn FilterFn, freeGasLimit uint64, freePrefixes ...string) authante.TxFeeChecker {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return nil, 0, errors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
		}
		feeCoins := feeTx.GetFee()
		gas := feeTx.GetGas()

		sigTx, ok := tx.(authsigning.SigVerifiableTx)
		if !ok {
			return nil, 0, errors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
		}
		signers, err := sigTx.GetSigners()
		if err != nil {
			return nil, 0, err
		}

		matchAll := true
		for _, msg := range tx.GetMsgs() {
			msgType := sdk.MsgTypeURL(msg)

			var match bool
			for _, prefix := range freePrefixes {
				if strings.HasPrefix(msgType, prefix) {
					match = true
					break
				}
			}
			if !match {
				matchAll = false
				break
			}
		}
		if matchAll { // All messages match a free prefix
			free := true
			for _, signer := range signers {
				if !freeFn(ctx, signer) {
					free = false
					break
				}
			}
			if free {
				//TODO cap gas limit for free transactions
				feeCoins = sdk.NewCoins() // No fee
				return feeCoins, 0, nil
			}
		}

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() {
			minGasPrices := ctx.MinGasPrices()
			if !minGasPrices.IsZero() {
				requiredFees := make(sdk.Coins, len(minGasPrices))

				// Determine the required fees by multiplying each required minimum gas
				// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
				glDec := sdkmath.LegacyNewDec(int64(gas))
				for i, gp := range minGasPrices {
					fee := gp.Amount.Mul(glDec)
					requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
				}

				if !feeCoins.IsAnyGTE(requiredFees) {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
				}
			}
		}

		priority := getTxPriority(feeCoins, int64(gas))
		return feeCoins, priority, nil
	}
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritize as expected.
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
