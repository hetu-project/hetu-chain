package keeper

import (
	"fmt"
	stdmath "math"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"math/big"

	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
	stakeworktypes "github.com/hetu-project/hetu/v1/x/stakework/types"
)

// CalculateAlphaEmission calculates the Alpha emission for a subnet based on its Alpha issuance
// This uses the same logarithmic decay algorithm as CalculateBlockEmission
// Improved version with high-precision calculations to avoid floating-point precision issues
func (k Keeper) CalculateAlphaEmission(ctx sdk.Context, netuid uint16) (math.Int, error) {
	params := k.GetParams(ctx)

	// Get subnet Alpha issuance: SubnetAlphaIn + SubnetAlphaOut
	subnetAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
	subnetAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)

	// 添加详细日志
	k.Logger(ctx).Debug("Alpha issuance components",
		"netuid", netuid,
		"subnet_alpha_in", subnetAlphaIn.String(),
		"subnet_alpha_out", subnetAlphaOut.String())

	alphaIssuance := subnetAlphaIn.Add(subnetAlphaOut)

	// Check if we have any Alpha issuance
	if !alphaIssuance.IsPositive() {
		k.Logger(ctx).Debug("Alpha issuance is zero or negative, returning zero emission",
			"netuid", netuid,
			"alpha_issuance", alphaIssuance.String())
		return math.ZeroInt(), nil
	}

	// Use high-precision math.LegacyDec calculations instead of float64
	alphaIssuanceDec := alphaIssuance.ToLegacyDec()
	totalSupplyDec := params.TotalSupply.ToLegacyDec()
	defaultBlockEmissionDec := params.DefaultBlockEmission.ToLegacyDec()

	// 添加详细日志
	k.Logger(ctx).Debug("Alpha emission calculation parameters",
		"netuid", netuid,
		"alpha_issuance_dec", alphaIssuanceDec.String(),
		"total_supply_dec", totalSupplyDec.String(),
		"default_block_emission_dec", defaultBlockEmissionDec.String())

	// Calculate the ratio: alpha_issuance / (2 * total_supply)
	twoTimesTotalSupply := totalSupplyDec.Mul(math.LegacyNewDec(2))
	ratio := alphaIssuanceDec.Quo(twoTimesTotalSupply)

	// 添加详细日志
	k.Logger(ctx).Debug("Alpha emission ratio calculation",
		"netuid", netuid,
		"two_times_total_supply", twoTimesTotalSupply.String(),
		"ratio", ratio.String())

	// If ratio >= 1.0, return 0
	if ratio.GTE(math.LegacyOneDec()) {
		k.Logger(ctx).Debug("Ratio >= 1.0, returning zero emission",
			"netuid", netuid,
			"ratio", ratio.String())
		return math.ZeroInt(), nil
	}

	// Calculate log2(1 / (1 - ratio)) using high-precision arithmetic
	// logArg = 1 / (1 - ratio)
	oneMinusRatio := math.LegacyOneDec().Sub(ratio)
	if oneMinusRatio.LTE(math.LegacyZeroDec()) {
		k.Logger(ctx).Debug("1 - ratio <= 0, returning zero emission",
			"netuid", netuid,
			"one_minus_ratio", oneMinusRatio.String())
		return math.ZeroInt(), nil
	}

	logArg := math.LegacyOneDec().Quo(oneMinusRatio)

	// 添加详细日志
	k.Logger(ctx).Debug("Log argument calculation",
		"netuid", netuid,
		"one_minus_ratio", oneMinusRatio.String(),
		"log_arg", logArg.String())

	// Convert to float64 for log2 calculation (this is the only place we need float64)
	logArgFloat := logArg.MustFloat64()
	logResult := stdmath.Log2(logArgFloat)

	// Floor the log result
	flooredLog := stdmath.Floor(logResult)
	flooredLogInt := int64(flooredLog)

	// 添加详细日志
	k.Logger(ctx).Debug("Log calculation",
		"netuid", netuid,
		"log_arg_float", logArgFloat,
		"log_result", logResult,
		"floored_log", flooredLogInt)

	// Calculate 2^flooredLog
	multiplier := stdmath.Pow(2.0, float64(flooredLogInt))

	// Calculate block emission percentage: 1 / multiplier
	blockEmissionPercentage := math.LegacyOneDec().Quo(math.LegacyNewDecWithPrec(int64(multiplier*1000), 3))

	// 添加详细日志
	k.Logger(ctx).Debug("Emission percentage calculation",
		"netuid", netuid,
		"multiplier", multiplier,
		"block_emission_percentage", blockEmissionPercentage.String())

	// Calculate actual Alpha emission using high-precision arithmetic
	alphaEmission := defaultBlockEmissionDec.Mul(blockEmissionPercentage)

	// Convert back to math.Int with proper rounding
	alphaEmissionInt := alphaEmission.TruncateInt()

	k.Logger(ctx).Debug("calculated Alpha emission (high-precision)",
		"netuid", netuid,
		"alpha_issuance", alphaIssuance.String(),
		"total_supply", params.TotalSupply.String(),
		"ratio", ratio.String(),
		"log_arg", logArg.String(),
		"log_result", fmt.Sprintf("%.6f", logResult),
		"floored_log", flooredLogInt,
		"multiplier", fmt.Sprintf("%.6f", multiplier),
		"emission_percentage", blockEmissionPercentage.String(),
		"alpha_emission", alphaEmissionInt.String(),
	)

	return alphaEmissionInt, nil
}

