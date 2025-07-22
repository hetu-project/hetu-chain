package keeper

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/stakework/types"
)

// 错误定义
var (
	ErrSubnetNotFound = errors.New("subnet not found")
	ErrEpochNotDue    = errors.New("epoch not due")
)

// RunEpoch 运行完整的 Bittensor epoch 算法
func (k Keeper) RunEpoch(ctx sdk.Context, netuid uint16, raoEmission uint64) (*types.EpochResult, error) {
	// 1. 获取子网数据
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return nil, ErrSubnetNotFound
	}

	// 2. 解析子网参数
	params := types.ParseEpochParams(subnet.Params)

	// 3. 检查是否应该运行 epoch
	if !k.shouldRunEpoch(ctx, netuid, params.Tempo) {
		return nil, ErrEpochNotDue
	}

	// 4. 获取子网的所有验证者数据
	validators := k.getSubnetValidators(ctx, netuid)
	if len(validators) == 0 {
		return &types.EpochResult{
			Netuid:    netuid,
			Accounts:  []string{},
			Emission:  []uint64{},
			Dividend:  []uint64{},
			Bonds:     [][]float64{},
			Consensus: []float64{},
		}, nil
	}

	// 5. 计算活跃状态（使用 activity_cutoff）
	active := k.calculateActive(ctx, netuid, validators, params)

	// 6. 获取质押权重
	stake := k.getStakeWeights(ctx, netuid, validators)

	// 7. 归一化质押权重
	activeStake := k.normalizeStake(stake, active)

	// 8. 获取权重矩阵
	weights := k.getWeightsMatrix(ctx, netuid, validators)

	// 9. 计算共识分数（加权中位数）
	consensus := k.weightedMedianCol(activeStake, weights, params.Kappa)

	// 10. 裁剪权重
	clippedWeights := k.clipWeights(weights, consensus, params.Delta)

	// 11. 计算 bonds（历史权重EMA）
	prevBonds := k.getPrevBonds(ctx, netuid, validators)
	var bonds [][]float64

	if params.LiquidAlphaEnabled {
		// 使用动态 alpha
		alphas := k.computeLiquidAlphaValues(weights, prevBonds, consensus, params)
		bonds = k.computeBondsWithDynamicAlpha(clippedWeights, prevBonds, alphas)
	} else {
		// 使用固定 alpha：使用 bonds_moving_average
		fixedAlpha := k.computeDisabledLiquidAlpha(params.BondsMovingAverage)
		bonds = k.computeBonds(clippedWeights, prevBonds, fixedAlpha)
	}

	// 12. 计算分红（dividends）
	dividends := k.computeDividends(bonds)

	// 13. 计算激励（incentive）- 使用 rho 参数
	incentive := k.computeIncentive(clippedWeights, activeStake, params.Rho)

	// 14. 归一化分红和激励
	normDividends := k.normalizeDividends(dividends, active)
	normIncentive := k.normalizeIncentive(incentive, active)

	// 15. 分配 emission（结合 incentive 和 dividends）
	emission := k.distributeEmission(normIncentive, normDividends, raoEmission)

	// 16. 构建结果
	result := &types.EpochResult{
		Netuid:    netuid,
		Accounts:  make([]string, len(validators)),
		Emission:  emission,
		Dividend:  make([]uint64, len(validators)),
		Incentive: make([]uint64, len(validators)),
		Bonds:     bonds,
		Consensus: consensus,
	}

	// 填充账户地址、分红和激励
	for i, validator := range validators {
		result.Accounts[i] = validator.Address
		result.Dividend[i] = uint64(normDividends[i] * float64(raoEmission))
		result.Incentive[i] = uint64(normIncentive[i] * float64(raoEmission))
	}

	// 17. 保存 bonds 到存储
	k.saveBonds(ctx, netuid, validators, bonds)

	// 18. 更新最后 epoch 时间
	// k.setLastEpochTime(ctx, netuid, ctx.BlockTime()) // 删除不再需要的函数

	return result, nil
}

