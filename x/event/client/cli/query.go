package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "subnet",
		Short:                      "Querying commands for the subnet",
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

			// Create a REST client
			route := fmt.Sprintf("custom/event/subnets")
			res, _, err := clientCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(string(res))
		},
	}

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

			netuid, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			// Create a REST client
			route := fmt.Sprintf("custom/event/subnet/%d", netuid)
			res, _, err := clientCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(string(res))
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

			netuid, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			// Create a REST client
			route := fmt.Sprintf("custom/event/subnet/%d/neurons", netuid)
			res, _, err := clientCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(string(res))
		},
	}

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

			netuid, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			// Create a REST client
			route := fmt.Sprintf("custom/event/subnet/%d/pool", netuid)
			res, _, err := clientCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(string(res))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
