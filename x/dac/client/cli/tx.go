package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sagaxyz/sagaevm/v8/x/dac/types"
)

// NewTxCmd returns a root CLI command handler for dac transaction commands
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "dac subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewAddAllowedCmd(),
		NewRemoveAllowedCmd(),
		NewAddAdminsCmd(),
		NewRemoveAdminsCmd(),
	)
	return txCmd
}

// NewAddAllowedCmd returns a CLI command handler for adding allowed addresses
func NewAddAllowedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-allowed <space-separated list of addresses>",
		Short: "Adds addresses to the allowed list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddAllowed(cliCtx.GetFromAddress(), args...)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewRemoveAllowedCmd returns a CLI command handler for adding allowed addresses
func NewRemoveAllowedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-allowed <space-separated list of addresses>",
		Short: "Removes addresses from the allowed list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveAllowed(cliCtx.GetFromAddress(), args...)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewAddAdminsCmd returns a CLI command handler for adding admin addresses
func NewAddAdminsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-admins <space-separated list of addresses>",
		Short: "Adds addresses to the admins list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddAdmins(cliCtx.GetFromAddress(), args...)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewRemoveAdminsCmd returns a CLI command handler for adding admin addresses
func NewRemoveAdminsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-admins <space-separated list of addresses>",
		Short: "Removes addresses from the admins list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveAdmins(cliCtx.GetFromAddress(), args...)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
