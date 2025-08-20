package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	stdmath "math"
	"math/big"
	"sort"

	cosmosmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/stakework/types"
)

// Error definitions
var (
	ErrSubnetNotFound = errors.New("subnet not found")
	ErrEpochNotDue    = errors.New("epoch not due")
)

// RunEpoch runs the complete Bittensor epoch algorithm
func (k Keeper) RunEpoch(ctx sdk.Context, netuid uint16, raoEmission cosmosmath.Int) (*types.EpochResult, error) {
	logger := k.Logger(ctx)
	logger.Debug("Starting epoch calculation", "netuid", netuid, "emission", raoEmission.String())

	// 1. Get subnet data
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		logger.Error("Subnet not found", "netuid", netuid)
		return nil, ErrSubnetNotFound
	}
	logger.Debug("Retrieved subnet data", "netuid", netuid, "subnet", subnet)

	// 2. Parse subnet parameters
	params, err := types.ParseEpochParams(subnet.Params)
	if err != nil {
		logger.Error("Failed to parse epoch parameters", "netuid", netuid, "err", err)
		return nil, fmt.Errorf("failed to parse epoch parameters: %w", err)
	}
	logger.Debug("Parsed subnet parameters",
		"kappa", params.Kappa,
		"alpha", params.Alpha,
		"delta", params.Delta,
		"tempo", params.Tempo,
		"rho", params.Rho,
		"liquid_alpha_enabled", params.LiquidAlphaEnabled)

	// 3. Check if epoch should run
	if !k.shouldRunEpoch(ctx, netuid, params.Tempo) {
		logger.Debug("Epoch not due yet",
			"netuid", netuid,
			"current_block", ctx.BlockHeight(),
			"tempo", params.Tempo)
		return nil, ErrEpochNotDue
	}
	logger.Debug("Epoch should run", "current_block", ctx.BlockHeight())

	// 4. Get all validator data for the subnet
	validators := k.getSubnetValidators(ctx, netuid)
	logger.Debug("Retrieved validators", "count", len(validators))
	if len(validators) == 0 {
		logger.Debug("No validators found, returning empty result")
		return &types.EpochResult{
			Netuid:    netuid,
			Accounts:  []string{},
			Emission:  []cosmosmath.Int{},
			Dividend:  []cosmosmath.Int{},
			Incentive: []cosmosmath.Int{},
			Bonds:     [][]float64{},
			Consensus: []float64{},
		}, nil
	}

	// 5. Calculate active status
	active := k.calculateActive(ctx, netuid, validators, params)
	activeCount := 0
	for _, isActive := range active {
		if isActive {
			activeCount++
		}
	}
	logger.Debug("Calculated active status",
		"total_validators", len(validators),
		"active_validators", activeCount)

	// 6. Get stake weights
	stake := k.getStakeWeights(ctx, netuid, validators)
	logger.Debug("Retrieved stake weights",
		"count", len(stake),
		"total_stake", k.sumArray(stake))

	// 7. Normalize stake weights
	activeStake := k.normalizeStake(stake, active)
	logger.Debug("Normalized stake weights",
		"normalized_sum", k.sumArray(activeStake))

	// 8. Get weights matrix
	weights := k.getWeightsMatrix(ctx, netuid, validators)
	logger.Debug("Retrieved weights matrix",
		"matrix_size", fmt.Sprintf("%dx%d", len(weights), len(weights[0])),
		"sample_weights", k.sampleWeights(weights, 3))

	// 9. Calculate consensus scores
	consensus := k.weightedMedianCol(activeStake, weights, params.Kappa)
	logger.Debug("Calculated consensus scores",
		"scores_count", len(consensus),
		"min_score", k.minArray(consensus),
		"max_score", k.maxArray(consensus),
		"avg_score", k.avgArray(consensus))

	// 10. Clip weights
	clippedWeights := k.clipWeights(weights, consensus, params.Delta)
	logger.Debug("Clipped weights",
		"matrix_size", fmt.Sprintf("%dx%d", len(clippedWeights), len(clippedWeights[0])),
		"sample_clipped", k.sampleWeights(clippedWeights, 3))

	// 11. Calculate bonds
	prevBonds := k.getPrevBonds(ctx, netuid, validators)
	var bonds [][]float64

	if params.LiquidAlphaEnabled {
		logger.Debug("Using dynamic alpha for bonds calculation")
		// Use dynamic alpha
		alphas := k.computeLiquidAlphaValues(weights, prevBonds, consensus, params)
		logger.Debug("Computed liquid alpha values",
			"matrix_size", fmt.Sprintf("%dx%d", len(alphas), len(alphas[0])),
			"min_alpha", k.minMatrix(alphas),
			"max_alpha", k.maxMatrix(alphas),
			"avg_alpha", k.avgMatrix(alphas))
		bonds = k.computeBondsWithDynamicAlpha(clippedWeights, prevBonds, alphas)
	} else {
		logger.Debug("Using fixed alpha for bonds calculation",
			"alpha", params.BondsMovingAverage)
		// Use fixed alpha
		fixedAlpha := k.computeDisabledLiquidAlpha(params.BondsMovingAverage)
		bonds = k.computeBonds(clippedWeights, prevBonds, fixedAlpha)
	}
	logger.Debug("Calculated bonds",
		"matrix_size", fmt.Sprintf("%dx%d", len(bonds), len(bonds[0])),
		"min_bond", k.minMatrix(bonds),
		"max_bond", k.maxMatrix(bonds),
		"avg_bond", k.avgMatrix(bonds))

	// 12. Calculate dividends
	dividends := k.computeDividends(bonds)
	logger.Debug("Calculated dividends",
		"count", len(dividends),
		"total_dividends", k.sumArray(dividends))

	// 13. Calculate incentive
	incentive := k.computeIncentive(clippedWeights, activeStake, params.Rho)
	logger.Debug("Calculated incentive",
		"count", len(incentive),
		"total_incentive", k.sumArray(incentive))

	// 14. Normalize dividends and incentive
	normDividends := k.normalizeDividends(dividends, active)
	normIncentive := k.normalizeIncentive(incentive, active)
	logger.Debug("Normalized rewards",
		"normalized_dividends_sum", k.sumArray(normDividends),
		"normalized_incentive_sum", k.sumArray(normIncentive))

	// 15. Distribute emission
	// 使用math.Int直接传递，不再转换为uint64
	emission := k.distributeEmission(normIncentive, normDividends, raoEmission)

	// 使用math.Int计算总emission
	totalEmission := cosmosmath.ZeroInt()
	for _, e := range emission {
		totalEmission = totalEmission.Add(e)
	}

	logger.Debug("Distributed emission",
		"total_emission", totalEmission.String(),
		"target_emission", raoEmission.String())

	// 16. Build result
	result := &types.EpochResult{
		Netuid:    netuid,
		Accounts:  make([]string, len(validators)),
		Emission:  emission,
		Dividend:  make([]cosmosmath.Int, len(validators)),
		Incentive: make([]cosmosmath.Int, len(validators)),
		Bonds:     bonds,
		Consensus: consensus,
	}

	// Populate account addresses, dividends, and incentive
	for i, validator := range validators {
		result.Accounts[i] = validator.Address
		// 计算奖励金额
		incentiveFloat := normIncentive[i]
		dividendsFloat := normDividends[i]

		// 将浮点数转换为Dec
		incentiveDec := cosmosmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", incentiveFloat))
		dividendsDec := cosmosmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", dividendsFloat))

		result.Dividend[i] = cosmosmath.NewIntFromBigInt(
			dividendsDec.MulInt(raoEmission).TruncateInt().BigInt(),
		)
		result.Incentive[i] = cosmosmath.NewIntFromBigInt(
			incentiveDec.MulInt(raoEmission).TruncateInt().BigInt(),
		)
	}

	// 17. Save bonds to storage
	k.saveBonds(ctx, netuid, validators, bonds)
	logger.Debug("Saved bonds to storage")

	logger.Debug("Epoch calculation completed successfully",
		"netuid", netuid,
		"validators_count", len(validators),
		"active_count", activeCount,
		"total_emission", totalEmission)

	return result, nil
}

