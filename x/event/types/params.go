package types

// DefaultParamsMap returns default parameter mapping
// When the param field in contract events doesn't specify certain parameters, use these default values
func DefaultParamsMap() map[string]string {
	return map[string]string{
		KeyRho:                        "0.5",   // Default rho value
		KeyKappa:                      "32767", // Default kappa value
		KeyMaxAllowedUids:             "4096",  // Default maximum allowed subnet count
		KeyImmunityPeriod:             "4096",  // Default immunity period
		KeyActivityCutoff:             "5000",  // Default activity cutoff
		KeyMaxWeightsLimit:            "1000",  // Default maximum weight limit
		"weights_version_key":         "0",     // Default weight version
		KeyMinAllowedWeights:          "8",     // Default minimum allowed weights
		KeyMaxAllowedValidators:       "128",   // Default maximum allowed validators
		KeyTempo:                      "100",   // Default tempo value
		"adjustment_interval":         "112",   // Default adjustment interval
		"adjustment_alpha":            "58982", // Default adjustment alpha value
		KeyBondsMovingAverage:         "0.9",   // Default moving average
		KeyWeightsSetRateLimit:        "1000",  // Default weight setting rate limit (corresponds to weights_rate_limit)
		"validator_prune_len":         "100",   // Default validator pruning length
		"validator_logits_divergence": "0.1",   // Default validator logits divergence
		"validator_sequence_length":   "100",   // Default validator sequence length
		"validator_epoch_length":      "100",   // Default validator epoch length
		"validator_epochs_per_reset":  "100",   // Default validator epoch reset interval
		KeyLiquidAlphaEnabled:         "true",  // Default liquid alpha enabled
		KeyAlphaEnabled:               "true",  // Default alpha enabled
		KeyAlphaHigh:                  "0.9",   // Default alpha upper limit
		KeyAlphaLow:                   "0.1",   // Default alpha lower limit
		KeyBondsPenalty:               "0.1",   // Default stake penalty

		// New stakework required parameters
		KeyAlpha:                 "0.1",  // Alpha parameter required for stakework epoch algorithm
		KeyDelta:                 "1.0",  // Delta parameter required for stakework epoch algorithm
		KeyAlphaSigmoidSteepness: "10.0", // Alpha sigmoid steepness
	}
}
