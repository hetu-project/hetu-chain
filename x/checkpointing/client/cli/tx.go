package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/hetu-project/hetu/v1/crypto/bls12381"
	hetutypes "github.com/hetu-project/hetu/v1/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		CmdRegistBLSValidator(),
	)

	return txCmd
}

// CmdRegistBLSValidator returns the command to regist a validator with a BLS key
func CmdRegistBLSValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regist-validator bls_pub_key validator_address",
		Short: "regist a new validator with a BLS key",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`regist-validator will regist a new validator information with a BLS key.
The BLS key could be generated before running the command (e.g., via hetud create-bls-key).
The validator address should be a valid eth hexstring address.
Notice, The validator needs staking on contract of the checkpoint network.`),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		valAddr := args[1]
		if err := hetutypes.ValidateAddress(valAddr); err != nil {
			return fmt.Errorf("invalid eth account address %w", err)
		}

		blsPubKey, err := bls12381.NewBlsPubKeyFromHex(args[0])
		if err != nil {
			return fmt.Errorf("invalid BLS public key %w", err)
		}

		msg := &types.MsgRegistValidator{
			ValidatorAddress: valAddr,
			BlsPubkey:        &blsPubKey,
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