// shouldRunEpoch checks if an epoch should run
// Based on Bittensor's actual implementation: based on block number formula
// (block_number + netuid + 1) % (tempo + 1) == 0
func (k Keeper) shouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool {
	currentBlock := uint64(ctx.BlockHeight())

	// Bittensor's epoch formula:
	// (block_number + netuid + 1) % (tempo + 1) == 0
	result := (currentBlock + uint64(netuid) + 1) % (tempo + 1)
	return result == 0
}

// ShouldRunEpoch export method, satisfies blockinflation/types.StakeworkKeeper interface
func (k Keeper) ShouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool {
	// Add debug information
	if k.eventKeeper == nil {
		k.Logger(ctx).Error("ShouldRunEpoch: eventKeeper is nil")
		return false
	}

	// Add debug information for the calculation
	currentBlock := uint64(ctx.BlockHeight())
	k.Logger(ctx).Debug("ShouldRunEpoch calculation", "netuid", netuid, "tempo", tempo, "currentBlock", currentBlock)

	return k.shouldRunEpoch(ctx, netuid, tempo)
}

// getSubnetValidators gets all validators for a subnet
func (k Keeper) getSubnetValidators(ctx sdk.Context, netuid uint16) []types.ValidatorInfo {
	// Get all stake information from the event module
	stakes := k.eventKeeper.GetAllValidatorStakesByNetuid(ctx, netuid)

	validators := make([]types.ValidatorInfo, 0, len(stakes))
	validatorMap := make(map[string]types.ValidatorInfo)

	// Process stake information
	for _, stake := range stakes {
		validator := types.ValidatorInfo{
			Address: stake.Validator,
			Stake:   stake.Amount, // Keep as string to avoid precision loss
			Weights: []uint64{},
			Active:  true, // Default active, will be updated later
		}
		validatorMap[stake.Validator] = validator
	}

	// Process weight information
	for validatorAddr := range validatorMap {
		weight, found := k.eventKeeper.GetValidatorWeight(ctx, netuid, validatorAddr)
		if found {
			// Convert map to array
			weights := make([]uint64, 0, len(weight.Weights))
			for _, w := range weight.Weights {
				weights = append(weights, w)
			}

			if v, exists := validatorMap[validatorAddr]; exists {
				v.Weights = weights
				validatorMap[validatorAddr] = v
			}
		}
	}

	// Convert to array
	for _, validator := range validatorMap {
		validators = append(validators, validator)
	}

	return validators
}

