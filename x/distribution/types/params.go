package types

import (
	"fmt"
	stdmath "math"

	"cosmossdk.io/math"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyBaseSubnetRewardRatio = []byte("BaseSubnetRewardRatio")
	KeyMaxSubnetRewardRatio  = []byte("MaxSubnetRewardRatio")
	KeyGrowthRateCoefficient = []byte("GrowthRateCoefficient")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the distribution module.
type Params struct {
	// BaseSubnetRewardRatio is the initial subnet reward ratio (e.g., 0.10)
	BaseSubnetRewardRatio math.LegacyDec `json:"base_subnet_reward_ratio" yaml:"base_subnet_reward_ratio"`
	// MaxSubnetRewardRatio is the maximum subnet reward ratio (e.g., 0.50)
	MaxSubnetRewardRatio math.LegacyDec `json:"max_subnet_reward_ratio" yaml:"max_subnet_reward_ratio"`
	// GrowthRateCoefficient is the growth rate coefficient (e.g., 0.1)
	GrowthRateCoefficient math.LegacyDec `json:"growth_rate_coefficient" yaml:"growth_rate_coefficient"`
}

// NewParams creates a new Params instance
func NewParams(baseRatio, maxRatio, growthRate math.LegacyDec) Params {
	return Params{
		BaseSubnetRewardRatio: baseRatio,
		MaxSubnetRewardRatio:  maxRatio,
		GrowthRateCoefficient: growthRate,
	}
}

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return NewParams(
		math.LegacyNewDecWithPrec(10, 2), // 0.10 (10%)
		math.LegacyNewDecWithPrec(50, 2), // 0.50 (50%)
		math.LegacyNewDecWithPrec(1, 1),  // 0.1
	)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyBaseSubnetRewardRatio, &p.BaseSubnetRewardRatio, validateBaseSubnetRewardRatio),
		paramstypes.NewParamSetPair(KeyMaxSubnetRewardRatio, &p.MaxSubnetRewardRatio, validateMaxSubnetRewardRatio),
		paramstypes.NewParamSetPair(KeyGrowthRateCoefficient, &p.GrowthRateCoefficient, validateGrowthRateCoefficient),
	}
}

// Validate performs basic validation on distribution parameters.
func (p *Params) Validate() error {
	if err := validateBaseSubnetRewardRatio(p.BaseSubnetRewardRatio); err != nil {
		return err
	}
	if err := validateMaxSubnetRewardRatio(p.MaxSubnetRewardRatio); err != nil {
		return err
	}
	if err := validateGrowthRateCoefficient(p.GrowthRateCoefficient); err != nil {
		return err
	}
	if p.BaseSubnetRewardRatio.GT(p.MaxSubnetRewardRatio) {
		return fmt.Errorf("base subnet reward ratio cannot be greater than max subnet reward ratio")
	}
	return nil
}

// CalculateSubnetRewardRatio calculates the dynamic subnet reward ratio based on subnet count
func (p *Params) CalculateSubnetRewardRatio(subnetCount int64) math.LegacyDec {
	if subnetCount <= 0 {
		return math.LegacyZeroDec()
	}

	// Calculate: base + k * log(1 + subnet_count)
	// Use standard library math.Log for natural logarithm
	logValue := stdmath.Log(float64(1 + subnetCount))
	logTerm := math.LegacyNewDecWithPrec(int64(logValue*1000000), 6) // Convert to LegacyDec with 6 decimal places
	growthTerm := p.GrowthRateCoefficient.Mul(logTerm)
	calculatedRatio := p.BaseSubnetRewardRatio.Add(growthTerm)

	// Apply min/max bounds
	if calculatedRatio.LT(math.LegacyZeroDec()) {
		calculatedRatio = math.LegacyZeroDec()
	}
	if calculatedRatio.GT(p.MaxSubnetRewardRatio) {
		calculatedRatio = p.MaxSubnetRewardRatio
	}

	return calculatedRatio
}

func validateBaseSubnetRewardRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("base subnet reward ratio cannot be nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("base subnet reward ratio cannot be negative")
	}
	if v.GT(math.LegacyNewDec(1)) {
		return fmt.Errorf("base subnet reward ratio cannot be greater than 1")
	}

	return nil
}

func validateMaxSubnetRewardRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("max subnet reward ratio cannot be nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("max subnet reward ratio cannot be negative")
	}
	if v.GT(math.LegacyNewDec(1)) {
		return fmt.Errorf("max subnet reward ratio cannot be greater than 1")
	}

	return nil
}

func validateGrowthRateCoefficient(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("growth rate coefficient cannot be nil")
	}
	if v.IsNegative() {
		return fmt.Errorf("growth rate coefficient cannot be negative")
	}

	return nil
}
