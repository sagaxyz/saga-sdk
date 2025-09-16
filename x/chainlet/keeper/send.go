package keeper

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ccvtypes "github.com/cosmos/interchain-security/v5/x/ccv/types"

	"github.com/sagaxyz/saga-sdk/x/chainlet/types"
)

func (k *Keeper) getConsumerConnectionID(ctx sdk.Context) (connectionID string, err error) {
	ccvChannelID, found := k.consumerKeeper.GetProviderChannel(ctx)
	if !found {
		err = errors.New("channel ID for consumer not found")
		return
	}
	ccvChannel, found := k.channelKeeper.GetChannel(ctx, ccvtypes.ConsumerPortID, ccvChannelID)
	if !found {
		err = fmt.Errorf("consumer channel %s not found", ccvChannelID)
		return
	}
	if len(ccvChannel.GetConnectionHops()) == 0 {
		err = fmt.Errorf("no connections for channel %s", ccvChannelID)
		return
	}
	connectionID = ccvChannel.GetConnectionHops()[0]
	return
}

func (k Keeper) Send(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Send only for the last block before an upgrade
	plan, err := k.upgradeKeeper.GetUpgradePlan(ctx)
	if err != nil {
		if errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
			return nil
		}
		return err
	}
	if sdkCtx.BlockHeight() != plan.Height-2 {
		k.Logger(sdkCtx).Debug(fmt.Sprintf("skipping until the upgrade height is reached: %d >= %d", plan.Height-1, sdkCtx.BlockHeight()))
		return nil
	}

	// Find a channel for the provider chain
	var sourceChannel *channeltypes.IdentifiedChannel
	ccvConnectionID, err := k.getConsumerConnectionID(sdkCtx)
	if err != nil {
		return err
	}
	channels := k.channelKeeper.GetAllChannelsWithPortPrefix(sdkCtx, types.PortID)
	for _, channel := range channels {
		if channel.State != channeltypes.OPEN {
			continue
		}
		if len(channel.ConnectionHops) == 0 || channel.ConnectionHops[0] != ccvConnectionID {
			continue
		}

		sourceChannel = &channel
	}
	if sourceChannel == nil {
		return errors.New("no channel open")
	}

	// Create the packet data
	packetData := types.ConfirmUpgradePacketData{
		ChainId: sdkCtx.ChainID(),
		Height:  uint64(sdkCtx.BlockHeight()),
		Plan:    plan.Name,
	}
	err = packetData.ValidateBasic()
	if err != nil {
		return err
	}

	// Timeout
	connEnd, found := k.connectionKeeper.GetConnection(sdkCtx, ccvConnectionID)
	if !found {
		return fmt.Errorf("connection %s not found", ccvConnectionID)
	}
	clientState, ex := k.clientKeeper.GetClientState(sdkCtx, connEnd.ClientId)
	if !ex {
		return fmt.Errorf("client state missing for client ID '%s'", connEnd.ClientId)
	}
	p := k.GetParams(sdkCtx)
	timeoutHeight := clienttypes.Height{
		RevisionNumber: clientState.GetLatestHeight().GetRevisionNumber(),
		RevisionHeight: clientState.GetLatestHeight().GetRevisionHeight() + p.TimeoutHeight,
	}
	timeoutTimestamp := uint64(sdkCtx.BlockTime().Add(p.TimeoutTime).UnixNano())

	_, err = k.TransmitConfirmUpgradePacket(sdkCtx, packetData, types.PortID, sourceChannel.ChannelId, timeoutHeight, timeoutTimestamp)
	if err != nil {
		return err
	}
	k.Logger(sdkCtx).Info("sent IBC message about reaching the upgrade height for the current plan")
	return nil
}

// ScheduleUpgrade schedules an upgrade based on the specified plan.
// If there is another Plan already scheduled, it will cancel and overwrite it.
// ScheduleUpgrade will also write the upgraded IBC ClientState to the upgraded client
// path if it is specified in the plan.
func (k Keeper) ScheduleUpgrade(ctx context.Context, plan upgradetypes.Plan) error {
	if err := plan.ValidateBasic(); err != nil {
		return err
	}

	// NOTE: allow for the possibility of chains to schedule upgrades in begin block of the same block
	// as a strategy for emergency hard fork recoveries
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if plan.Height < sdkCtx.HeaderInfo().Height {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
	}

	doneHeight, err := k.upgradeKeeper.GetDoneHeight(ctx, plan.Name)
	if err != nil {
		return err
	}

	if doneHeight != 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "upgrade with name %s has already been completed", plan.Name)
	}

	//store := k.storeService.OpenKVStore(ctx)
	store := sdkCtx.KVStore(storetypes.NewKVStoreKey(upgradetypes.StoreKey))

	// clear any old IBC state stored by previous plan
	oldPlan, err := k.upgradeKeeper.GetUpgradePlan(ctx)
	// if there's an error but it's not ErrNoUpgradePlanFound, return error
	if err != nil && !errors.Is(err, upgradetypes.ErrNoUpgradePlanFound) {
		return err
	}

	if err == nil {
		err = k.ClearIBCState(ctx, oldPlan.Height)
		if err != nil {
			return err
		}
	}

	bz, err := k.cdc.Marshal(&plan)
	if err != nil {
		return err
	}

	store.Set(upgradetypes.PlanKey(), bz)
	/*err = store.Set(upgradetypes.PlanKey(), bz)
	if err != nil {
		return err
	}*/

	return nil
}

// ClearIBCState clears any planned IBC state
func (k Keeper) ClearIBCState(ctx context.Context, lastHeight int64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// delete IBC client and consensus state from store if this is IBC plan
	//store := k.storeService.OpenKVStore(ctx)
	store := sdkCtx.KVStore(storetypes.NewKVStoreKey(upgradetypes.StoreKey))
	store.Delete(upgradetypes.UpgradedClientKey(lastHeight))
	/*err := store.Delete(upgradetypes.UpgradedClientKey(lastHeight))
	if err != nil {
		return err
	}

	return store.Delete(upgradetypes.UpgradedConsStateKey(lastHeight))*/
	store.Delete(upgradetypes.UpgradedConsStateKey(lastHeight))
	return nil
}