// calculateActive calculates active status
func (k Keeper) calculateActive(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo, params types.EpochParams) []bool {
	active := make([]bool, len(validators))
	currentBlock := uint64(ctx.BlockHeight())

	for i, validator := range validators {
		// Check if the validator has activity within the active cutoff time
		// This needs to get the last active time from the event module
		// For now, using a simple logic: all validators are set to active
		// TODO: Implement actual active status check logic
		lastActiveBlock := k.getLastActiveBlock(ctx, netuid, validator.Address)
		if currentBlock-lastActiveBlock <= params.ActivityCutoff {
			active[i] = true
		} else {
			active[i] = false
		}
	}

	return active
}

// getLastActiveBlock gets the last active block for a validator (temporary implementation)
func (k Keeper) getLastActiveBlock(ctx sdk.Context, netuid uint16, validator string) uint64 {
	// TODO: Get the last active time for the validator from the event module
	// For now, return the current block, indicating all validators are active
	return uint64(ctx.BlockHeight())
}

// getStakeWeights gets stake weights
func (k Keeper) getStakeWeights(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) []float64 {
	stake := make([]float64, len(validators))
	for i, validator := range validators {
		// Parse string stake to big.Int, then convert to float64
		bigIntStake, err := validator.GetStakeBigInt()
		if err != nil {
			k.Logger(ctx).Error("Failed to parse stake", "validator", validator.Address, "stake", validator.Stake, "error", err)
			continue // Skip this validator or set to 0
		}

		// Convert to float64 (with potential precision loss, but needed for algorithm)
		stakeFloat := new(big.Float).SetInt(bigIntStake)
		stakeValue, _ := stakeFloat.Float64()
		stake[i] = stakeValue
	}
	return stake
}

