package cli

import (
	"fmt"
	"strconv"

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
		GetCmdQuerySubnetPrice(),
		GetCmdQueryPendingEmission(),
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
				fmt.Sprintf("Subnet Reward Max Ratio: %s\n", res.Params.SubnetRewardMaxRatio.String()) +
				fmt.Sprintf("Subnet Moving Alpha: %s\n", res.Params.SubnetMovingAlpha.String()) +
				fmt.Sprintf("Subnet Owner Cut: %s\n", res.Params.SubnetOwnerCut.String()))
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
			return clientCtx.PrintString(
				fmt.Sprintf("Subnet Reward Base: %s\n", res.Params.SubnetRewardBase.String()) +
					fmt.Sprintf("Subnet Reward K: %s\n", res.Params.SubnetRewardK.String()) +
					fmt.Sprintf("Subnet Reward Max Ratio: %s\n", res.Params.SubnetRewardMaxRatio.String()),
			)
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
			u, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid (must be 0-65535): %s", args[0])
			}
			netuid := uint16(u)

			// For now, just print a message since we don't have gRPC service implemented
			fmt.Printf("Subnet emission data query for netuid %d\n", netuid)
			fmt.Println("This will show calculated tao_in, alpha_in, alpha_out, owner_cut, and root_divs for the subnet")
			fmt.Println("The calculation uses the same logarithmic decay algorithm as block emission")
			fmt.Println("Step 4 (Injection) has been implemented:")
			fmt.Println("  - alpha_in is added to subnet Alpha in pool (for liquidity)")
			fmt.Println("  - alpha_out is added to subnet Alpha out pool (for distribution)")
			fmt.Println("  - tao_in is added to subnet TAO pool")
			fmt.Println("Step 5 (Owner Cuts) has been implemented:")
			fmt.Println("  - Owner cut = alpha_out × SubnetOwnerCut (default 18%)")
			fmt.Println("  - Owner cut is subtracted from subnet Alpha out pool")
			fmt.Println("  - Owner cut is added to pending owner cut pool for later distribution")
			fmt.Println("Also shows cumulative emission statistics:")
			fmt.Println("  - SubnetAlphaInEmission: Total Alpha in emission over time")
			fmt.Println("  - SubnetAlphaOutEmission: Total Alpha out emission over time")
			fmt.Println("  - SubnetTaoInEmission: Total TAO in emission over time")
			fmt.Println("Note: gRPC service not yet implemented - this is a placeholder")

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
			fmt.Println("All subnet emission data query")
			fmt.Println("This will show calculated emission data for all subnets that have first emission block set")
			fmt.Println("Each subnet's rewards are calculated based on:")
			fmt.Println("  - TAO reward: block_emission × (moving_price / total_moving_prices)")
			fmt.Println("  - Alpha emission: logarithmic decay based on subnet Alpha issuance")
			fmt.Println("  - Alpha in: min(tao_in / price, alpha_emission)")
			fmt.Println("  - Alpha out: alpha_emission")
			fmt.Println("Step 4 (Injection) has been implemented:")
			fmt.Println("  - alpha_in is added to subnet Alpha in pool (for liquidity)")
			fmt.Println("  - alpha_out is added to subnet Alpha out pool (for distribution)")
			fmt.Println("  - tao_in is added to subnet TAO pool")
			fmt.Println("  - Cumulative emission statistics are tracked")
			fmt.Println("Step 5 (Owner Cuts) has been implemented:")
			fmt.Println("  - Owner cut = alpha_out × SubnetOwnerCut (default 18%)")
			fmt.Println("  - Owner cut is subtracted from subnet Alpha out pool")
			fmt.Println("  - Owner cut is added to pending owner cut pool for later distribution")
			fmt.Println("Note: gRPC service not yet implemented - this is a placeholder")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnetPrice implements the subnet price query command.
func GetCmdQuerySubnetPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet-price [netuid]",
		Short: "Query price data for a specific subnet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse netuid
			u, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid (must be 0-65535): %s", args[0])
			}
			netuid := uint16(u)

			// For now, just print a message since we don't have gRPC service implemented
			fmt.Printf("Subnet price query for netuid %d - gRPC service not yet implemented\n", netuid)
			fmt.Println("This will show alpha_price, moving_price, subnet_tao, subnet_alpha_in, subnet_alpha_out, and volume")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPendingEmission implements the pending emission query command.
func GetCmdQueryPendingEmission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-emission [netuid]",
		Short: "Query the pending emission for a specific subnet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse netuid
			u, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid (must be 0-65535): %s", args[0])
			}
			netuid := uint16(u)

			// For now, just print a message since we don't have gRPC service implemented
			fmt.Printf("Pending emission query for netuid %d\n", netuid)
			fmt.Println("This will show the pending emission amount for the subnet")
			fmt.Println("Pending emission accumulates alpha_out from each block")
			fmt.Println("Step 6 (Pending Emission) has been implemented:")
			fmt.Println("  - pending_alpha = alpha_out_i")
			fmt.Println("  - Add pending_alpha to subnet's pending emission")
			fmt.Println("  - This is used for epoch-based emission distribution")
			fmt.Println("Note: gRPC service not yet implemented - this is a placeholder")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
