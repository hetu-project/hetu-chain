package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/hetu-project/hetu/v1/x/subnet/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQuerySubnets(),
		GetCmdQuerySubnet(),
		GetCmdQuerySubnetNeurons(),
		GetCmdQuerySubnetPool(),
	)

	return cmd
}

// GetCmdQuerySubnets implements the query subnets command.
func GetCmdQuerySubnets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnets",
		Short: "Query all subnets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QuerySubnetsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.Subnets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "subnets")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnet implements the query subnet command.
func GetCmdQuerySubnet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet [netuid]",
		Short: "Query a subnet by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			netuid, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySubnetRequest{
				Netuid: uint32(netuid),
			}

			res, err := queryClient.Subnet(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnetNeurons implements the query subnet neurons command.
func GetCmdQuerySubnetNeurons() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "neurons [netuid]",
		Short: "Query all neurons in a subnet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			netuid, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QuerySubnetNeuronsRequest{
				Netuid:     uint32(netuid),
				Pagination: pageReq,
			}

			res, err := queryClient.SubnetNeurons(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "neurons")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQuerySubnetPool implements the query subnet pool command.
func GetCmdQuerySubnetPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [netuid]",
		Short: "Query pool info for a subnet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			netuid, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySubnetPoolRequest{
				Netuid: uint32(netuid),
			}

			res, err := queryClient.SubnetPool(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