// normalizeStake normalizes stake weights
func (k Keeper) normalizeStake(stake []float64, active []bool) []float64 {
	sum := 0.0
	for i, s := range stake {
		if active[i] {
			sum += s
		}
	}

	if sum == 0 {
		return make([]float64, len(stake))
	}

	out := make([]float64, len(stake))
	for i, s := range stake {
		if active[i] {
			out[i] = s / sum
		}
	}
	return out
}

// getWeightsMatrix gets the weights matrix
func (k Keeper) getWeightsMatrix(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) [][]float64 {
	n := len(validators)
	weights := make([][]float64, n)

	for i := 0; i < n; i++ {
		weights[i] = make([]float64, n)
		if i < len(validators[i].Weights) {
			for j := 0; j < n; j++ {
				if j < len(validators[i].Weights) {
					weights[i][j] = float64(validators[i].Weights[j])
				}
			}
		}
	}

	return weights
}

// weightedMedianCol calculates weighted median
/*
Example
Assume the following inputs:

    stake = [100, 200, 300]: three validator's stake.
    weights = [[0.1, 0.2, 0.3], [0.4, 0.5, 0.6], [0.7, 0.8, 0.9]]: three validators' weights for three nodes.
    kappa = 0.5: threshold for median calculation.

We expect the output to be the consensus score for each node, i.e., the weighted median for each node.
Calculation process:

    For node 0:
        Scores = [1000.1, 2000.4, 300*0.7] = [10, 80, 210]
        Sorted = [10, 80, 210]
        Weighted Median = 80 (because 0.5 * (10+80+210) = 125, 80 is the first number greater than or equal to 125)
    For node 1:
        Scores = [1000.2, 2000.5, 300*0.8] = [20, 100, 240]
        Sorted = [20, 100, 240]
        Weighted Median = 100 (because 0.5 * (20+100+240) = 140, 100 is the first number greater than or equal to 140)
    For node 2:
        Scores = [1000.3, 2000.6, 300*0.9] = [30, 120, 270]
        Sorted = [30, 120, 270]
        Weighted Median = 120 (because 0.5 * (30+120+270) = 195, 120 is the first number greater than or equal to 195)

Output:

    consensus = [80, 100, 120]
*/
func (k Keeper) weightedMedianCol(stake []float64, weights [][]float64, kappa float64) []float64 {
	n := len(weights[0]) // Number of nodes
	m := len(stake)      // Number of validators
	consensus := make([]float64, n)

	for j := 0; j < n; j++ {
		// Collect all scores for node j from validators
		type pair struct {
			w, s float64
		}
		var pairs []pair
		for i := 0; i < m; i++ {
			if i < len(weights) && j < len(weights[i]) {
				pairs = append(pairs, pair{w: weights[i][j], s: stake[i]})
			}
		}

		// Sort by score in ascending order
		sort.Slice(pairs, func(a, b int) bool {
			return pairs[a].w < pairs[b].w
		})

		// Calculate weighted median
		total := 0.0
		for _, p := range pairs {
			total += p.s
		}

		acc := 0.0
		for _, p := range pairs {
			acc += p.s
			if acc >= kappa*total {
				consensus[j] = p.w
				break
			}
		}
	}

	return consensus
}

