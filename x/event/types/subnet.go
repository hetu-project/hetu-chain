package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Subnet represents a subnet in the network
type Subnet struct {
	Netuid                uint16            `json:"netuid" yaml:"netuid"`
	Owner                 string            `json:"owner" yaml:"owner"`
	LockedAmount          string            `json:"lock_amount" yaml:"lock_amount"` // Renamed from LockAmount for consistency
	BurnedAmount          string            `json:"burned_tao" yaml:"burned_tao"`   // Renamed from BurnedTao for consistency
	AmmPool               string            `json:"pool" yaml:"pool"`               // Renamed from Pool for consistency
	Params                map[string]string `json:"params" yaml:"params"`
	FirstEmissionBlock    uint64            `json:"first_emission_block" yaml:"first_emission_block"`         // First emission block number
	Mechanism             uint8             `json:"mechanism" yaml:"mechanism"`                               // Subnet mechanism (0=stable, 1=dynamic)
	EMAPriceHalvingBlocks uint64            `json:"ema_price_halving_blocks" yaml:"ema_price_halving_blocks"` // EMA price halving blocks (default 201600=4 weeks)
}

// GetLockedAmountInt returns the LockedAmount as math.Int
func (s Subnet) GetLockedAmountInt() (math.Int, error) {
	if s.LockedAmount == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(s.LockedAmount)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid locked amount: %s", s.LockedAmount)
	}
	return amount, nil
}

// GetBurnedAmountInt returns the BurnedAmount as math.Int
func (s Subnet) GetBurnedAmountInt() (math.Int, error) {
	if s.BurnedAmount == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(s.BurnedAmount)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid burned amount: %s", s.BurnedAmount)
	}
	return amount, nil
}

// SubnetHyperparams represents the hyperparameters for a subnet
type SubnetHyperparams struct {
	// === Core network parameters ===
	Rho                  uint16 `json:"rho" yaml:"rho"`
	Kappa                uint16 `json:"kappa" yaml:"kappa"`
	ImmunityPeriod       uint16 `json:"immunity_period" yaml:"immunity_period"`
	Tempo                uint16 `json:"tempo" yaml:"tempo"`
	MaxValidators        uint16 `json:"max_validators" yaml:"max_validators"`
	ActivityCutoff       uint16 `json:"activity_cutoff" yaml:"activity_cutoff"`
	MaxAllowedUids       uint16 `json:"max_allowed_uids" yaml:"max_allowed_uids"`
	MaxAllowedValidators uint16 `json:"max_allowed_validators" yaml:"max_allowed_validators"`
	MinAllowedWeights    uint16 `json:"min_allowed_weights" yaml:"min_allowed_weights"`
	MaxWeightsLimit      uint16 `json:"max_weights_limit" yaml:"max_weights_limit"`

	// === Economic parameters ===
	BaseNeuronCost        string `json:"base_neuron_cost" yaml:"base_neuron_cost"`
	CurrentDifficulty     uint64 `json:"current_difficulty" yaml:"current_difficulty"`
	TargetRegsPerInterval uint16 `json:"target_regs_per_interval" yaml:"target_regs_per_interval"`
	MaxRegsPerBlock       uint16 `json:"max_regs_per_block" yaml:"max_regs_per_block"`
	WeightsRateLimit      uint64 `json:"weights_rate_limit" yaml:"weights_rate_limit"`

	// === Governance parameters ===
	RegistrationAllowed bool   `json:"registration_allowed" yaml:"registration_allowed"`
	CommitRevealEnabled bool   `json:"commit_reveal_enabled" yaml:"commit_reveal_enabled"`
	CommitRevealPeriod  uint64 `json:"commit_reveal_period" yaml:"commit_reveal_period"`
	ServingRateLimit    uint64 `json:"serving_rate_limit" yaml:"serving_rate_limit"`
	ValidatorThreshold  string `json:"validator_threshold" yaml:"validator_threshold"`
	NeuronThreshold     string `json:"neuron_threshold" yaml:"neuron_threshold"`
}

// SubnetInfo represents subnet information from contract events
type SubnetInfo struct {
	Netuid         uint16 `json:"netuid" yaml:"netuid"`
	Owner          string `json:"owner" yaml:"owner"`
	AlphaToken     string `json:"alpha_token" yaml:"alpha_token"`
	AmmPool        string `json:"amm_pool" yaml:"amm_pool"`
	LockedAmount   string `json:"locked_amount" yaml:"locked_amount"`
	PoolInitialTao string `json:"pool_initial_tao" yaml:"pool_initial_tao"`
	BurnedAmount   string `json:"burned_amount" yaml:"burned_amount"`
	CreatedAt      uint64 `json:"created_at" yaml:"created_at"`
	IsActive       bool   `json:"is_active" yaml:"is_active"`
	Name           string `json:"name" yaml:"name"`
	Description    string `json:"description" yaml:"description"`
	ActivatedAt    uint64 `json:"activated_at" yaml:"activated_at"`       // Activation timestamp
	ActivatedBlock uint64 `json:"activated_block" yaml:"activated_block"` // Activation block
}

