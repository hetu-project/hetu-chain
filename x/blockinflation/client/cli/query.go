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
		GetCmdQuerySubnetEmissionData(),
		GetCmdQueryAllSubnetEmissionData(),
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

			// Use PrintString instead of PrintProto since our types don't implement protobuf
			return clientCtx.PrintString(fmt.Sprintf("Enable Block Inflation: %t\n", res.Params.EnableBlockInflation) +
				fmt.Sprintf("Mint Denom: %s\n", res.Params.MintDenom) +
				fmt.Sprintf("Total Supply: %s\n", res.Params.TotalSupply.String()) +
				fmt.Sprintf("Default Block Emission: %s\n", res.Params.DefaultBlockEmission.String()) +
				fmt.Sprintf("Subnet Reward Base: %s\n", res.Params.SubnetRewardBase.String()) +
				fmt.Sprintf("Subnet Reward K: %s\n", res.Params.SubnetRewardK.String()) +
				fmt.Sprintf("Subnet Reward Max Ratio: %s\n", res.Params.SubnetRewardMaxRatio.String()))
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

			// Use PrintString instead of PrintProto since our types don't implement protobuf
			return clientCtx.PrintString(fmt.Sprintf("Pending Subnet Rewards: %s\n", res.PendingSubnetRewards))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnetEmissionData implements the subnet emission data query command.
func GetCmdQuerySubnetEmissionData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet-emission [netuid]",
		Short: "Query emission data for a specific subnet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse netuid
			var netuid uint16
			if _, err := fmt.Sscanf(args[0], "%d", &netuid); err != nil {
				return fmt.Errorf("invalid netuid: %s", args[0])
			}

			// For now, just print a message since we don't have gRPC service implemented
			fmt.Printf("Subnet emission data query for netuid %d - gRPC service not yet implemented\n", netuid)
			fmt.Println("This will show tao_in, alpha_in, alpha_out, owner_cut, and root_divs for the subnet")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAllSubnetEmissionData implements the all subnet emission data query command.
func GetCmdQueryAllSubnetEmissionData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-subnet-emission",
		Short: "Query emission data for all subnets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// For now, just print a message since we don't have gRPC service implemented
			fmt.Println("All subnet emission data query - gRPC service not yet implemented")
			fmt.Println("This will show emission data for all subnets that have first emission block set")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
