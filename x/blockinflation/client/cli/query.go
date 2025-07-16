package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group blockinflation queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQuerySubnetRewardParams(),
		GetCmdQueryPendingSubnetRewards(),
	)

	return cmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current blockinflation parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnetRewardParams implements the subnet reward params query command.
func GetCmdQuerySubnetRewardParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet-reward-params",
		Short: "Query the current subnet reward parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			// Print subnet reward specific parameters
			fmt.Printf("Subnet Reward Base: %s\n", res.Params.SubnetRewardBase.String())
			fmt.Printf("Subnet Reward K: %s\n", res.Params.SubnetRewardK.String())
			fmt.Printf("Subnet Reward Max Ratio: %s\n", res.Params.SubnetRewardMaxRatio.String())

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPendingSubnetRewards implements the pending subnet rewards query command.
func GetCmdQueryPendingSubnetRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-subnet-rewards",
		Short: "Query the current pending subnet rewards",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PendingSubnetRewards(cmd.Context(), &types.QueryPendingSubnetRewardsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
