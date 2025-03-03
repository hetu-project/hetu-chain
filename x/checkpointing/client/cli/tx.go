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
	"github.com/hetu-project/hetu/v1/utils"
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
		Use:   "regist-validator validator_address bls_pub_key",
		Short: "regist a new validator with a BLS key",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(`regist-validator will regist a new validator information with a BLS key.
The BLS key could be generated before running the command (e.g., via hetud keys create-bls-key).
The validator address should be a valid eth hexstring address.
Notice, The validator needs staking on contract of the checkpoint network.`),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		valAddr := args[0]
		if err := hetutypes.ValidateAddress(valAddr); err != nil {
			return fmt.Errorf("invalid eth account address %w", err)
		}

		var blspublic string
		if utils.Has0xPrefix(args[1]) {
			blspublic = args[1][2:]
		} else {
			blspublic = args[1]
		}
		blsPubKey, err := bls12381.NewBlsPubKeyFromHex(blspublic)
		if err != nil {
			return fmt.Errorf("invalid BLS public key %w", err)
		}

		sender := clientCtx.GetFromAddress()
		msg := &types.MsgRegistValidator{
			ValidatorAddress: valAddr,
			BlsPubkey:        &blsPubKey,
			Sender: 		 sender.String(),
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
