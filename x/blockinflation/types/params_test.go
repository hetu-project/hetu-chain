package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Verify total supply: 21 million HETU = 21,000,000 * 10^18 aHETU
	expectedTotalSupply, ok := math.NewIntFromString("21000000000000000000000000")
	require.True(t, ok, "failed to parse expected total supply")
	require.True(t, params.TotalSupply.Equal(expectedTotalSupply), "Total supply should be 21 million HETU")

	// Verify default block reward: 1 HETU = 10^18 aHETU
	expectedBlockEmission, ok := math.NewIntFromString("1000000000000000000")
	require.True(t, ok, "failed to parse expected block emission")
	require.True(t, params.DefaultBlockEmission.Equal(expectedBlockEmission), "Default block reward should be 1 HETU")

	// Verify other parameters
	require.Equal(t, "ahetu", params.MintDenom, "Token denomination should be ahetu")
	require.True(t, params.EnableBlockInflation, "Block inflation should be enabled by default")

	// Verify subnet reward parameters
	require.True(t, params.SubnetRewardBase.Equal(math.LegacyNewDecWithPrec(10, 2)), "Subnet reward base should be 0.10")
	require.True(t, params.SubnetRewardK.Equal(math.LegacyNewDecWithPrec(10, 2)), "Subnet reward K should be 0.10")
	require.True(t, params.SubnetRewardMaxRatio.Equal(math.LegacyNewDecWithPrec(90, 2)), "Subnet reward max ratio should be 0.90")
	require.True(t, params.SubnetMovingAlpha.Equal(math.LegacyNewDecWithPrec(3, 6)), "Subnet moving alpha should be 0.000003")
	require.True(t, params.SubnetOwnerCut.Equal(math.LegacyNewDecWithPrec(18, 2)), "Subnet owner cut should be 0.18")

	t.Logf("Total supply: %s aHETU", params.TotalSupply.String())
	t.Logf("Default block reward: %s aHETU", params.DefaultBlockEmission.String())

	// Verify numerical correctness
	require.Equal(t, "21000000000000000000000000", params.TotalSupply.String(), "Total supply string representation")
	require.Equal(t, "1000000000000000000", params.DefaultBlockEmission.String(), "Default block emission string representation")
}

func TestParamsValidation(t *testing.T) {
	// Test case where SubnetRewardBase > SubnetRewardMaxRatio
	invalidParams := DefaultParams()
	invalidParams.SubnetRewardBase = math.LegacyNewDecWithPrec(95, 2)     // 0.95
	invalidParams.SubnetRewardMaxRatio = math.LegacyNewDecWithPrec(90, 2) // 0.90

	err := invalidParams.Validate()
	require.Error(t, err, "Validation should fail when SubnetRewardBase > SubnetRewardMaxRatio")
	require.Contains(t, err.Error(), "subnet reward base", "Error message should mention subnet reward base")
	require.Contains(t, err.Error(), "cannot exceed max ratio", "Error message should explain the constraint")
}
