package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	controllercli "github.com/sagaxyz/saga-sdk/x/assetctl/controller/client/cli"
	hostcli "github.com/sagaxyz/saga-sdk/x/assetctl/host/client/cli"
	// "x/assetctl/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assetctl",
		Short:                      fmt.Sprintf("%s transactions subcommands", "assetctl"),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		controllercli.GetTxCmd(),
		hostcli.GetTxCmd(),
	)

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