// GetLockedAmountInt returns the LockedAmount as math.Int
func (s SubnetInfo) GetLockedAmountInt() (math.Int, error) {
	if s.LockedAmount == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(s.LockedAmount)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid locked amount: %s", s.LockedAmount)
	}
	return amount, nil
}

// GetBurnedAmountInt returns the BurnedAmount as math.Int
func (s SubnetInfo) GetBurnedAmountInt() (math.Int, error) {
	if s.BurnedAmount == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(s.BurnedAmount)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid burned amount: %s", s.BurnedAmount)
	}
	return amount, nil
}

// GetPoolInitialTaoInt returns the PoolInitialTao as math.Int
func (s SubnetInfo) GetPoolInitialTaoInt() (math.Int, error) {
	if s.PoolInitialTao == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(s.PoolInitialTao)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid pool initial tao: %s", s.PoolInitialTao)
	}
	return amount, nil
}

// ToSubnet converts SubnetInfo to Subnet
func (s SubnetInfo) ToSubnet(params map[string]string, mechanism uint8, emaPriceHalvingBlocks uint64) Subnet {
	return Subnet{
		Netuid:                s.Netuid,
		Owner:                 s.Owner,
		LockedAmount:          s.LockedAmount,
		BurnedAmount:          s.BurnedAmount,
		AmmPool:               s.AmmPool,
		Params:                params,
		FirstEmissionBlock:    s.ActivatedBlock,
		Mechanism:             mechanism,
		EMAPriceHalvingBlocks: emaPriceHalvingBlocks,
	}
}

// NeuronInfo represents neuron information from contract events
type NeuronInfo struct {
	Account                string `json:"account" yaml:"account"`
	Netuid                 uint16 `json:"netuid" yaml:"netuid"`
	IsActive               bool   `json:"is_active" yaml:"is_active"`
	IsValidator            bool   `json:"is_validator" yaml:"is_validator"`
	RequestedValidatorRole bool   `json:"requested_validator_role" yaml:"requested_validator_role"`
	Stake                  string `json:"stake" yaml:"stake"`
	RegistrationBlock      uint64 `json:"registration_block" yaml:"registration_block"`
	LastUpdate             uint64 `json:"last_update" yaml:"last_update"`
	AxonEndpoint           string `json:"axon_endpoint" yaml:"axon_endpoint"`
	AxonPort               uint32 `json:"axon_port" yaml:"axon_port"`
	PrometheusEndpoint     string `json:"prometheus_endpoint" yaml:"prometheus_endpoint"`
	PrometheusPort         uint32 `json:"prometheus_port" yaml:"prometheus_port"`
}

// GetStakeInt returns the Stake as math.Int
func (n NeuronInfo) GetStakeInt() (math.Int, error) {
	if n.Stake == "" {
		return math.ZeroInt(), nil
	}

	amount, ok := math.NewIntFromString(n.Stake)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid stake amount: %s", n.Stake)
	}
	return amount, nil
}

// SubnetPriceData represents subnet price-related data
type SubnetPriceData struct {
	MovingPrice    math.LegacyDec `json:"moving_price" yaml:"moving_price"`         // Moving price
	AlphaPrice     math.LegacyDec `json:"alpha_price" yaml:"alpha_price"`           // Alpha price
	SubnetTAO      math.Int       `json:"subnet_tao" yaml:"subnet_tao"`             // Amount of TAO in subnet
	SubnetAlphaIn  math.Int       `json:"subnet_alpha_in" yaml:"subnet_alpha_in"`   // Amount of Alpha in pool
	SubnetAlphaOut math.Int       `json:"subnet_alpha_out" yaml:"subnet_alpha_out"` // Amount of Alpha in subnet
	Volume         math.Int       `json:"volume" yaml:"volume"`                     // Total trading volume
}

// SubnetEmissionData represents emission data for a subnet
type SubnetEmissionData struct {
	TaoInEmission    math.Int `json:"tao_in_emission" yaml:"tao_in_emission"`       // TAO input emission
	AlphaInEmission  math.Int `json:"alpha_in_emission" yaml:"alpha_in_emission"`   // Alpha input emission
	AlphaOutEmission math.Int `json:"alpha_out_emission" yaml:"alpha_out_emission"` // Alpha output emission
}
