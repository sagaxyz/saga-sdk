package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
)

const (
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	listSeparator              = ","
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		NewUpgradeChainCmd(),
	)

	// this line is used by starport scaffolding # 1

	return cmd
}

func NewUpgradeChainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade-chain [chainlet-stack-version] [height] [channel-id]",
		Short: "Broadcast message send-upgrade",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			version := args[0]
			_ = version

			height := args[1]
			channelID := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			h, err := strconv.ParseUint(height, 10, 64)
			if err != nil {
				return
			}
			msg := types.NewMsgSendUpgrade(
				clientCtx.GetFromAddress().String(),
				types.PortID,
				channelID,
				uint64(time.Now().Add(1*time.Hour).UnixNano()),
				h,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
