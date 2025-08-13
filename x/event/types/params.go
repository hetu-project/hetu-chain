package types

// DefaultParamsMap returns default parameter mapping
// When the param field in contract events doesn't specify certain parameters, use these default values
func DefaultParamsMap() map[string]string {
	return map[string]string{
		"rho":                         "0.5",   // Default rho value
		"kappa":                       "32767", // Default kappa value
		"max_allowed_uids":            "4096",  // Default maximum allowed subnet count
		"immunity_period":             "4096",  // Default immunity period
		"activity_cutoff":             "5000",  // Default activity cutoff
		"max_weights_limit":           "1000",  // Default maximum weight limit
		"weights_version_key":         "0",     // Default weight version
		"min_allowed_weights":         "8",     // Default minimum allowed weights
		"max_allowed_validators":      "128",   // Default maximum allowed validators
		"tempo":                       "100",   // Default tempo value
		"adjustment_interval":         "112",   // Default adjustment interval
		"adjustment_alpha":            "58982", // Default adjustment alpha value
		"bonds_moving_average":        "0.9",   // Default moving average
		"weights_set_rate_limit":      "1000",  // Default weight setting rate limit (corresponds to weights_rate_limit)
		"validator_prune_len":         "100",   // Default validator pruning length
		"validator_logits_divergence": "0.1",   // Default validator logits divergence
		"validator_sequence_length":   "100",   // Default validator sequence length
		"validator_epoch_length":      "100",   // Default validator epoch length
		"validator_epochs_per_reset":  "100",   // Default validator epoch reset interval
		"liquid_alpha_enabled":        "true",  // Default liquid alpha enabled
		"alpha_enabled":               "true",  // Default alpha enabled
		"alpha_high":                  "0.9",   // Default alpha upper limit
		"alpha_low":                   "0.1",   // Default alpha lower limit
		"bonds_penalty":               "0.1",   // Default stake penalty

		// New stakework required parameters
		"alpha":                   "0.1",  // Alpha parameter required for stakework epoch algorithm
		"delta":                   "1.0",  // Delta parameter required for stakework epoch algorithm
		"alpha_sigmoid_steepness": "10.0", // Alpha sigmoid steepness
	}
}
