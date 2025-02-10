package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/hetu-project/hetu/v1/app"
	hetud "github.com/hetu-project/hetu/v1/cmd/hetud"
	"github.com/hetu-project/hetu/v1/utils"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := hetud.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",      // Test the init cmd
		"hetu-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, utils.TestnetChainID+"-1"),
	})

	err := svrcmd.Execute(rootCmd, "hetud", app.DefaultNodeHome)
	require.NoError(t, err)
}

func TestAddKeyLedgerCmd(t *testing.T) {
	rootCmd, _ := hetud.NewRootCmd()
	rootCmd.SetArgs([]string{
		"keys",
		"add",
		"dev0",
		fmt.Sprintf("--%s", flags.FlagUseLedger),
	})

	err := svrcmd.Execute(rootCmd, "HETUD", app.DefaultNodeHome)
	require.Error(t, err)
}
