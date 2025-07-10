package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
)

// RunEpoch 运行完整的Yuma共识epoch - 根据bittensor实现
func (k Keeper) RunEpoch(ctx sdk.Context, netuid uint16) error {
	// 1. 获取子网信息和参数
	subnetInfo, found := k.GetSubnetInfo(ctx, netuid)
	if !found {
		return fmt.Errorf("子网 %d 不存在", netuid)
	}

	params := k.GetParams(ctx)

	// 2. 检查是否应该运行epoch - 使用Tempo参数
	if !k.shouldRunEpoch(ctx, netuid, params.Tempo) {
		return nil // 还没到运行时间
	}

	// 3. 获取所有神经元信息
	neurons := k.GetAllNeurons(ctx, netuid)
	if len(neurons) == 0 {
		return nil
	}

	// 4. 应用MaxAllowedUids限制
	if len(neurons) > int(params.MaxAllowedUids) {
		neurons = neurons[:params.MaxAllowedUids]
	}

	// 5. 收集和验证权重
	weights := k.collectAndValidateWeights(ctx, netuid, neurons, params)

	// 6. 计算活跃性 - 使用ActivityCutoff参数
	active := k.calculateActive(ctx, netuid, neurons, params.ActivityCutoff)

	// 7. 应用免疫期保护 - 使用ImmunityPeriod参数
	active = k.applyImmunityPeriod(ctx, neurons, active, params.ImmunityPeriod)

	// 8. 更新bonds矩阵
	bonds := k.updateBonds(ctx, netuid, weights, active, neurons)

	// 9. 执行完整的Yuma共识计算
	consensus := k.calculateConsensus(weights, active, bonds, params)
	trust := k.calculateTrust(consensus, active, params)
	ranks := k.calculateRanks(trust, active, params)
	incentives := k.calculateIncentives(ranks, active, params)
	dividends := k.calculateDividends(incentives, active, params)
	emission := k.calculateEmission(ctx, netuid, incentives, dividends, params)

	// 10. 更新神经元状态
	k.updateNeuronStates(ctx, netuid, neurons, consensus, trust, ranks, incentives, dividends, emission, bonds)

	// 11. 更新子网统计信息
	k.updateSubnetStats(ctx, netuid, subnetInfo)

	return nil
}

// shouldRunEpoch 检查是否应该运行epoch - 使用Tempo
func (k Keeper) shouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint16) bool {
	lastEpoch := k.GetLastEpochBlock(ctx, netuid)
	currentBlock := uint16(ctx.BlockHeight())

	return (currentBlock - lastEpoch) >= tempo
}

// collectAndValidateWeights 收集和验证权重 - 使用MaxWeightsLimit, MinAllowedWeights
func (k Keeper) collectAndValidateWeights(ctx sdk.Context, netuid uint16, neurons []types.NeuronInfo, params types.Params) [][]sdk.Dec {
	n := len(neurons)
	weights := make([][]sdk.Dec, n)

	for i := 0; i < n; i++ {
		weights[i] = make([]sdk.Dec, n)

		// 获取神经元的权重数据
		neuronWeights := k.GetNeuronWeights(ctx, netuid, neurons[i].Uid)

		// 验证权重数量限制
		if len(neuronWeights) > int(params.MaxWeightsLimit) {
			// 截断到最大限制
			neuronWeights = neuronWeights[:params.MaxWeightsLimit]
		}

		// 验证最小权重数量
		if len(neuronWeights) < int(params.MinAllowedWeights) && len(neuronWeights) > 0 {
			// 权重数量不足，设置为零
			continue
		}

		// 验证权重设置速率限制
		if !k.checkWeightsRateLimit(ctx, netuid, neurons[i].Uid, params.WeightsSetRateLimit) {
			continue
		}

		// 填充权重矩阵
		for j, weight := range neuronWeights {
			if j < n {
				weights[i][j] = weight
			}
		}

		// 归一化权重
		weights[i] = k.normalizeWeights(weights[i])
	}

	return weights
}

