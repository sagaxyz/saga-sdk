package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/sagaxyz/saga-sdk/x/assetctl/controller/types"
)

// GetTxCmd returns the transaction commands for the controller module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "controller",
		Short:                      "Controller transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetManageAssetsCmd(),
		GetSupportAssetsCmd(),
		GetUpdateParamsCmd(),
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetManageAssetsCmd returns the command to manage assets (register and/or unregister)
func GetManageAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage-assets [channel-id] [assets-to-register] [assets-to-unregister]",
		Short: "Manage assets in the controller module (register and/or unregister)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			assetsToRegisterJSON := args[1]
			assetsToUnregisterJSON := args[2]

			var assetsToRegister []banktypes.Metadata
			if err := json.Unmarshal([]byte(assetsToRegisterJSON), &assetsToRegister); err != nil {
				return fmt.Errorf("failed to unmarshal assets to register: %w", err)
			}

			var assetsToUnregister []string
			if err := json.Unmarshal([]byte(assetsToUnregisterJSON), &assetsToUnregister); err != nil {
				return fmt.Errorf("failed to unmarshal assets to unregister: %w", err)
			}

			msg := &types.MsgManageRegisteredAssets{
				Authority:          clientCtx.GetFromAddress().String(),
				ChannelId:          channelID,
				AssetsToRegister:   assetsToRegister,
				AssetsToUnregister: assetsToUnregister,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetSupportAssetsCmd returns the command to manage asset support
func GetSupportAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage-support [channel-id] [add-denoms] [remove-denoms]",
		Short: "Manage asset support in the controller module (add and/or remove support)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			addDenomsJSON := args[1]
			removeDenomsJSON := args[2]

			var addDenoms, removeDenoms []string
			if err := json.Unmarshal([]byte(addDenomsJSON), &addDenoms); err != nil {
				return fmt.Errorf("failed to unmarshal add denoms: %w", err)
			}
			if err := json.Unmarshal([]byte(removeDenomsJSON), &removeDenoms); err != nil {
				return fmt.Errorf("failed to unmarshal remove denoms: %w", err)
			}

			msg := &types.MsgManageSupportedAssets{
				Authority:       clientCtx.GetFromAddress().String(),
				ChannelId:       channelID,
				AddIbcDenoms:    addDenoms,
				RemoveIbcDenoms: removeDenoms,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetUpdateParamsCmd returns the command to update module parameters
func GetUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [params-json]",
		Short: "Update the controller module parameters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			paramsJSON := args[0]

			var params types.Params
			if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}

			msg := &types.MsgUpdateParams{
				Authority: clientCtx.GetFromAddress().String(),
				Params:    &params,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
