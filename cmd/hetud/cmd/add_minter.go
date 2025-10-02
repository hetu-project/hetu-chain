package cmd

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// AddMinterCmd returns add-minter cobra Command.
func AddMinterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-minter [netuid] [subnet-manager-address]",
		Short: "Add blockinflation module as authorized minter for a subnet",
		Long: `Add blockinflation module as authorized minter for a subnet.
This command helps subnet owners to authorize the blockinflation module to mint Alpha tokens.

Example:
$ hetud tx add-minter 1 0x6bF0ECa02A91Ffe3260e3104CF449CCaa1CedbE0 --from=<key-name>
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			_ = clientCtx

			// Parse netuid
			netuid, err := strconv.ParseUint(args[0], 10, 16)
			if err != nil {
				return fmt.Errorf("invalid netuid: %w", err)
			}

			// Parse subnet manager address
			subnetManagerAddress := args[1]
			if !common.IsHexAddress(subnetManagerAddress) {
				return fmt.Errorf("invalid subnet manager address: %s", subnetManagerAddress)
			}

			// Get blockinflation module address
			moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
			moduleHexAddress := common.BytesToAddress(moduleAddress.Bytes()).Hex()

			// Print information
			fmt.Printf("Adding blockinflation module as authorized minter:\n")
			fmt.Printf("Subnet ID: %d\n", netuid)
			fmt.Printf("Subnet Manager: %s\n", subnetManagerAddress)
			fmt.Printf("Module Address: %s\n", moduleHexAddress)

			// TODO: Implement the actual transaction
			// This would require implementing a custom transaction type
			// that calls the SubnetManager contract's addSubnetMinter method
			fmt.Printf("\nThis command is not fully implemented yet.\n")
			fmt.Printf("Please manually call addSubnetMinter(%d, %s) on the SubnetManager contract at %s\n",
				netuid, moduleHexAddress, subnetManagerAddress)

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
