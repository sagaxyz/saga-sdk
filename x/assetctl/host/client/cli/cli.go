package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
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
	// Add transaction commands here
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetQueryCmd returns the query commands for the host module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "host",
		Short:                      "Host query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
	// Add query commands here
	)

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
