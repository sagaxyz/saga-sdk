package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/sagaxyz/saga-sdk/x/dac/types"
)

// GetQueryCmd returns the parent command for all dac CLI query commands
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the dac module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		ListAllowedCmd(),
		ListAdminsCmd(),
	)
	return cmd
}

// ListAllowed queries all allowed addresses
func ListAllowedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-allowed",
		Short: "Gets allowed addresses",
		Long:  "Gets allowed addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryListAllowedRequest{}

			res, err := queryClient.ListAllowed(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ListAdmins queries all admin addresses
func ListAdminsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-admins",
		Short: "Gets admin addresses",
		Long:  "Gets admin addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryListAdminsRequest{}

			res, err := queryClient.ListAdmins(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