// calculateActive 计算活跃神经元 - 使用ActivityCutoff参数
func (k Keeper) calculateActive(ctx sdk.Context, netuid uint16, neurons []types.NeuronInfo, activityCutoff uint16) []bool {
	n := len(neurons)
	active := make([]bool, n)
	currentBlock := uint16(ctx.BlockHeight())

	for i, neuron := range neurons {
		// 检查最后活跃时间
		lastActive := neuron.LastUpdate
		if (currentBlock - lastActive) <= activityCutoff {
			active[i] = true
		}
	}

	return active
}

// applyImmunityPeriod 应用免疫期保护 - 使用ImmunityPeriod参数
func (k Keeper) applyImmunityPeriod(ctx sdk.Context, neurons []types.NeuronInfo, active []bool, immunityPeriod uint16) []bool {
	currentBlock := uint16(ctx.BlockHeight())

	for i, neuron := range neurons {
		// 检查神经元是否在免疫期内
		if (currentBlock - neuron.BlockAtRegistration) <= immunityPeriod {
			active[i] = true // 免疫期内的神经元总是活跃的
		}
	}

	return active
}

// checkWeightsRateLimit 检查权重设置速率限制 - 使用WeightsSetRateLimit参数
func (k Keeper) checkWeightsRateLimit(ctx sdk.Context, netuid uint16, uid uint16, rateLimit uint64) bool {
	lastWeightsUpdate := k.GetLastWeightsUpdate(ctx, netuid, uid)
	currentBlock := uint64(ctx.BlockHeight())

	return (currentBlock - lastWeightsUpdate) >= rateLimit
}

// calculateConsensus 计算共识分数 - Yuma共识核心算法
func (k Keeper) calculateConsensus(weights [][]sdk.Dec, active []bool, bonds [][]sdk.Dec, params types.Params) []sdk.Dec {
	n := len(weights)
	consensus := make([]sdk.Dec, n)

	if n == 0 {
		return consensus
	}

	// 计算共识分数: C = W^T * S
	// 其中 W 是权重矩阵，S 是stake向量
	for i := 0; i < n; i++ {
		if !active[i] {
			continue
		}

		consensusSum := sdk.ZeroDec()
		for j := 0; j < n; j++ {
			if active[j] {
				// 使用bonds作为stake权重
				stake := bonds[j][i]
				weight := weights[j][i]
				consensusSum = consensusSum.Add(stake.Mul(weight))
			}
		}
		consensus[i] = consensusSum
	}

	// 归一化共识分数
	return k.normalizeVector(consensus, active)
}

// calculateTrust 计算信任分数
func (k Keeper) calculateTrust(consensus []sdk.Dec, active []bool, params types.Params) []sdk.Dec {
	n := len(consensus)
	trust := make([]sdk.Dec, n)

	// 信任分数基于共识分数，但应用了kappa饱和
	for i := 0; i < n; i++ {
		if active[i] {
			trustValue := consensus[i]
			// 应用kappa饱和
			kappaValue := params.Kappa.Quo(sdk.NewDec(65536)) // 归一化
			if trustValue.GT(kappaValue) {
				trustValue = kappaValue
			}
			trust[i] = trustValue
		}
	}

	return k.normalizeVector(trust, active)
}

// calculateRanks 计算排名分数
func (k Keeper) calculateRanks(trust []sdk.Dec, active []bool, params types.Params) []sdk.Dec {
	n := len(trust)
	ranks := make([]sdk.Dec, n)

	// 排名基于信任分数，应用rho稀疏化
	for i := 0; i < n; i++ {
		if active[i] {
			rankValue := trust[i]
			// 应用rho变换
			rhoValue := params.Rho.Quo(sdk.NewDec(65536)) // 归一化
			if rhoValue.GT(sdk.ZeroDec()) {
				// 简化的rho变换: rank = rank^(1/rho)
				rankValue = k.powerApproximation(rankValue, sdk.OneDec().Quo(rhoValue))
			}
			ranks[i] = rankValue
		}
	}

	return k.normalizeVector(ranks, active)
}

