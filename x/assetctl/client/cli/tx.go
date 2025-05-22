package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// "github.com/cosmos/cosmos-sdk/client/tx"
	// "x/assetctl/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assetctl", // types.ModuleName
		Short:                      fmt.Sprintf("%s transactions subcommands", "assetctl" /*types.ModuleName*/),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// cmd.AddCommand(CmdRegisterAsset())
	// cmd.AddCommand(CmdSupportAsset())
	// this line is used by SdkApp scaffolding # 1

	return cmd
}
