package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// EpochResult epoch calculation result
type EpochResult struct {
	Netuid    uint16      `json:"netuid"`
	Accounts  []string    `json:"accounts"`
	Emission  []uint64    `json:"emission"`
	Dividend  []uint64    `json:"dividend"`
	Incentive []uint64    `json:"incentive"` // New: incentive allocation
	Bonds     [][]float64 `json:"bonds"`
	Consensus []float64   `json:"consensus"`
}

// Validate checks if EpochResult has consistent lengths across all fields
func (e EpochResult) Validate() error {
	n := len(e.Accounts)
	checks := []struct {
		name string
		len  int
	}{
		{"emission", len(e.Emission)},
		{"dividend", len(e.Dividend)},
		{"incentive", len(e.Incentive)},
		{"consensus", len(e.Consensus)},
	}

	for _, c := range checks {
		if c.len != n {
			return fmt.Errorf("epoch result length mismatch: accounts=%d %s=%d", n, c.name, c.len)
		}
	}

	if len(e.Bonds) != n {
		return fmt.Errorf("epoch result bonds length mismatch: accounts=%d bonds=%d", n, len(e.Bonds))
	}

	return nil
}

// EpochParams epoch parameters (parsed from event module's Subnet.Params)
type EpochParams struct {
	// Core parameters
	Kappa float64 `json:"kappa"` // Majority threshold (0.5)
	Alpha float64 `json:"alpha"` // EMA parameter (0.1-0.9)
	Delta float64 `json:"delta"` // Weight clipping range (1.0)

	// Activity parameters
	ActivityCutoff uint64 `json:"activity_cutoff"` // Activity cutoff time
	ImmunityPeriod uint64 `json:"immunity_period"` // Immunity period

	// Weight parameters
	MaxWeightsLimit     uint64 `json:"max_weights_limit"`      // Maximum weight count
	MinAllowedWeights   uint64 `json:"min_allowed_weights"`    // Minimum weight count
	WeightsSetRateLimit uint64 `json:"weights_set_rate_limit"` // Weight setting rate limit

	// Other parameters
	Tempo              uint64  `json:"tempo"`                // Epoch run frequency
	BondsPenalty       float64 `json:"bonds_penalty"`        // Bonds penalty
	BondsMovingAverage float64 `json:"bonds_moving_average"` // Bonds moving average

	// New parameters
	Rho                   float64 `json:"rho"`                     // Incentive parameter
	LiquidAlphaEnabled    bool    `json:"liquid_alpha_enabled"`    // Whether to enable dynamic alpha
	AlphaSigmoidSteepness float64 `json:"alpha_sigmoid_steepness"` // Alpha sigmoid steepness
	AlphaLow              float64 `json:"alpha_low"`               // Alpha lower bound
	AlphaHigh             float64 `json:"alpha_high"`              // Alpha upper bound
}

// Validate checks if EpochParams are within valid ranges
func (p EpochParams) Validate() error {
	if p.Kappa < 0 || p.Kappa > 1 {
		return fmt.Errorf("kappa must be between 0 and 1, got %f", p.Kappa)
	}
	if p.Alpha < 0 || p.Alpha > 1 {
		return fmt.Errorf("alpha must be between 0 and 1, got %f", p.Alpha)
	}
	if p.Delta < 0 {
		return fmt.Errorf("delta must be non-negative, got %f", p.Delta)
	}
	if p.Rho < 0 || p.Rho > 1 {
		return fmt.Errorf("rho must be between 0 and 1, got %f", p.Rho)
	}
	if p.AlphaLow >= p.AlphaHigh {
		return fmt.Errorf("alpha_low must be less than alpha_high")
	}
	if p.MinAllowedWeights > p.MaxWeightsLimit {
		return fmt.Errorf("min_allowed_weights cannot exceed max_weights_limit")
	}
	if p.BondsPenalty < 0 || p.BondsPenalty > 1 {
		return fmt.Errorf("bonds_penalty must be between 0 and 1, got %f", p.BondsPenalty)
	}
	if p.BondsMovingAverage < 0 || p.BondsMovingAverage > 1 {
		return fmt.Errorf("bonds_moving_average must be between 0 and 1, got %f", p.BondsMovingAverage)
	}
	if p.AlphaSigmoidSteepness <= 0 {
		return fmt.Errorf("alpha_sigmoid_steepness must be positive, got %f", p.AlphaSigmoidSteepness)
	}
	if p.Tempo == 0 {
		return fmt.Errorf("tempo must be greater than 0")
	}
	return nil
}

// clamp01 clamps a float value to the range [0, 1]
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

