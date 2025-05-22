package cli

import (
	"fmt"
	// "context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// "x/assetctl/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assetctl", // types.ModuleName
		Short:                      fmt.Sprintf("Querying commands for the %s module", "assetctl" /*types.ModuleName*/),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// cmd.AddCommand(CmdAssetDirectory())
	// this line is used by SdkApp scaffolding # 1

	return cmd
}
