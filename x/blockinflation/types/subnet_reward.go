package types

import (
	stdmath "math"

	"cosmossdk.io/math"
)

// CalculateSubnetRewardRatio calculates the subnet reward ratio based on the formula:
// subnet_reward_ratio = min( max_ratio, base + k * log(1 + subnet_count) )
func CalculateSubnetRewardRatio(params Params, subnetCount uint64) math.LegacyDec {
	if subnetCount == 0 {
		return math.LegacyZeroDec()
	}

	// Calculate log(1 + subnet_count) and convert to LegacyDec
	logFloat := stdmath.Log(float64(1 + subnetCount))
	logValue := math.LegacyNewDecFromInt(math.NewInt(int64(logFloat * 1000000))).Quo(math.LegacyNewDec(1000000))

	// Calculate base + k * log(1 + subnet_count)
	ratio := params.SubnetRewardBase.Add(params.SubnetRewardK.Mul(logValue))

	// Return min(ratio, max_ratio)
	if ratio.GT(params.SubnetRewardMaxRatio) {
		return params.SubnetRewardMaxRatio
	}

	return ratio
}