// shouldRunEpoch 检查是否应该运行 epoch
// 按照 Bittensor 的真实实现：基于区块号的公式
// (block_number + netuid + 1) % (tempo + 1) == 0
func (k Keeper) shouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool {
	currentBlock := uint64(ctx.BlockHeight())

	// Bittensor 的 epoch 公式：
	// (block_number + netuid + 1) % (tempo + 1) == 0
	result := (currentBlock + uint64(netuid) + 1) % (tempo + 1)
	return result == 0
}

// ShouldRunEpoch 导出方法，满足 blockinflation/types.StakeworkKeeper 接口
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

// getSubnetValidators 获取子网的所有验证者
func (k Keeper) getSubnetValidators(ctx sdk.Context, netuid uint16) []types.ValidatorInfo {
	// 从 event 模块获取所有质押信息
	stakes := k.eventKeeper.GetAllValidatorStakesByNetuid(ctx, netuid)

	validators := make([]types.ValidatorInfo, 0, len(stakes))
	validatorMap := make(map[string]types.ValidatorInfo)

	// 处理质押信息
	for _, stake := range stakes {
		amount, _ := new(big.Int).SetString(stake.Amount, 10)
		stakeFloat := new(big.Float).SetInt(amount)
		stakeValue, _ := stakeFloat.Float64()

		validator := types.ValidatorInfo{
			Address: stake.Validator,
			Stake:   stakeValue,
			Weights: []uint64{},
			Active:  true, // 默认活跃，后续会更新
		}
		validatorMap[stake.Validator] = validator
	}

	// 处理权重信息
	for validatorAddr := range validatorMap {
		weight, found := k.eventKeeper.GetValidatorWeight(ctx, netuid, validatorAddr)
		if found {
			// 将 map 转换为数组
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

	// 转换为数组
	for _, validator := range validatorMap {
		validators = append(validators, validator)
	}

	return validators
}

// calculateActive 计算活跃状态
func (k Keeper) calculateActive(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo, params types.EpochParams) []bool {
	active := make([]bool, len(validators))
	currentBlock := uint64(ctx.BlockHeight())

	for i, validator := range validators {
		// 检查验证者是否在活跃截止时间内有活动
		// 这里需要从 event 模块获取最后活跃时间
		// 暂时使用简单的逻辑：所有验证者都设为活跃
		// TODO: 实现真正的活跃性检查逻辑
		lastActiveBlock := k.getLastActiveBlock(ctx, netuid, validator.Address)
		if currentBlock-lastActiveBlock <= params.ActivityCutoff {
			active[i] = true
		} else {
			active[i] = false
		}
	}

	return active
}

// getLastActiveBlock 获取验证者最后活跃区块（临时实现）
func (k Keeper) getLastActiveBlock(ctx sdk.Context, netuid uint16, validator string) uint64 {
	// TODO: 从 event 模块获取验证者的最后活跃时间
	// 暂时返回当前区块，表示所有验证者都是活跃的
	return uint64(ctx.BlockHeight())
}

// getStakeWeights 获取质押权重
func (k Keeper) getStakeWeights(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) []float64 {
	stake := make([]float64, len(validators))
	for i, validator := range validators {
		stake[i] = validator.Stake
	}
	return stake
}

// normalizeStake 归一化质押权重
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

// getWeightsMatrix 获取权重矩阵
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

// weightedMedianCol 加权中位数计算
/*
示例
假设有以下输入：

    stake = [100, 200, 300]：三个验证者的质押量。
    weights = [[0.1, 0.2, 0.3], [0.4, 0.5, 0.6], [0.7, 0.8, 0.9]]：三个验证者对三个节点的权重。
    kappa = 0.5：计算中位数的阈值。

我们期望的输出是每个节点的共识分数，即每个节点的加权中位数。
计算过程：

    对于节点 0：
        打分 = [1000.1, 2000.4, 300*0.7] = [10, 80, 210]
        排序后 = [10, 80, 210]
        加权中位数 = 80（因为 0.5 * (10+80+210) = 125，80 是第一个大于或等于 125 的数）
    对于节点 1：
        打分 = [1000.2, 2000.5, 300*0.8] = [20, 100, 240]
        排序后 = [20, 100, 240]
        加权中位数 = 100（因为 0.5 * (20+100+240) = 140，100 是第一个大于或等于 140 的数）
    对于节点 2：
        打分 = [1000.3, 2000.6, 300*0.9] = [30, 120, 270]
        排序后 = [30, 120, 270]
        加权中位数 = 120（因为 0.5 * (30+120+270) = 195，120 是第一个大于或等于 195 的数）

输出：

    consensus = [80, 100, 120]
*/
func (k Keeper) weightedMedianCol(stake []float64, weights [][]float64, kappa float64) []float64 {
	n := len(weights[0]) // 节点数
	m := len(stake)      // 验证者数
	consensus := make([]float64, n)

	for j := 0; j < n; j++ {
		// 收集所有验证者对节点j的打分
		type pair struct {
			w, s float64
		}
		var pairs []pair
		for i := 0; i < m; i++ {
			if i < len(weights) && j < len(weights[i]) {
				pairs = append(pairs, pair{w: weights[i][j], s: stake[i]})
			}
		}

		// 按分数升序排序
		sort.Slice(pairs, func(a, b int) bool {
			return pairs[a].w < pairs[b].w
		})

		// 计算加权中位数
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

// clipWeights 裁剪权重 防止单个节点的权重过大或过小
/*

假设有以下输入：

    weights = [[0.7, 0.8, 0.9], [0.4, 0.5, 0.6], [0.1, 0.2, 0.3]]：一个 3x3 的权重矩阵。
    consensus = [0.5, 0.6, 0.7]：每个节点的共识分数。
    delta = 0.1：权重调整的幅度。

计算过程：
对于每个节点 j 的权重，计算最大值 max 和最小值 min：

    对于节点 0：
        最大权重 max = 0.6 + 0.1 = 0.7
        最小权重 min = 0.5 - 0.1 = 0.4
        裁剪后的权重：[0.4, 0.6, 0.7]（原始权重 [0.7, 0.8, 0.9] 中的 0.8 和 0.9 超出了范围）
    对于节点 1：
        最大权重 max = 0.7 + 0.1 = 0.8
        最小权重 min = 0.6 - 0.1 = 0.5
        裁剪后的权重：[0.5, 0.5, 0.6]（所有权重都在范围内）
    对于节点 2：
        最大权重 max = 0.8 + 0.1 = 0.9
        最小权重 min = 0.7 - 0.1 = 0.6
        裁剪后的权重：[0.6, 0.7, 0.7]（原始权重 [0.1, 0.2, 0.3] 都在范围内）

输出：

    clipped = [[0.4, 0.6, 0.7], [0.5, 0.5, 0.6], [0.6, 0.7, 0.7]]
*/
func (k Keeper) clipWeights(weights [][]float64, consensus []float64, delta float64) [][]float64 {
	m := len(weights)               // 权重矩阵的行数
	n := len(weights[0])            // 权重矩阵的列数
	clipped := make([][]float64, m) // 初始化裁剪后的权重矩阵

	for i := 0; i < m; i++ {
		clipped[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if j < len(consensus) {
				min := consensus[j] - delta // 计算最小权重值
				max := consensus[j] + delta // 计算最大权重值
				if weights[i][j] < min {    // 如果原始权重小于最小值
					clipped[i][j] = min
				} else if weights[i][j] > max { // 如果原始权重大于最大值
					clipped[i][j] = max
				} else { // 如果原始权重在最小值和最大值之间
					clipped[i][j] = weights[i][j] // 否则保持原始权重
				}
			}
		}
	}

	return clipped //返回裁剪后的权重矩阵
}

// computeBonds Bonds 的 EMA 计算
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

// computeDividends 分红计算
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

// normalizeDividends 归一化分红
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

// distributeEmission 激励分配
func (k Keeper) distributeEmission(normIncentive, normDividends []float64, raoEmission uint64) []uint64 {
	n := len(normIncentive)
	emission := make([]uint64, n)

	for i := 0; i < n; i++ {
		emission[i] = uint64(normIncentive[i]*float64(raoEmission)) + uint64(normDividends[i]*float64(raoEmission))
	}

	return emission
}

// getPrevBonds 获取上一轮的 bonds
func (k Keeper) getPrevBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) [][]float64 {
	n := len(validators)
	bonds := make([][]float64, n)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("bonds:"))

	// 从存储中获取历史 bonds
	for i := 0; i < n; i++ {
		bonds[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			// 创建键：netuid:validator_i:validator_j
			key := fmt.Sprintf("%d:%s:%s", netuid, validators[i].Address, validators[j].Address)

			// 从存储中读取 bonds 值
			bz := store.Get([]byte(key))
			if bz != nil {
				// 将字符串转换回 float64
				if bondStr := string(bz); bondStr != "" {
					if bondValue, err := strconv.ParseFloat(bondStr, 64); err == nil {
						bonds[i][j] = bondValue
					}
				}
			}
			// 如果没有找到历史数据，默认为 0.0
		}
	}

	k.Logger(ctx).Debug("Retrieved previous bonds matrix",
		"netuid", netuid,
		"validators_count", n,
		"bonds_matrix_size", fmt.Sprintf("%dx%d", n, n),
	)

	return bonds
}

// saveBonds 保存 bonds 到存储
// bonds 是历史权重的指数移动平均（EMA），用于下一轮 epoch 计算
func (k Keeper) saveBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo, bonds [][]float64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("bonds:"))

	// 为每个验证者保存 bonds 数据
	for i, validator := range validators {
		for j, bondValue := range bonds[i] {
			// 创建键：netuid:validator_i:validator_j
			key := fmt.Sprintf("%d:%s:%s", netuid, validator.Address, validators[j].Address)

			// 将 float64 转换为字符串存储
			bondStr := fmt.Sprintf("%.10f", bondValue)
			store.Set([]byte(key), []byte(bondStr))
		}
	}

	k.Logger(ctx).Debug("Saved bonds matrix",
		"netuid", netuid,
		"validators_count", len(validators),
		"bonds_matrix_size", fmt.Sprintf("%dx%d", len(bonds), len(bonds[0])),
	)
}

