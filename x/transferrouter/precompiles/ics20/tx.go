// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ics20

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/evmos/evmos/v20/x/evm/core/vm"

	cmn "github.com/evmos/evmos/v20/precompiles/common"
	"github.com/evmos/evmos/v20/utils"
)

const (
	// TransferMethod defines the ABI method name for the ICS20 Transfer
	// transaction.
	TransferMethod = "transfer"
)

// Transfer implements the ICS20 transfer transactions.
func (p *Precompile) Transfer(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	msg, sender, err := NewMsgTransfer(method, args)
	if err != nil {
		return nil, err
	}

	// check if channel exists and is open
	if !p.channelKeeper.HasChannel(ctx, msg.SourcePort, msg.SourceChannel) {
		return nil, errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", msg.SourcePort, msg.SourceChannel)
	}

	// isCallerSender is true when the contract caller is the same as the sender
	isCallerSender := contract.CallerAddress == sender

	// If the contract caller is not the same as the sender, the sender must be the origin
	if !isCallerSender && origin != sender {
		return nil, fmt.Errorf(ErrDifferentOriginFromSender, origin.String(), sender.String())
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	resp, expiration, err := CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	res, err := p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	if err := UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, resp); err != nil {
		return nil, err
	}

	if contract.CallerAddress != origin && msg.Token.Denom == utils.BaseDenom {
		// escrow address is also changed on this tx, and it is not a module account
		// so we need to account for this on the UpdateDirties
		escrowAccAddress := transfertypes.GetEscrowAddress(msg.SourcePort, msg.SourceChannel)
		escrowHexAddr := common.BytesToAddress(escrowAccAddress)
		// NOTE: This ensures that the changes in the bank keeper are correctly mirrored to the EVM stateDB
		// when calling the precompile from another smart contract.
		// This prevents the stateDB from overwriting the changed balance in the bank keeper when committing the EVM state.
		amt := msg.Token.Amount.BigInt()
		p.SetBalanceChangeEntries(
			cmn.NewBalanceChangeEntry(sender, amt, cmn.Sub),
			cmn.NewBalanceChangeEntry(escrowHexAddr, amt, cmn.Add),
		)
	}

	if err = EmitIBCTransferEvent(
		ctx,
		stateDB,
		p.ABI.Events[EventTypeIBCTransfer],
		p.Address(),
		sender,
		msg.Receiver,
		msg.SourcePort,
		msg.SourceChannel,
		msg.Token,
		msg.Memo,
	); err != nil {
		return nil, err
	}

	// get ERC20 contract address for the token denom, and only emit the event if it is an ERC20 token (or has been registered)
	tokenPairID := p.erc20Keeper.GetTokenPairID(ctx, msg.Token.Denom)
	found := false
	var tokenPair erc20types.TokenPair

	if len(tokenPairID) != 0 {
		tokenPair, found = p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	}

	fullDenomPath := tokenPair.Denom
	// deconstruct the token denomination into the denomination trace info
	// to determine if the sender is the source chain
	if strings.HasPrefix(tokenPair.Denom, "ibc/") {
		fullDenomPath, err = p.transferKeeper.DenomPathFromHash(ctx, tokenPair.Denom)
		if err != nil {
			return nil, err
		}
	}

	// This mimics the behavior of the IBC transfer module, by emitting transfers to the escrow address
	// if the sender is the source chain, and to the null address (burn) if the sender is the destination chain
	if found {
		erc20Addr := tokenPair.GetERC20Contract()
		if transfertypes.SenderChainIsSource(msg.SourcePort, msg.SourceChannel, fullDenomPath) {
			// obtain the escrow address for the source channel end
			escrowAddress := transfertypes.GetEscrowAddress(msg.SourcePort, msg.SourceChannel)
			escrowHexAddr := common.BytesToAddress(escrowAddress)

			// Emit Transfer event sending the tokens to the escrow address
			if err := EmitTransferEvent(ctx, stateDB, erc20Addr, sender, escrowHexAddr, msg.Token.Amount.BigInt()); err != nil {
				return nil, err
			}

		} else {
			// Emit Transfer event sending the tokens to the null address to show a burn
			if err := EmitTransferEvent(ctx, stateDB, erc20Addr, sender, common.HexToAddress("0x0000000000000000000000000000000000000000"), msg.Token.Amount.BigInt()); err != nil {
				return nil, err
			}
		}

	}
	return method.Outputs.Pack(res.Sequence)
}
