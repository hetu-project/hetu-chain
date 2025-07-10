package keeper

import (
	"errors"
	"math/big"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
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

	// 5. 计算活跃状态
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
	bonds := k.computeBonds(clippedWeights, prevBonds, params.Alpha)

	// 12. 计算分红（dividends）
	dividends := k.computeDividends(bonds)

	// 13. 归一化分红
	normDividends := k.normalizeDividends(dividends, active)

	// 14. 分配 emission
	emission := k.distributeEmission(normDividends, raoEmission)

	// 15. 构建结果
	result := &types.EpochResult{
		Netuid:    netuid,
		Accounts:  make([]string, len(validators)),
		Emission:  emission,
		Dividend:  make([]uint64, len(validators)),
		Bonds:     bonds,
		Consensus: consensus,
	}

	// 填充账户地址和分红
	for i, validator := range validators {
		result.Accounts[i] = validator.Address
		result.Dividend[i] = uint64(normDividends[i] * float64(raoEmission))
	}

	// 16. 保存 bonds 到存储
	k.saveBonds(ctx, netuid, validators, bonds)

	// 17. 更新最后 epoch 时间
	k.setLastEpochBlock(ctx, netuid, uint64(ctx.BlockHeight()))

	return result, nil
}

// shouldRunEpoch 检查是否应该运行 epoch
func (k Keeper) shouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool {
	lastEpoch := k.getLastEpochBlock(ctx, netuid)
	currentBlock := uint64(ctx.BlockHeight())
	return (currentBlock - lastEpoch) >= tempo
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

	// TODO: 这里需要从 event 模块获取最后活跃时间
	// 暂时所有验证者都设为活跃
	for i := range validators {
		active[i] = true
	}

	return active
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

// clipWeights 裁剪权重
func (k Keeper) clipWeights(weights [][]float64, consensus []float64, delta float64) [][]float64 {
	m := len(weights)
	n := len(weights[0])
	clipped := make([][]float64, m)

	for i := 0; i < m; i++ {
		clipped[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if j < len(consensus) {
				min := consensus[j] - delta
				max := consensus[j] + delta
				if weights[i][j] < min {
					clipped[i][j] = min
				} else if weights[i][j] > max {
					clipped[i][j] = max
				} else {
					clipped[i][j] = weights[i][j]
				}
			}
		}
	}

	return clipped
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
func (k Keeper) distributeEmission(normDividends []float64, raoEmission uint64) []uint64 {
	n := len(normDividends)
	emission := make([]uint64, n)

	for i := 0; i < n; i++ {
		emission[i] = uint64(normDividends[i] * float64(raoEmission))
	}

	return emission
}

// getPrevBonds 获取上一轮的 bonds
func (k Keeper) getPrevBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo) [][]float64 {
	n := len(validators)
	bonds := make([][]float64, n)

	// TODO: 从存储中获取历史 bonds
	// 暂时返回零矩阵
	for i := 0; i < n; i++ {
		bonds[i] = make([]float64, n)
	}

	return bonds
}

// saveBonds 保存 bonds 到存储
func (k Keeper) saveBonds(ctx sdk.Context, netuid uint16, validators []types.ValidatorInfo, bonds [][]float64) {
	// TODO: 保存 bonds 到存储
	// 这里可以保存到 KVStore 中
}

// getLastEpochBlock 获取最后 epoch 区块
func (k Keeper) getLastEpochBlock(ctx sdk.Context, netuid uint16) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := []byte("last_epoch_" + strconv.FormatUint(uint64(netuid), 10))

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	// 简单的字节转换
	var lastBlock uint64
	for i, b := range bz {
		if i < 8 {
			lastBlock |= uint64(b) << (i * 8)
		}
	}
	return lastBlock
}

// setLastEpochBlock 设置最后 epoch 区块
func (k Keeper) setLastEpochBlock(ctx sdk.Context, netuid uint16, blockHeight uint64) {
	store := ctx.KVStore(k.storeKey)
	key := []byte("last_epoch_" + strconv.FormatUint(uint64(netuid), 10))

	// 简单的字节转换
	bz := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bz[i] = byte(blockHeight >> (i * 8))
	}

	store.Set(key, bz)
}
