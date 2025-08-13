package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Verify total supply: 21 million HETU = 21,000,000 * 10^18 aHETU
	expectedTotalSupply, _ := math.NewIntFromString("21000000000000000000000000")
	require.Equal(t, expectedTotalSupply, params.TotalSupply, "Total supply should be 21 million HETU")

	// Verify default block reward: 1 HETU = 10^18 aHETU
	expectedBlockEmission, _ := math.NewIntFromString("1000000000000000000")
	require.Equal(t, expectedBlockEmission, params.DefaultBlockEmission, "Default block reward should be 1 HETU")

	// Verify other parameters
	require.Equal(t, "ahetu", params.MintDenom, "Token denomination should be ahetu")
	require.True(t, params.EnableBlockInflation, "Block inflation should be enabled by default")

	t.Logf("Total supply: %s aHETU", params.TotalSupply.String())
	t.Logf("Default block reward: %s aHETU", params.DefaultBlockEmission.String())

	// Verify numerical correctness
	require.Equal(t, "21000000000000000000000000", params.TotalSupply.String(), "Total supply string representation")
	require.Equal(t, "1000000000000000000", params.DefaultBlockEmission.String(), "Default block emission string representation")
}
