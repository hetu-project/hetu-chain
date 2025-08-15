package types

import (
	stdmath "math"
	"strconv"

	"cosmossdk.io/math"
)

// CalculateSubnetRewardRatio calculates the subnet reward ratio based on the formula:
// subnet_reward_ratio = min( max_ratio, base + k * log(1 + subnet_count) )
func CalculateSubnetRewardRatio(params Params, subnetCount uint64) math.LegacyDec {
	if subnetCount == 0 {
		return math.LegacyZeroDec()
	}

	// Calculate log(1 + subnet_count) with better numerical stability and convert to LegacyDec
	logFloat := stdmath.Log1p(float64(subnetCount))
	logStr := strconv.FormatFloat(logFloat, 'f', 18, 64) // 18 fractional digits to match LegacyDec precision
	logValue, err := math.LegacyNewDecFromStr(logStr)
	if err != nil {
		// Fallback: reformat with 'g' and attempt parse; if it still fails, degrade to 0 (ratio => base)
		alt := strconv.FormatFloat(logFloat, 'g', -1, 64)
		if v2, err2 := math.LegacyNewDecFromStr(alt); err2 == nil {
			logValue = v2
		} else {
			logValue = math.LegacyZeroDec()
		}
	}

	// Calculate base + k * log(1 + subnet_count)
	ratio := params.SubnetRewardBase.Add(params.SubnetRewardK.Mul(logValue))

	// Return min(ratio, max_ratio)
	if ratio.GT(params.SubnetRewardMaxRatio) {
		return params.SubnetRewardMaxRatio
	}

	return ratio
}
