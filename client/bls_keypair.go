package client

import (
	"encoding/hex"
	"fmt"
	"strings"

	blst "github.com/supranational/blst/bindings/go"

	"github.com/hetu-project/hetu/v1/crypto/bls12381"
	"github.com/spf13/cobra"
)

func CreateBlsKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-bls-key",
		Short: "Create a pair of BLS keys for a validator of the checkpointing network",
		Long: strings.TrimSpace(`create-bls will create a pair of BLS keys that are used to
send BLS signatures for checkpointing network.

BLS keys that new created just are displayed once, please keep it carefully.

Example:
$ hetud keys create-bls-key`,
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			skSerialized := bls12381.GenPrivKey()
			sk := new(blst.SecretKey)
			sk.Deserialize(skSerialized)
			pk := new(bls12381.BlsPubKey).From(sk)
			pkSer := pk.Serialize()
			privKey := hex.EncodeToString(skSerialized)
			pubKey := hex.EncodeToString(pkSer)

			fmt.Printf("Private Key: 0x%x\n", privKey)
			fmt.Printf("Public Key: 0x%x\n", pubKey)
			return nil
		},
	}

	return cmd
}