// computeLiquidAlphaValues 计算动态 alpha 矩阵
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

// alphaSigmoid 计算 sigmoid alpha
func (k Keeper) alphaSigmoid(consensus, weight, bond float64, params types.EpochParams) float64 {
	diffBuy := k.clamp(weight-consensus, 0, 1)
	diffSell := k.clamp(bond-weight, 0, 1)
	combinedDiff := diffBuy
	if weight < bond {
		combinedDiff = diffSell
	}

	// sigmoid = 1 / (1 + exp(-steepness * (combined_diff - 0.5)))
	sigmoid := 1.0 / (1.0 + math.Exp(-params.AlphaSigmoidSteepness*(combinedDiff-0.5)))
	alpha := params.AlphaLow + sigmoid*(params.AlphaHigh-params.AlphaLow)
	return k.clamp(alpha, params.AlphaLow, params.AlphaHigh)
}

// clamp 限制值在指定范围内
func (k Keeper) clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// computeBondsWithDynamicAlpha 使用动态 alpha 计算 bonds
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

// computeDisabledLiquidAlpha 计算固定 alpha
func (k Keeper) computeDisabledLiquidAlpha(bondsMovingAverage float64) float64 {
	return 1.0 - bondsMovingAverage
}

// computeIncentive 计算激励（使用 rho 参数）
func (k Keeper) computeIncentive(clippedWeights [][]float64, activeStake []float64, rho float64) []float64 {
	// 计算 ranks
	ranks := k.matMul(clippedWeights, activeStake)
	// 归一化
	ranks = k.normalize(ranks)
	// 应用 rho 参数
	incentive := make([]float64, len(ranks))
	for i := range ranks {
		incentive[i] = ranks[i] * rho
	}
	return incentive
}

// matMul 矩阵乘法
func (k Keeper) matMul(weights [][]float64, stake []float64) []float64 {
	m := len(weights)
	n := len(weights[0])
	result := make([]float64, n)

	for j := 0; j < n; j++ {
		sum := 0.0
		for i := 0; i < m; i++ {
			if i < len(weights) && j < len(weights[i]) {
				sum += weights[i][j] * stake[i]
			}
		}
		result[j] = sum
	}
	return result
}

// normalize 归一化数组
func (k Keeper) normalize(values []float64) []float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}

	if sum == 0 {
		return make([]float64, len(values))
	}

	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = v / sum
	}
	return result
}

// normalizeIncentive 归一化激励
func (k Keeper) normalizeIncentive(incentive []float64, active []bool) []float64 {
	return k.normalizeDividends(incentive, active) // 复用相同逻辑
}
