package types

import (
	"cosmossdk.io/math"
)

// SubnetInfo defines the basic information of a subnet
type SubnetInfo struct {
	Netuid         uint32   `json:"netuid"`
	Owner          string   `json:"owner"`
	AlphaToken     string   `json:"alpha_token"`
	AmmPool        string   `json:"amm_pool"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	IsActive       bool     `json:"is_active"`
	CreatedAt      uint64   `json:"created_at"`
	LockedAmount   math.Int `json:"locked_amount"`
	PoolInitialTao math.Int `json:"pool_initial_tao"`
	BurnedAmount   math.Int `json:"burned_amount"`
}

// SubnetHyperparams defines the hyperparameters of a subnet
type SubnetHyperparams struct {
	Tempo                     uint64         `json:"tempo"`
	SubnetEmissionValue       uint32         `json:"subnet_emission_value"`
	SubnetOwnerCut            math.LegacyDec `json:"subnet_owner_cut"`
	MaxAllowedValidators      uint32         `json:"max_allowed_validators"`
	MaxAllowedUids            uint32         `json:"max_allowed_uids"`
	ImmunityPeriod            uint32         `json:"immunity_period"`
	MinAllowedWeights         uint32         `json:"min_allowed_weights"`
	MaxWeightLimit            uint32         `json:"max_weight_limit"`
	MaxWeightAge              math.LegacyDec `json:"max_weight_age"`
	WeightConsensus           math.LegacyDec `json:"weight_consensus"`
	WeightMaxAge              uint32         `json:"weight_max_age"`
	ScalingLawPower           math.LegacyDec `json:"scaling_law_power"`
	ValidatorExcludeQuantile  math.LegacyDec `json:"validator_exclude_quantile"`
	ValidatorPruneLen         math.LegacyDec `json:"validator_prune_len"`
	ValidatorLogitsDivergence math.LegacyDec `json:"validator_logits_divergence"`
	BlocksSinceLastStep       uint32         `json:"blocks_since_last_step"`
	LastMechanismStepBlock    uint64         `json:"last_mechanism_step_block"`
	BlocksPerStep             uint32         `json:"blocks_per_step"`
	BondsMovingAverage        uint32         `json:"bonds_moving_average"`
	SubnetMovingAlpha         math.LegacyDec `json:"subnet_moving_alpha"`
	EmaPriceHalvingBlocks     uint32         `json:"ema_price_halving_blocks"`
}

// NeuronInfo defines the information of a neuron in a subnet
type NeuronInfo struct {
	Uid        string         `json:"uid"`
	Hotkey     string         `json:"hotkey"`
	Coldkey    string         `json:"coldkey"`
	Stake      math.Int       `json:"stake"`
	LastUpdate uint64         `json:"last_update"`
	Rank       uint32         `json:"rank"`
	Emission   math.Int       `json:"emission"`
	Incentive  math.Int       `json:"incentive"`
	Trust      math.LegacyDec `json:"trust"`
	Consensus  math.LegacyDec `json:"consensus"`
	Dividends  math.Int       `json:"dividends"`
	IsActive   bool           `json:"is_active"`
}

// PoolInfo defines the information of a subnet pool
type PoolInfo struct {
	Netuid        uint32         `json:"netuid"`
	TaoIn         math.Int       `json:"tao_in"`
	AlphaIn       math.Int       `json:"alpha_in"`
	AlphaOut      math.Int       `json:"alpha_out"`
	CurrentPrice  math.LegacyDec `json:"current_price"`
	MovingPrice   math.LegacyDec `json:"moving_price"`
	TotalVolume   math.Int       `json:"total_volume"`
	MechanismType uint32         `json:"mechanism_type"`
}
