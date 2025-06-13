package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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
		GetRegisterAssetsCmd(),
		GetUnregisterAssetsCmd(),
		GetSupportAssetsCmd(),
		GetUpdateParamsCmd(),
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetRegisterAssetsCmd returns the command to register assets
func GetRegisterAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-assets [channel-id] [assets-json]",
		Short: "Register assets in the controller module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			assetsJSON := args[1]

			var assets []types.AssetDetails
			if err := json.Unmarshal([]byte(assetsJSON), &assets); err != nil {
				return fmt.Errorf("failed to unmarshal assets: %w", err)
			}

			msg := &types.MsgRegisterAssets{
				Authority:        clientCtx.GetFromAddress().String(),
				ChannelId:        channelID,
				AssetsToRegister: assets,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetUnregisterAssetsCmd returns the command to unregister assets
func GetUnregisterAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister-assets [channel-id] [ibc-denoms]",
		Short: "Unregister assets from the controller module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			ibcDenoms := args[1]

			var denoms []string
			if err := json.Unmarshal([]byte(ibcDenoms), &denoms); err != nil {
				return fmt.Errorf("failed to unmarshal ibc denoms: %w", err)
			}

			msg := &types.MsgUnregisterAssets{
				Authority: clientCtx.GetFromAddress().String(),
				ChannelId: channelID,
				IbcDenoms: denoms,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetSupportAssetsCmd returns the command to support assets
func GetSupportAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "support-assets [channel-id] [ibc-denoms]",
		Short: "Support assets in the controller module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]
			ibcDenoms := args[1]

			var denoms []string
			if err := json.Unmarshal([]byte(ibcDenoms), &denoms); err != nil {
				return fmt.Errorf("failed to unmarshal ibc denoms: %w", err)
			}

			msg := &types.MsgSupportAssets{
				Authority: clientCtx.GetFromAddress().String(),
				ChannelId: channelID,
				IbcDenoms: denoms,
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