// RunCoinbase executes the coinbase logic for distributing rewards to subnets
// This is equivalent to the run_coinbase.rs function
func (k Keeper) RunCoinbase(ctx sdk.Context, blockEmission math.Int) error {
	// --- 0. Get current block
	currentBlock := ctx.BlockHeight()
	k.Logger(ctx).Debug("Current block", "block", currentBlock)
	// --- 1. Get all netuids (filter out root)
	allSubnets := k.eventKeeper.GetAllSubnetNetuids(ctx)
	k.Logger(ctx).Debug("All subnet netuids", "subnets", allSubnets)

	// Filter out subnets with no first emission block number
	subnetsToEmitTo := k.eventKeeper.GetSubnetsToEmitTo(ctx)
	k.Logger(ctx).Debug("Subnets to emit to", "subnets", subnetsToEmitTo)

	// If no subnets to emit to, return early
	if len(subnetsToEmitTo) == 0 {
		k.Logger(ctx).Info("No subnets to emit to, skipping coinbase")
		return nil
	}

	// --- 2. Get sum of moving prices
	totalMovingPrices := math.LegacyZeroDec()
	// Only get price EMA for subnets that we emit to
	for _, netuid := range subnetsToEmitTo {
		// Get and update the moving price of each subnet adding the total together
		movingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)
		totalMovingPrices = totalMovingPrices.Add(movingPrice)

		k.Logger(ctx).Debug("Subnet moving price",
			"netuid", netuid,
			"moving_price", movingPrice.String(),
		)
	}
	k.Logger(ctx).Debug("Total moving prices", "total", totalMovingPrices)

	// --- 3. Calculate subnet terms (tao_in, alpha_in, alpha_out)
	rewards, err := k.CalculateSubnetRewards(ctx, blockEmission, subnetsToEmitTo)
	if err != nil {
		k.Logger(ctx).Error("failed to calculate subnet rewards", "error", err)
		return err
	}

	// --- 4. Injection - Add rewards to subnet pools
	// Apply the calculated rewards (alpha_in, alpha_out, tao_in)
	if err := k.ApplySubnetRewards(ctx, rewards); err != nil {
		k.Logger(ctx).Error("failed to apply subnet rewards", "error", err)
		return err
	}

	// --- 新增: 同步链上状态到合约，只注入HETU代币
	for netuid, reward := range rewards {
		if err := k.SyncChainStateToContract(ctx, netuid, reward); err != nil {
			k.Logger(ctx).Error("Failed to sync chain state to contract",
				"netuid", netuid,
				"error", err,
			)
			// 继续处理其他子网，不要因为同步失败而中断整个流程
		}
	}

	// --- 5. Calculate owner cuts and update alpha_out
	// Calculate owner cuts and subtract from alpha_out
	if err := k.CalculateOwnerCuts(ctx, rewards); err != nil {
		k.Logger(ctx).Error("failed to calculate owner cuts", "error", err)
		return err
	}

	// --- 6. Add alpha_out to pending emission for each subnet
	// Since there's no root subnet, add alpha_out to pending emission for each subnet
	if err := k.AddToPendingEmission(ctx, rewards); err != nil {
		k.Logger(ctx).Error("failed to add to pending emission", "error", err)
		return err
	}

	// --- 7. Update moving prices after using them in the emission calculation
	// Only update price EMA for subnets that we emit to
	for _, netuid := range subnetsToEmitTo {
		// Get subnet to access EMAPriceHalvingBlocks
		subnet, exists := k.eventKeeper.GetSubnet(ctx, netuid)
		if !exists {
			k.Logger(ctx).Error("subnet not found for moving price update", "netuid", netuid)
			continue
		}

		// Get moving alpha from blockinflation params
		params := k.GetParams(ctx)
		movingAlpha := params.SubnetMovingAlpha
		halvingBlocks := subnet.EMAPriceHalvingBlocks

		// Update moving prices after using them above
		k.eventKeeper.UpdateMovingPrice(ctx, netuid, movingAlpha, halvingBlocks)
	}
	k.Logger(ctx).Debug("Moving prices updated")

	// --- 7. Drain pending emission through the subnet based on tempo (epoch)
	for _, netuid := range subnetsToEmitTo {
		subnet, exists := k.eventKeeper.GetSubnet(ctx, netuid)
		if !exists {
			k.Logger(ctx).Error("subnet not found", "netuid", netuid)
			continue
		}
		params, err := stakeworktypes.ParseEpochParams(subnet.Params)
		if err != nil {
			k.Logger(ctx).Error("failed to parse epoch parameters", "netuid", netuid, "err", err)
			continue
		}
		tempo := params.Tempo
		if k.stakeworkKeeper.ShouldRunEpoch(ctx, netuid, tempo) {
			k.Logger(ctx).Debug("Epoch triggered", "netuid", netuid, "block", ctx.BlockHeight())

			// Reset counters
			k.eventKeeper.SetBlocksSinceLastStep(ctx, netuid, 0)
			k.eventKeeper.SetLastMechanismStepBlock(ctx, netuid, ctx.BlockHeight())

			// Extract and clear pending emission and owner cut
			pendingAlpha := k.eventKeeper.GetPendingEmission(ctx, netuid)
			k.eventKeeper.SetPendingEmission(ctx, netuid, math.ZeroInt())
			ownerCut := k.eventKeeper.GetPendingOwnerCut(ctx, netuid)
			k.eventKeeper.SetPendingOwnerCut(ctx, netuid, math.ZeroInt())

			k.Logger(ctx).Debug("Draining pending emission", "netuid", netuid, "pending_alpha", pendingAlpha.String(), "owner_cut", ownerCut.String())

			// Run epoch consensus
			epochResult, err := k.stakeworkKeeper.RunEpoch(ctx, netuid, pendingAlpha)
			if err != nil {
				k.Logger(ctx).Error("RunEpoch failed", "netuid", netuid, "error", err)
				continue
			}

			// Calculate incentive sum
			incentiveSum := math.ZeroInt()
			for _, v := range epochResult.Incentive {
				incentiveSum = incentiveSum.Add(v)
			}

			// Calculate validator-allocatable alpha
			var pendingValidatorAlpha math.Int
			if !incentiveSum.IsZero() {
				// 分配一半给验证者
				pendingValidatorAlpha = pendingAlpha.QuoRaw(2)
			} else {
				// 全部分配给验证者
				pendingValidatorAlpha = pendingAlpha
			}

			// Build dividend account list
			dividendAccounts := make([]string, len(epochResult.Accounts))
			copy(dividendAccounts, epochResult.Accounts)

			// Get subnet stake amounts
			stakeMap := k.getStakeMap(ctx, netuid, dividendAccounts)
			k.Logger(ctx).Debug("Stake map", "netuid", netuid, "stake_map", stakeMap)

			// Calculate dividends
			dividends := make(map[string]math.Int)
			for i, addr := range epochResult.Accounts {
				dividends[addr] = epochResult.Dividend[i]
			}
			totalAlphaDivs := math.ZeroInt()
			for _, v := range dividends {
				totalAlphaDivs = totalAlphaDivs.Add(v)
			}

			// Dividend allocation (no parent-child relationship, direct allocation, weight by subnet stake)
			alphaDividends := make(map[string]math.Int)
			if !totalAlphaDivs.IsZero() && !pendingValidatorAlpha.IsZero() {
				// 按比例分配
				for addr, d := range dividends {
					// 计算分配比例: d / totalAlphaDivs * pendingValidatorAlpha
					ratio := math.LegacyNewDecFromInt(d).QuoInt(totalAlphaDivs)
					alloc := ratio.MulInt(pendingValidatorAlpha).TruncateInt()
					alphaDividends[addr] = alloc
				}

				// 检查是否有剩余未分配的奖励
				allocated := math.ZeroInt()
				for _, amount := range alphaDividends {
					allocated = allocated.Add(amount)
				}

				// 如果有剩余，按地址顺序分配
				if allocated.LT(pendingValidatorAlpha) {
					remainder := pendingValidatorAlpha.Sub(allocated)
					addrs := make([]string, 0, len(dividends))
					for a := range dividends {
						addrs = append(addrs, a)
					}
					sort.Strings(addrs)

					// 每个地址分配1个单位，直到分完
					oneUnit := math.NewInt(1)
					for i := 0; i < len(addrs) && !remainder.IsZero(); i++ {
						alphaDividends[addrs[i]] = alphaDividends[addrs[i]].Add(oneUnit)
						remainder = remainder.Sub(oneUnit)
					}
				}
			}
			k.Logger(ctx).Debug("Alpha dividends", "netuid", netuid, "alpha_dividends", alphaDividends)

			// Incentive allocation
			incentives := make(map[string]math.Int)
			for i, addr := range epochResult.Accounts {
				incentives[addr] = epochResult.Incentive[i]
			}
			k.Logger(ctx).Debug("Incentives", "netuid", netuid, "incentives", incentives)

			// Log distribution
			for addr, amount := range alphaDividends {
				k.Logger(ctx).Info("Alpha dividend distributed", "netuid", netuid, "account", addr, "amount", amount.String())
				if err := k.MintAlphaTokens(ctx, netuid, addr, amount.BigInt()); err != nil {
					k.Logger(ctx).Error("Failed to mint alpha tokens for validator dividend",
						"netuid", netuid,
						"address", addr,
						"amount", amount.String(),
						"error", err,
					)
				}
			}
			for addr, amount := range incentives {
				if err := k.MintAlphaTokens(ctx, netuid, addr, amount.BigInt()); err != nil {
					k.Logger(ctx).Error("Failed to mint alpha tokens for incentives",
						"netuid", netuid,
						"address", addr,
						"amount", amount.String(),
						"error", err,
					)
				}
			}
			if ownerCut.IsPositive() {
				k.Logger(ctx).Info("Owner cut distributed", "netuid", netuid, "amount", ownerCut.String())
				if err := k.MintAlphaTokens(ctx, netuid, subnet.Owner, ownerCut.BigInt()); err != nil {
					k.Logger(ctx).Error("Failed to mint alpha tokens for subnet owner",
						"netuid", netuid,
						"owner", subnet.Owner,
						"amount", ownerCut.String(),
						"error", err,
					)
				}
			}
		} else {
			blocks := k.eventKeeper.GetBlocksSinceLastStep(ctx, netuid)
			k.eventKeeper.SetBlocksSinceLastStep(ctx, netuid, blocks+1)
			k.Logger(ctx).Debug("Not epoch, increment counter", "netuid", netuid, "blocks_since_last", blocks+1)
		}
	}

	// --- 8. Drain pending emission (placeholder)
	// TODO: Implement epoch-based emission draining
	k.Logger(ctx).Debug("Pending emission drained")

	// --- 9. 同步 AMM 池状态
	for _, netuid := range subnetsToEmitTo {
		if err := k.SyncAMMPoolState(ctx, netuid); err != nil {
			k.Logger(ctx).Error("Failed to sync AMM pool state",
				"netuid", netuid,
				"error", err,
			)
		}
	}

	// Emit event for coinbase execution
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"coinbase_executed",
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", currentBlock)),
			sdk.NewAttribute("block_emission", blockEmission.String()),
			sdk.NewAttribute("subnets_count", fmt.Sprintf("%d", len(subnetsToEmitTo))),
			sdk.NewAttribute("total_moving_prices", totalMovingPrices.String()),
		),
	)

	k.Logger(ctx).Info("Coinbase executed successfully",
		"block", currentBlock,
		"emission", blockEmission.String(),
		"subnets", len(subnetsToEmitTo),
	)

	return nil
}

