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
		// Fallback to a less precise but safe conversion if string parsing fails
		logValue = math.LegacyNewDecWithPrec(int64(logFloat*1000000000000000000), 18)
	}

	// Calculate base + k * log(1 + subnet_count)
	ratio := params.SubnetRewardBase.Add(params.SubnetRewardK.Mul(logValue))

	// Return min(ratio, max_ratio)
	if ratio.GT(params.SubnetRewardMaxRatio) {
		return params.SubnetRewardMaxRatio
	}

	return ratio
}