// clipWeights clips weights to prevent single node weights from being too large or too small
/*

Assume the following inputs:

    weights = [[0.7, 0.8, 0.9], [0.4, 0.5, 0.6], [0.1, 0.2, 0.3]]: a 3x3 weight matrix.
    consensus = [0.5, 0.6, 0.7]: consensus score for each node.
    delta = 0.1: magnitude of weight adjustment.

Calculation process:
For each node j's weight, calculate max and min:

    For node 0:
        Max weight max = 0.6 + 0.1 = 0.7
        Min weight min = 0.5 - 0.1 = 0.4
        Clipped weight: [0.4, 0.6, 0.7] (0.8 and 0.9 in original weights are outside the range)
    For node 1:
        Max weight max = 0.7 + 0.1 = 0.8
        Min weight min = 0.6 - 0.1 = 0.5
        Clipped weight: [0.5, 0.5, 0.6] (all weights are within the range)
    For node 2:
        Max weight max = 0.8 + 0.1 = 0.9
        Min weight min = 0.7 - 0.1 = 0.6
        Clipped weight: [0.6, 0.7, 0.7] (all weights in original weights are within the range)

Output:

    clipped = [[0.4, 0.6, 0.7], [0.5, 0.5, 0.6], [0.6, 0.7, 0.7]]
*/
func (k Keeper) clipWeights(weights [][]float64, consensus []float64, delta float64) [][]float64 {
	m := len(weights)               // Number of rows in the weight matrix
	n := len(weights[0])            // Number of columns in the weight matrix
	clipped := make([][]float64, m) // Initialize the clipped weight matrix

	for i := 0; i < m; i++ {
		clipped[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if j < len(consensus) {
				min := consensus[j] - delta // Calculate minimum weight value
				max := consensus[j] + delta // Calculate maximum weight value
				if weights[i][j] < min {    // If original weight is less than minimum
					clipped[i][j] = min
				} else if weights[i][j] > max { // If original weight is greater than maximum
					clipped[i][j] = max
				} else { // If original weight is between minimum and maximum
					clipped[i][j] = weights[i][j] // Otherwise, keep original weight
				}
			}
		}
	}

	return clipped // Return the clipped weight matrix
}

// computeBonds calculates EMA for Bonds
func (k Keeper) computeBonds(clippedWeights, prevBonds [][]float64, alpha float64) [][]float64 {
	m := len(clippedWeights)
	n := len(clippedWeights[0])
	bonds := make([][]float64, m)

	for i := 0; i < m; i++ {
		bonds[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			prevBond := 0.0
			if i < len(prevBonds) && j < len(prevBonds[i]) {
				prevBond = prevBonds[i][j]
			}
			bonds[i][j] = (1-alpha)*prevBond + alpha*clippedWeights[i][j]
		}
	}

	return bonds
}

// computeDividends calculates dividends
func (k Keeper) computeDividends(bonds [][]float64) []float64 {
	n := len(bonds[0])
	dividends := make([]float64, n)

	for j := 0; j < n; j++ {
		sum := 0.0
		for i := 0; i < len(bonds); i++ {
			if j < len(bonds[i]) {
				sum += bonds[i][j]
			}
		}
		dividends[j] = sum
	}

	return dividends
}

// normalizeDividends normalizes dividends
func (k Keeper) normalizeDividends(dividends []float64, active []bool) []float64 {
	sum := 0.0
	for i, d := range dividends {
		if active[i] {
			sum += d
		}
	}

	if sum == 0 {
		return make([]float64, len(dividends))
	}

	out := make([]float64, len(dividends))
	for i, d := range dividends {
		if active[i] {
			out[i] = d / sum
		}
	}
	return out
}

// distributeEmission distributes emission
func (k Keeper) distributeEmission(normIncentive, normDividends []float64, raoEmission cosmosmath.Int) []cosmosmath.Int {
	n := len(normIncentive)
	emission := make([]cosmosmath.Int, n)

	for i := 0; i < n; i++ {
		// 使用高精度计算
		incentiveDec := cosmosmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", normIncentive[i]))
		dividendsDec := cosmosmath.LegacyMustNewDecFromStr(fmt.Sprintf("%f", normDividends[i]))

		// 计算总份额
		totalShare := incentiveDec.Add(dividendsDec)

		// 计算奖励金额
		emission[i] = cosmosmath.NewIntFromBigInt(
			totalShare.MulInt(raoEmission).TruncateInt().BigInt(),
		)
	}

	return emission
}