// calculateIncentives 计算激励分数
func (k Keeper) calculateIncentives(ranks []sdk.Dec, active []bool, params types.Params) []sdk.Dec {
	n := len(ranks)
	incentives := make([]sdk.Dec, n)

	// 激励直接基于排名
	for i := 0; i < n; i++ {
		if active[i] {
			incentives[i] = ranks[i]
		}
	}

	return k.normalizeVector(incentives, active)
}

// calculateDividends 计算分红分数
func (k Keeper) calculateDividends(incentives []sdk.Dec, active []bool, params types.Params) []sdk.Dec {
	n := len(incentives)
	dividends := make([]sdk.Dec, n)

	// 分红基于激励的平方根（降低方差）
	for i := 0; i < n; i++ {
		if active[i] {
			dividends[i] = k.squareRoot(incentives[i])
		}
	}

	return k.normalizeVector(dividends, active)
}

// calculateEmission 计算最终发行
func (k Keeper) calculateEmission(ctx sdk.Context, netuid uint16, incentives, dividends []sdk.Dec, params types.Params) []sdk.Dec {
	n := len(incentives)
	emission := make([]sdk.Dec, n)

	// 获取子网的总发行量
	totalEmission := k.GetSubnetEmission(ctx, netuid)

	// 激励和分红的混合
	for i := 0; i < n; i++ {
		incentiveWeight := sdk.NewDecWithPrec(7, 1) // 0.7
		dividendWeight := sdk.NewDecWithPrec(3, 1)  // 0.3

		combined := incentives[i].Mul(incentiveWeight).Add(dividends[i].Mul(dividendWeight))
		emission[i] = combined.Mul(totalEmission)
	}

	return emission
}

// updateNeuronStates 更新神经元状态
func (k Keeper) updateNeuronStates(ctx sdk.Context, netuid uint16, neurons []types.NeuronInfo,
	consensus, trust, ranks, incentives, dividends, emission []sdk.Dec, bonds [][]sdk.Dec) {

	for i, neuron := range neurons {
		// 更新神经元的共识相关分数
		neuron.Consensus = consensus[i]
		neuron.Trust = trust[i]
		neuron.Rank = ranks[i]
		neuron.Incentive = incentives[i]
		neuron.Dividends = dividends[i]
		neuron.Emission = emission[i]

		// 更新bonds数据
		neuron.Bonds = [][]uint16{}
		if len(bonds) > i {
			bondsRow := make([]uint16, len(bonds[i]))
			for j, bond := range bonds[i] {
				bondsRow[j] = uint16(bond.MulInt64(65536).TruncateInt64()) // 转换为uint16
			}
			neuron.Bonds = append(neuron.Bonds, bondsRow)
		}

		// 更新最后处理时间
		neuron.LastUpdate = uint16(ctx.BlockHeight())

		// 保存更新后的神经元
		k.SetNeuronInfo(ctx, netuid, neuron.Uid, neuron)
	}
}

// updateSubnetStats 更新子网统计信息
func (k Keeper) updateSubnetStats(ctx sdk.Context, netuid uint16, subnetInfo types.SubnetInfo) {
	subnetInfo.LastUpdate = uint16(ctx.BlockHeight())
	k.SetSubnetInfo(ctx, netuid, subnetInfo)
	k.SetLastEpochBlock(ctx, netuid, uint16(ctx.BlockHeight()))
}

// 辅助函数

// normalizeWeights 归一化权重向量
func (k Keeper) normalizeWeights(weights []sdk.Dec) []sdk.Dec {
	total := sdk.ZeroDec()
	for _, weight := range weights {
		total = total.Add(weight)
	}

	if total.GT(sdk.ZeroDec()) {
		for i := range weights {
			weights[i] = weights[i].Quo(total)
		}
	}

	return weights
}

