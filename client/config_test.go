package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cosmos/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func TestInitConfigNonNotExistError(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "nonPerms")
	if err := os.Mkdir(subDir, 0o600); err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}
	cmd := &cobra.Command{}
	cmd.PersistentFlags().String(flags.FlagHome, "", "")
	if err := cmd.PersistentFlags().Set(flags.FlagHome, subDir); err != nil {
		t.Fatalf("Could not set home flag [%T] %v", err, err)
	}

	if err := InitConfig(cmd); !os.IsPermission(err) {
		t.Fatalf("Failed to catch permissions error, got: [%T] %v", err, err)
	}
}

func TestDecodeBen32AndReplace(t *testing.T) {
	originalAddress := "hhub15cvq3ljql6utxseh0zau9m8ve2j8erz8jplma4"
	newPrefix := "hetu"

	hrp, decoded, err := bech32.Decode(originalAddress, 64)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}

	if strings.HasPrefix(hrp, "hhub") {
		hrp = newPrefix + hrp[len("hhub"):]

		newAddress, err := bech32.Encode(hrp, decoded)
		if err != nil {
			fmt.Println("encode error:", err)
			return
		}

		fmt.Println("After:", newAddress)
	} else {
		fmt.Println("Not 'hhub' begin, no needs")
	}
}