// getPrevBonds gets previous epoch's bonds
func (k Keeper) getPrevBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) [][]float64 {
	n := len(validators)
	bonds := make([][]float64, n)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("bonds:"))

	// Get historical bonds from storage
	for i := 0; i < n; i++ {
		bonds[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// Create key: netuid:validator_i:validator_j
			key := fmt.Sprintf("%d:%s:%s", netuid, validators[i].Address, validators[j].Address)

			// Read bonds value from storage
			bz := store.Get([]byte(key))
			if bz != nil && len(bz) == 8 {
				// Convert binary back to float64
				bondValue := stdmath.Float64frombits(binary.BigEndian.Uint64(bz))
				bonds[i][j] = bondValue
			}
			// If no historical data found or invalid format, default to 0.0
		}
	}

	k.Logger(ctx).Debug("Retrieved previous bonds matrix",
		"netuid", netuid,
		"validators_count", n,
		"bonds_matrix_size", fmt.Sprintf("%dx%d", n, n),
	)

	return bonds
}

// saveBonds saves bonds to storage
// bonds are the exponential moving average (EMA) of historical weights, used for the next epoch calculation
func (k Keeper) saveBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo, bonds [][]float64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("bonds:"))

	// Save bonds data for each validator
	for i, validator := range validators {
		for j, bondValue := range bonds[i] {
			// Create key: netuid:validator_i:validator_j
			key := fmt.Sprintf("%d:%s:%s", netuid, validator.Address, validators[j].Address)

			// Store float64 as binary for efficiency
			bz := make([]byte, 8)
			binary.BigEndian.PutUint64(bz, stdmath.Float64bits(bondValue))
			store.Set([]byte(key), bz)
		}
	}

	k.Logger(ctx).Debug("Saved bonds matrix",
		"netuid", netuid,
		"validators_count", len(validators),
		"bonds_matrix_size", fmt.Sprintf("%dx%d", len(bonds), len(bonds[0])),
	)
}

// computeLiquidAlphaValues calculates dynamic alpha matrix
func (k Keeper) computeLiquidAlphaValues(weights [][]float64, bonds [][]float64, consensus []float64, params types.EpochParams) [][]float64 {
	m := len(weights)
	n := len(weights[0])
	alphas := make([][]float64, m)

	for i := 0; i < m; i++ {
		alphas[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			alphas[i][j] = k.alphaSigmoid(consensus[j], weights[i][j], bonds[i][j], params)
		}
	}
	return alphas
}

// alphaSigmoid calculates sigmoid alpha
func (k Keeper) alphaSigmoid(consensus, weight, bond float64, params types.EpochParams) float64 {
	diffBuy := k.clamp(weight-consensus, 0, 1)
	diffSell := k.clamp(bond-weight, 0, 1)
	combinedDiff := diffBuy
	if weight < bond {
		combinedDiff = diffSell
	}

	// Use a deterministic approximation for sigmoid
	x := params.AlphaSigmoidSteepness * (combinedDiff - 0.5)
	// Clamp x to avoid extreme values
	x = k.clamp(x, -10, 10)

	// Use a simple deterministic approximation
	var sigmoid float64
	if x < -2 {
		sigmoid = 0.1 // Near 0 for large negative x
	} else if x > 2 {
		sigmoid = 0.9 // Near 1 for large positive x
	} else {
		// Linear interpolation in the middle range
		sigmoid = 0.5 + x*0.2
	}

	alpha := params.AlphaLow + sigmoid*(params.AlphaHigh-params.AlphaLow)
	return k.clamp(alpha, params.AlphaLow, params.AlphaHigh)
}

// clamp limits value within a specified range
func (k Keeper) clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// computeBondsWithDynamicAlpha calculates bonds using dynamic alpha
func (k Keeper) computeBondsWithDynamicAlpha(weights, bonds [][]float64, alphas [][]float64) [][]float64 {
	m := len(weights)
	n := len(weights[0])
	result := make([][]float64, m)

	for i := 0; i < m; i++ {
		result[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			alpha := alphas[i][j]
			result[i][j] = (1-alpha)*bonds[i][j] + alpha*weights[i][j]
		}
	}
	return result
}

