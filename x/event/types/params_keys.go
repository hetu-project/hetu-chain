// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package types

// Parameter key constants - canonical form is snake_case
const (
	// Core parameters
	KeyRho                  = "rho"
	KeyKappa                = "kappa"
	KeyImmunityPeriod       = "immunity_period"
	KeyTempo                = "tempo"
	KeyMaxValidators        = "max_validators"
	KeyActivityCutoff       = "activity_cutoff"
	KeyMaxAllowedUids       = "max_allowed_uids"
	KeyMaxAllowedValidators = "max_allowed_validators"
	KeyMinAllowedWeights    = "min_allowed_weights"
	KeyMaxWeightsLimit      = "max_weights_limit"

	// Economic parameters
	KeyBaseNeuronCost        = "base_neuron_cost"
	KeyCurrentDifficulty     = "current_difficulty"
	KeyTargetRegsPerInterval = "target_regs_per_interval"
	KeyMaxRegsPerBlock       = "max_regs_per_block"
	KeyWeightsRateLimit      = "weights_rate_limit"
	KeyWeightsSetRateLimit   = "weights_set_rate_limit"

	// Governance parameters
	KeyRegistrationAllowed = "registration_allowed"
	KeyCommitRevealEnabled = "commit_reveal_enabled"
	KeyCommitRevealPeriod  = "commit_reveal_period"
	KeyServingRateLimit    = "serving_rate_limit"
	KeyValidatorThreshold  = "validator_threshold"
	KeyNeuronThreshold     = "neuron_threshold"

	// Bonds and alpha parameters
	KeyBondsMovingAverage    = "bonds_moving_average"
	KeyBondsPenalty          = "bonds_penalty"
	KeyLiquidAlphaEnabled    = "liquid_alpha_enabled"
	KeyAlphaEnabled          = "alpha_enabled"
	KeyAlphaHigh             = "alpha_high"
	KeyAlphaLow              = "alpha_low"
	KeyAlpha                 = "alpha"
	KeyDelta                 = "delta"
	KeyAlphaSigmoidSteepness = "alpha_sigmoid_steepness"
)

// KeyAliases maps camelCase keys from ABI to canonical snake_case keys
var KeyAliases = map[string]string{
	// Core parameters
	"rho":                  KeyRho,
	"kappa":                KeyKappa,
	"immunityPeriod":       KeyImmunityPeriod,
	"tempo":                KeyTempo,
	"maxValidators":        KeyMaxValidators,
	"activityCutoff":       KeyActivityCutoff,
	"maxAllowedUids":       KeyMaxAllowedUids,
	"maxAllowedValidators": KeyMaxAllowedValidators,
	"minAllowedWeights":    KeyMinAllowedWeights,
	"maxWeightsLimit":      KeyMaxWeightsLimit,

	// Economic parameters
	"baseNeuronCost":        KeyBaseNeuronCost,
	"currentDifficulty":     KeyCurrentDifficulty,
	"targetRegsPerInterval": KeyTargetRegsPerInterval,
	"maxRegsPerBlock":       KeyMaxRegsPerBlock,
	"weightsRateLimit":      KeyWeightsRateLimit,
	"weightsSetRateLimit":   KeyWeightsSetRateLimit,

	// Governance parameters
	"registrationAllowed": KeyRegistrationAllowed,
	"commitRevealEnabled": KeyCommitRevealEnabled,
	"commitRevealPeriod":  KeyCommitRevealPeriod,
	"servingRateLimit":    KeyServingRateLimit,
	"validatorThreshold":  KeyValidatorThreshold,
	"neuronThreshold":     KeyNeuronThreshold,

	// Bonds and alpha parameters
	"bondsMovingAverage":    KeyBondsMovingAverage,
	"bondsPenalty":          KeyBondsPenalty,
	"liquidAlphaEnabled":    KeyLiquidAlphaEnabled,
	"alphaEnabled":          KeyAlphaEnabled,
	"alphaHigh":             KeyAlphaHigh,
	"alphaLow":              KeyAlphaLow,
	"alpha":                 KeyAlpha,
	"delta":                 KeyDelta,
	"alphaSigmoidSteepness": KeyAlphaSigmoidSteepness,
}

// NormalizeParamKey converts a parameter key to its canonical form
// If the key is already in canonical form, it is returned unchanged
// If the key is an alias (camelCase), it is converted to its canonical form (snake_case)
func NormalizeParamKey(key string) string {
	if canonicalKey, exists := KeyAliases[key]; exists {
		return canonicalKey
	}
	return key
}

// NormalizeParamKeys normalizes all keys in a parameter map
func NormalizeParamKeys(params map[string]string) map[string]string {
	normalized := make(map[string]string, len(params))

	// 1st pass: set all alias-derived keys
	for key, value := range params {
		if canon := NormalizeParamKey(key); canon != key {
			normalized[canon] = value
		}
	}

	// 2nd pass: set canonical keys last (canonical wins on conflicts)
	for key, value := range params {
		if NormalizeParamKey(key) == key {
			normalized[key] = value
		}
	}

	return normalized
}