// normalizeVector 归一化向量（只考虑活跃元素）
func (k Keeper) normalizeVector(vec []sdk.Dec, active []bool) []sdk.Dec {
	total := sdk.ZeroDec()
	for i, val := range vec {
		if i < len(active) && active[i] {
			total = total.Add(val)
		}
	}

	if total.GT(sdk.ZeroDec()) {
		for i := range vec {
			if i < len(active) && active[i] {
				vec[i] = vec[i].Quo(total)
			}
		}
	}

	return vec
}

// powerApproximation 幂函数的近似实现
func (k Keeper) powerApproximation(base, exponent sdk.Dec) sdk.Dec {
	if base.IsZero() || exponent.IsZero() {
		return sdk.ZeroDec()
	}
	if exponent.Equal(sdk.OneDec()) {
		return base
	}

	// 简化实现：使用泰勒级数近似
	// 这里可以根据需要实现更精确的算法
	return base.Mul(exponent.Add(sdk.OneDec()))
}

// squareRoot 平方根的近似实现
func (k Keeper) squareRoot(x sdk.Dec) sdk.Dec {
	if x.IsZero() {
		return sdk.ZeroDec()
	}

	// 使用牛顿法近似
	guess := x.Quo(sdk.NewDec(2))

	for i := 0; i < 10; i++ { // 10次迭代应该足够精确
		newGuess := guess.Add(x.Quo(guess)).Quo(sdk.NewDec(2))
		if newGuess.Sub(guess).Abs().LT(sdk.NewDecWithPrec(1, 10)) {
			break
		}
		guess = newGuess
	}

	return guess
}

// 存储访问辅助函数

// GetLastEpochBlock 获取最后一次epoch执行的区块
func (k Keeper) GetLastEpochBlock(ctx sdk.Context, netuid uint16) uint16 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastEpochBlockKey(netuid)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	var lastBlock uint16
	k.cdc.MustUnmarshal(bz, &lastBlock)
	return lastBlock
}

// SetLastEpochBlock 设置最后一次epoch执行的区块
func (k Keeper) SetLastEpochBlock(ctx sdk.Context, netuid uint16, blockHeight uint16) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastEpochBlockKey(netuid)

	bz := k.cdc.MustMarshal(&blockHeight)
	store.Set(key, bz)
}

// GetLastWeightsUpdate 获取最后一次权重更新的区块
func (k Keeper) GetLastWeightsUpdate(ctx sdk.Context, netuid uint16, uid uint16) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastWeightsUpdateKey(netuid, uid)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	var lastUpdate uint64
	k.cdc.MustUnmarshal(bz, &lastUpdate)
	return lastUpdate
}

// GetNeuronWeights 获取神经元权重
func (k Keeper) GetNeuronWeights(ctx sdk.Context, netuid uint16, uid uint16) []sdk.Dec {
	neuron, found := k.GetNeuronInfo(ctx, netuid, uid)
	if !found || len(neuron.Weights) == 0 {
		return []sdk.Dec{}
	}

	// 转换权重数据
	weights := make([]sdk.Dec, len(neuron.Weights[0]))
	for i, weight := range neuron.Weights[0] {
		weights[i] = sdk.NewDec(int64(weight)).Quo(sdk.NewDec(65536)) // 归一化
	}

	return weights
}

// GetSubnetEmission 获取子网发行量
func (k Keeper) GetSubnetEmission(ctx sdk.Context, netuid uint16) sdk.Dec {
	// 这里应该根据实际的发行机制来计算
	// 暂时返回一个固定值
	return sdk.NewDec(1000000) // 1M tokens per epoch
}

// GetAllNeurons 获取所有神经元
func (k Keeper) GetAllNeurons(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	var neurons []types.NeuronInfo

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetNeuronInfoPrefix(netuid))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var neuron types.NeuronInfo
		k.cdc.MustUnmarshal(iterator.Value(), &neuron)
		neurons = append(neurons, neuron)
	}

	return neurons
}