// getStakeMap retrieves stake amount for each account in this subnet (no parent subnet)
func (k Keeper) getStakeMap(ctx sdk.Context, netuid uint16, accounts []string) map[string]uint64 {
	stakeMap := map[string]uint64{}
	stakes := k.eventKeeper.GetAllValidatorStakesByNetuid(ctx, netuid)
	for _, acc := range accounts {
		stakeMap[acc] = 0
		for _, s := range stakes {
			if s.Validator == acc {
				if amount, ok := new(big.Int).SetString(s.Amount, 10); ok {
					stakeMap[acc] += amount.Uint64()
				} else {
					k.Logger(ctx).Error("invalid stake amount string",
						"netuid", netuid,
						"validator", acc,
						"amount", s.Amount)
				}
			}
		}
	}
	return stakeMap
}

// GetSubnetEmissionData returns emission data for a specific subnet
// This is a helper function for testing and debugging
func (k Keeper) GetSubnetEmissionData(ctx sdk.Context, netuid uint16) (blockinflationtypes.SubnetEmissionData, error) {
	// Check if subnet exists
	_, exists := k.eventKeeper.GetSubnetFirstEmissionBlock(ctx, netuid)
	if !exists {
		return blockinflationtypes.SubnetEmissionData{}, fmt.Errorf("subnet %d not found", netuid)
	}

	// Get current block emission for calculation
	blockEmission, err := k.CalculateBlockEmission(ctx)
	if err != nil {
		return blockinflationtypes.SubnetEmissionData{}, fmt.Errorf("failed to calculate block emission: %w", err)
	}

	// Calculate rewards for this specific subnet
	rewards, err := k.CalculateSubnetRewards(ctx, blockEmission, []uint16{netuid})
	if err != nil {
		return blockinflationtypes.SubnetEmissionData{}, fmt.Errorf("failed to calculate rewards: %w", err)
	}

	reward, exists := rewards[netuid]
	if !exists {
		return blockinflationtypes.SubnetEmissionData{}, fmt.Errorf("no reward data for subnet %d", netuid)
	}

	return blockinflationtypes.SubnetEmissionData{
		Netuid:                 netuid,
		TaoIn:                  reward.TaoIn,
		AlphaIn:                reward.AlphaIn,
		AlphaOut:               reward.AlphaOut,
		OwnerCut:               reward.OwnerCut,
		RootDivs:               math.ZeroInt(), // No root dividends - no root subnet
		SubnetAlphaInEmission:  k.eventKeeper.GetSubnetAlphaInEmission(ctx, netuid),
		SubnetAlphaOutEmission: k.eventKeeper.GetSubnetAlphaOutEmission(ctx, netuid),
		SubnetTaoInEmission:    k.eventKeeper.GetSubnetTaoInEmission(ctx, netuid),
	}, nil
}