// DefaultEpochParams default parameters
func DefaultEpochParams() EpochParams {
	return EpochParams{
		Kappa:                 0.5,
		Alpha:                 0.1,
		Delta:                 1.0,
		ActivityCutoff:        5000,
		ImmunityPeriod:        4096,
		MaxWeightsLimit:       1000,
		MinAllowedWeights:     8,
		WeightsSetRateLimit:   100,
		Tempo:                 100,
		BondsPenalty:          0.1,
		BondsMovingAverage:    0.9,
		Rho:                   0.5,
		LiquidAlphaEnabled:    false,
		AlphaSigmoidSteepness: 10.0,
		AlphaLow:              0.01,
		AlphaHigh:             0.99,
	}
}

// ParseEpochParams parses parameters from event module's Subnet.Params
func ParseEpochParams(paramMap map[string]string) (EpochParams, error) {
	params := DefaultEpochParams()
	var parseErrors []string

	// Parse parameters (if they exist)
	if val, exists := paramMap["kappa"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Kappa = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("kappa: %v", err))
		}
	}

	if val, exists := paramMap["alpha"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Alpha = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("alpha: %v", err))
		}
	}

	if val, exists := paramMap["delta"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 {
			params.Delta = f
		} else if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("delta: %v", err))
		} else {
			parseErrors = append(parseErrors, "delta must be positive")
		}
	}

	if val, exists := paramMap["activity_cutoff"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.ActivityCutoff = f
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("activity_cutoff: %v", err))
		}
	}

	if val, exists := paramMap["immunity_period"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.ImmunityPeriod = f
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("immunity_period: %v", err))
		}
	}

	if val, exists := paramMap["max_weights_limit"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.MaxWeightsLimit = f
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("max_weights_limit: %v", err))
		}
	}

	if val, exists := paramMap["min_allowed_weights"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.MinAllowedWeights = f
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("min_allowed_weights: %v", err))
		}
	}

	if val, exists := paramMap["weights_set_rate_limit"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.WeightsSetRateLimit = f
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("weights_set_rate_limit: %v", err))
		}
	}

	if val, exists := paramMap["tempo"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil && f > 0 {
			params.Tempo = f
		} else if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("tempo: %v", err))
		} else {
			parseErrors = append(parseErrors, "tempo must be positive")
		}
	}

	if val, exists := paramMap["bonds_penalty"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.BondsPenalty = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("bonds_penalty: %v", err))
		}
	}

	if val, exists := paramMap["bonds_moving_average"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.BondsMovingAverage = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("bonds_moving_average: %v", err))
		}
	}

	// Parse new parameters
	if val, exists := paramMap["rho"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Rho = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("rho: %v", err))
		}
	}

	if val, exists := paramMap["liquid_alpha_enabled"]; exists {
		if b, err := strconv.ParseBool(val); err == nil {
			params.LiquidAlphaEnabled = b
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("liquid_alpha_enabled: %v", err))
		}
	}

	if val, exists := paramMap["alpha_sigmoid_steepness"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 {
			params.AlphaSigmoidSteepness = f
		} else if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("alpha_sigmoid_steepness: %v", err))
		} else {
			parseErrors = append(parseErrors, "alpha_sigmoid_steepness must be positive")
		}
	}

	if val, exists := paramMap["alpha_low"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.AlphaLow = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("alpha_low: %v", err))
		}
	}

	if val, exists := paramMap["alpha_high"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.AlphaHigh = clamp01(f)
		} else {
			parseErrors = append(parseErrors, fmt.Sprintf("alpha_high: %v", err))
		}
	}

	// Validate the final parameters
	if err := params.Validate(); err != nil {
		parseErrors = append(parseErrors, fmt.Sprintf("validation: %v", err))
	}

	if len(parseErrors) > 0 {
		return params, fmt.Errorf("parameter parsing errors: %s", strings.Join(parseErrors, "; "))
	}

	return params, nil
}

// ValidatorInfo validator information (obtained from event module)
type ValidatorInfo struct {
	Address string   `json:"address"`
	Stake   string   `json:"stake"` // bigint string to avoid precision loss
	Weights []uint64 `json:"weights"`
	Active  bool     `json:"active"`
}

// GetStakeBigInt returns the Stake as *big.Int
func (v ValidatorInfo) GetStakeBigInt() (*big.Int, error) {
	if v.Stake == "" {
		return big.NewInt(0), nil
	}

	amount := new(big.Int)
	if _, ok := amount.SetString(v.Stake, 10); !ok {
		return nil, fmt.Errorf("invalid stake amount: %s", v.Stake)
	}
	return amount, nil
}

// SubnetEpochData subnet epoch data
type SubnetEpochData struct {
	Netuid     uint16          `json:"netuid"`
	Validators []ValidatorInfo `json:"validators"`
	Params     EpochParams     `json:"params"`
	Emission   uint64          `json:"emission"`
}
