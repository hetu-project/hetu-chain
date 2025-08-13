package types

import (
	"strconv"
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
func ParseEpochParams(paramMap map[string]string) EpochParams {
	params := DefaultEpochParams()

	// Parse parameters (if they exist)
	if val, exists := paramMap["kappa"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Kappa = f
		}
	}

	if val, exists := paramMap["alpha"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Alpha = f
		}
	}

	if val, exists := paramMap["delta"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Delta = f
		}
	}

	if val, exists := paramMap["activity_cutoff"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.ActivityCutoff = f
		}
	}

	if val, exists := paramMap["immunity_period"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.ImmunityPeriod = f
		}
	}

	if val, exists := paramMap["max_weights_limit"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.MaxWeightsLimit = f
		}
	}

	if val, exists := paramMap["min_allowed_weights"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.MinAllowedWeights = f
		}
	}

	if val, exists := paramMap["weights_set_rate_limit"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.WeightsSetRateLimit = f
		}
	}

	if val, exists := paramMap["tempo"]; exists {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			params.Tempo = f
		}
	}

	if val, exists := paramMap["bonds_penalty"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.BondsPenalty = f
		}
	}

	if val, exists := paramMap["bonds_moving_average"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.BondsMovingAverage = f
		}
	}

	// Parse new parameters
	if val, exists := paramMap["rho"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.Rho = f
		}
	}

	if val, exists := paramMap["liquid_alpha_enabled"]; exists {
		if b, err := strconv.ParseBool(val); err == nil {
			params.LiquidAlphaEnabled = b
		}
	}

	if val, exists := paramMap["alpha_sigmoid_steepness"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.AlphaSigmoidSteepness = f
		}
	}

	if val, exists := paramMap["alpha_low"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.AlphaLow = f
		}
	}

	if val, exists := paramMap["alpha_high"]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			params.AlphaHigh = f
		}
	}

	return params
}

// ValidatorInfo validator information (obtained from event module)
type ValidatorInfo struct {
	Address string   `json:"address"`
	Stake   float64  `json:"stake"`
	Weights []uint64 `json:"weights"`
	Active  bool     `json:"active"`
}

// SubnetEpochData subnet epoch data
type SubnetEpochData struct {
	Netuid     uint16          `json:"netuid"`
	Validators []ValidatorInfo `json:"validators"`
	Params     EpochParams     `json:"params"`
	Emission   uint64          `json:"emission"`
}