// GetAllSubnetEmissionData returns emission data for all subnets
func (k Keeper) GetAllSubnetEmissionData(ctx sdk.Context) []blockinflationtypes.SubnetEmissionData {
	subnets := k.eventKeeper.GetSubnetsToEmitTo(ctx)
	if len(subnets) == 0 {
		return []blockinflationtypes.SubnetEmissionData{}
	}

	// Get current block emission for calculation
	blockEmission, err := k.CalculateBlockEmission(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to calculate block emission", "error", err)
		return []blockinflationtypes.SubnetEmissionData{}
	}

	// Calculate rewards for all subnets
	rewards, err := k.CalculateSubnetRewards(ctx, blockEmission, subnets)
	if err != nil {
		k.Logger(ctx).Error("failed to calculate rewards", "error", err)
		return []blockinflationtypes.SubnetEmissionData{}
	}

	var data []blockinflationtypes.SubnetEmissionData
	for _, netuid := range subnets {
		if reward, exists := rewards[netuid]; exists {
			data = append(data, blockinflationtypes.SubnetEmissionData{
				Netuid:                 netuid,
				TaoIn:                  reward.TaoIn,
				AlphaIn:                reward.AlphaIn,
				AlphaOut:               reward.AlphaOut,
				OwnerCut:               reward.OwnerCut,
				RootDivs:               math.ZeroInt(), // No root dividends - no root subnet
				SubnetAlphaInEmission:  k.eventKeeper.GetSubnetAlphaInEmission(ctx, netuid),
				SubnetAlphaOutEmission: k.eventKeeper.GetSubnetAlphaOutEmission(ctx, netuid),
				SubnetTaoInEmission:    k.eventKeeper.GetSubnetTaoInEmission(ctx, netuid),
			})
		}
	}

	// Sort by netuid for consistent output
	sort.Slice(data, func(i, j int) bool {
		return data[i].Netuid < data[j].Netuid
	})

	return data
}