// computeDisabledLiquidAlpha calculates fixed alpha
func (k Keeper) computeDisabledLiquidAlpha(bondsMovingAverage float64) float64 {
	return bondsMovingAverage
}

// computeIncentive calculates incentive (using rho parameter)
func (k Keeper) computeIncentive(clippedWeights [][]float64, activeStake []float64, rho float64) []float64 {
	// Calculate ranks
	ranks := k.matMul(clippedWeights, activeStake)
	// Normalize
	normalized := k.normalize(ranks)
	// Apply rho parameter
	for i := range normalized {
		if normalized[i] > 0 {
			// Use deterministic power calculation based on common rho values
			if rho == 0.5 {
				// Square root - use Newton's method for determinism
				normalized[i] = k.deterministicSqrt(normalized[i])
			} else if rho == 2.0 {
				normalized[i] = normalized[i] * normalized[i]
			} else if rho == 1.0 {
				// No change needed for rho = 1
			} else {
				// For other values, use math.Pow but log a warning
				// Using standard math.Pow as fallback
				normalized[i] = stdmath.Pow(normalized[i], rho)
			}
		}
	}
	return k.normalize(normalized)
}

// matMul matrix multiplication
func (k Keeper) matMul(matrix [][]float64, vector []float64) []float64 {
	m := len(matrix)
	n := len(matrix[0])
	result := make([]float64, n)

	for j := 0; j < n; j++ {
		sum := 0.0
		for i := 0; i < m; i++ {
			if i < len(matrix) && j < len(matrix[i]) {
				sum += matrix[i][j] * vector[i]
			}
		}
		result[j] = sum
	}
	return result
}

// normalize normalizes array
func (k Keeper) normalize(arr []float64) []float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	if sum == 0 {
		return arr
	}
	normalized := make([]float64, len(arr))
	for i, v := range arr {
		normalized[i] = v / sum
	}
	return normalized
}

// normalizeIncentive normalizes incentive
func (k Keeper) normalizeIncentive(incentive []float64, active []bool) []float64 {
	return k.normalizeDividends(incentive, active) // Reuse the same logic
}

// Helper functions for debug logging
func (k Keeper) sumArray(arr []float64) float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum
}

func (k Keeper) minArray(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	min := arr[0]
	for _, v := range arr {
		if v < min {
			min = v
		}
	}
	return min
}

func (k Keeper) maxArray(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}

func (k Keeper) avgArray(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	return k.sumArray(arr) / float64(len(arr))
}

func (k Keeper) minMatrix(matrix [][]float64) float64 {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return 0
	}
	min := matrix[0][0]
	for _, row := range matrix {
		for _, v := range row {
			if v < min {
				min = v
			}
		}
	}
	return min
}

func (k Keeper) maxMatrix(matrix [][]float64) float64 {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return 0
	}
	max := matrix[0][0]
	for _, row := range matrix {
		for _, v := range row {
			if v > max {
				max = v
			}
		}
	}
	return max
}

func (k Keeper) avgMatrix(matrix [][]float64) float64 {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return 0
	}
	sum := 0.0
	count := 0
	for _, row := range matrix {
		for _, v := range row {
			sum += v
			count++
		}
	}
	return sum / float64(count)
}

func (k Keeper) sampleWeights(matrix [][]float64, sampleSize int) string {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return "empty"
	}
	sample := make([]float64, 0)
	for i := 0; i < min(sampleSize, len(matrix)); i++ {
		for j := 0; j < min(sampleSize, len(matrix[i])); j++ {
			sample = append(sample, matrix[i][j])
		}
	}
	return fmt.Sprintf("sample[%d]: %v", len(sample), sample)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// deterministicSqrt calculates square root in a deterministic way
func (k Keeper) deterministicSqrt(x float64) float64 {
	// Handle special cases
	if x < 0 {
		return 0 // Return 0 for negative inputs instead of NaN
	}
	if x == 0 || x == 1 {
		return x
	}

	// Newton's method for square root (deterministic)
	// Initial guess
	z := x / 2.0

	// Fixed number of iterations for determinism
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}

	return z
}
