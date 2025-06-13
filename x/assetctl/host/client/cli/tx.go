package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/sagaxyz/saga-sdk/x/assetctl/host/types"
)

// GetTxCmd returns the transaction commands for the host module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "host",
		Short:                      "Host transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetRegisterDenomsCmd(),
		GetSupportAssetsCmd(),
		GetUpdateParamsCmd(),
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetRegisterDenomsCmd returns the command to register denoms
func GetRegisterDenomsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-denoms [ibc-denoms]",
		Short: "Register denoms in the host module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ibcDenoms := args[0]

			var denoms []string
			if err := json.Unmarshal([]byte(ibcDenoms), &denoms); err != nil {
				return fmt.Errorf("failed to unmarshal ibc denoms: %w", err)
			}

			msg := &types.MsgRegisterDenoms{
				Authority: clientCtx.GetFromAddress().String(),
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
		Use:   "support-assets [ibc-denoms]",
		Short: "Support assets in the host module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ibcDenoms := args[0]

			var denoms []string
			if err := json.Unmarshal([]byte(ibcDenoms), &denoms); err != nil {
				return fmt.Errorf("failed to unmarshal ibc denoms: %w", err)
			}

			msg := &types.MsgSupportAssets{
				Authority: clientCtx.GetFromAddress().String(),
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
		Short: "Update the host module parameters",
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
