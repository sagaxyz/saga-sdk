package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"fmt"
	"os"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/sagaxyz/saga-sdk/x/admin/types"
)

// NewTxCmd returns a root CLI command handler for admin transaction commands
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "admin subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSetMetadataCmd(),
		NewEnableSetMetadataCmd(),
		NewDisableSetMetadataCmd(),
	)

	return txCmd
}

func NewSetMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-metadata [metadata-file | metadata-json]",
		Short: "Set denom token metadata",
		Long: `Set denom token metadata. Metadata can be provided either as a JSON file or as a JSON string.
Example:
  # From a file
  $ simd tx admin set-metadata metadata.json
metadata.json:
{
  "description": "The native staking token of an arbitrary cosmos sdk chain.",
  "denom_units": [{ "denom": "stake", "exponent": 0, "aliases": ["stake"] }],
  "base": "stake",
  "display": "stake",
  "name": "stake",
  "symbol": "stake"
}
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var metadataBytes []byte
			input := args[0]

			if _, err := os.Stat(input); err == nil {
				metadataBytes, err = os.ReadFile(input)
				if err != nil {
					return fmt.Errorf("failed to read metadata file: %w", err)
				}
			} else {
				metadataBytes = []byte(input)
			}

			var metadata banktypes.Metadata
			if err := clientCtx.Codec.UnmarshalJSON(metadataBytes, &metadata); err != nil {
				return fmt.Errorf("failed to parse metadata JSON: %w", err)
			}

			msg := types.NewMsgSetMetadata(clientCtx.GetFromAddress().String(), metadata)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewEnableSetMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable-set-metadata",
		Short: "Enable setting metadata",
		Long:  "Enable acl admin permission for setting denom metadata.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgEnableSetMetadata(clientCtx.GetFromAddress().String())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewDisableSetMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-set-metadata",
		Short: "Disable setting metadata",
		Long:  "Disable acl admin permission for setting denom metadata.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDisableSetMetadata(clientCtx.GetFromAddress().String())

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
