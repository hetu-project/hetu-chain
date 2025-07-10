package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/yuma/types"
)

// 验证器相关功能 - 使用MaxAllowedValidators等参数

// SelectValidators 选择验证器 - 使用MaxAllowedValidators参数
func (k Keeper) SelectValidators(ctx sdk.Context, netuid uint16, neurons []types.NeuronInfo) []types.NeuronInfo {
	params := k.GetParams(ctx)

	// 按照trust分数排序选择top验证器
	validators := make([]types.NeuronInfo, 0, params.MaxAllowedValidators)

	// 创建一个按trust分数排序的副本
	sortedNeurons := make([]types.NeuronInfo, len(neurons))
	copy(sortedNeurons, neurons)

	// 简单的冒泡排序（按trust分数降序）
	for i := 0; i < len(sortedNeurons)-1; i++ {
		for j := 0; j < len(sortedNeurons)-i-1; j++ {
			if sortedNeurons[j].Trust.LT(sortedNeurons[j+1].Trust) {
				sortedNeurons[j], sortedNeurons[j+1] = sortedNeurons[j+1], sortedNeurons[j]
			}
		}
	}

	// 选择top验证器
	count := int(params.MaxAllowedValidators)
	if count > len(sortedNeurons) {
		count = len(sortedNeurons)
	}

	for i := 0; i < count; i++ {
		validators = append(validators, sortedNeurons[i])
	}

	return validators
}

// ValidateValidatorSequence 验证验证器序列 - 使用ValidatorSequenceLength参数
func (k Keeper) ValidateValidatorSequence(ctx sdk.Context, netuid uint16, validators []types.NeuronInfo) error {
	params := k.GetParams(ctx)

	// 检查验证器序列长度
	if len(validators) > int(params.ValidatorSequenceLength) {
		return fmt.Errorf("验证器序列长度 %d 超过最大限制 %d", len(validators), params.ValidatorSequenceLength)
	}

	// 验证每个验证器的有效性
	for _, validator := range validators {
		if !k.isValidValidator(ctx, netuid, validator, params) {
			return fmt.Errorf("验证器 %d 无效", validator.Uid)
		}
	}

	return nil
}

// isValidValidator 检查是否为有效验证器
func (k Keeper) isValidValidator(ctx sdk.Context, netuid uint16, validator types.NeuronInfo, params types.Params) bool {
	// 检查验证器是否活跃
	if !validator.Active {
		return false
	}

	// 检查验证器的trust分数
	if validator.Trust.IsZero() {
		return false
	}

	// 检查验证器的最后更新时间
	currentBlock := uint16(ctx.BlockHeight())
	if (currentBlock - validator.LastUpdate) > params.ActivityCutoff {
		return false
	}

	return true
}

// RunValidatorEpoch 运行验证器epoch - 使用ValidatorEpochLength参数
func (k Keeper) RunValidatorEpoch(ctx sdk.Context, netuid uint16) error {
	params := k.GetParams(ctx)

	// 检查是否应该运行验证器epoch
	if !k.shouldRunValidatorEpoch(ctx, netuid, params.ValidatorEpochLength) {
		return nil
	}

	// 获取所有神经元
	neurons := k.GetAllNeurons(ctx, netuid)

	// 选择验证器
	validators := k.SelectValidators(ctx, netuid, neurons)

	// 验证验证器序列
	if err := k.ValidateValidatorSequence(ctx, netuid, validators); err != nil {
		return err
	}

	// 执行验证器相关的共识计算
	if err := k.processValidatorConsensus(ctx, netuid, validators, params); err != nil {
		return err
	}

	// 更新验证器epoch时间
	k.SetLastValidatorEpoch(ctx, netuid, uint16(ctx.BlockHeight()))

	return nil
}

// shouldRunValidatorEpoch 检查是否应该运行验证器epoch
func (k Keeper) shouldRunValidatorEpoch(ctx sdk.Context, netuid uint16, validatorEpochLength uint16) bool {
	lastValidatorEpoch := k.GetLastValidatorEpoch(ctx, netuid)
	currentBlock := uint16(ctx.BlockHeight())

	return (currentBlock - lastValidatorEpoch) >= validatorEpochLength
}

// processValidatorConsensus 处理验证器共识
func (k Keeper) processValidatorConsensus(ctx sdk.Context, netuid uint16, validators []types.NeuronInfo, params types.Params) error {
	// 计算验证器logits分歧 - 使用ValidatorLogitsDivergence参数
	for i, validator := range validators {
		logitsDivergence := k.calculateLogitsDivergence(ctx, netuid, validator, params)

		// 如果分歧过大，标记为无效
		if logitsDivergence.GT(params.ValidatorLogitsDivergence) {
			validators[i].Active = false
			k.SetNeuronInfo(ctx, netuid, validator.Uid, validators[i])
		}
	}

	// 检查是否需要重置验证器 - 使用ValidatorEpochsPerReset参数
	if k.shouldResetValidators(ctx, netuid, params.ValidatorEpochsPerReset) {
		k.resetValidators(ctx, netuid, validators)
	}

	return nil
}

// calculateLogitsDivergence 计算logits分歧
func (k Keeper) calculateLogitsDivergence(ctx sdk.Context, netuid uint16, validator types.NeuronInfo, params types.Params) sdk.Dec {
	// 获取验证器的历史logits
	historicalLogits := k.getHistoricalLogits(ctx, netuid, validator.Uid)

	// 计算当前logits
	currentLogits := k.getCurrentLogits(ctx, netuid, validator.Uid)

	// 计算KL散度或其他分歧度量
	divergence := k.calculateKLDivergence(historicalLogits, currentLogits)

	return divergence
}

// shouldResetValidators 检查是否应该重置验证器
func (k Keeper) shouldResetValidators(ctx sdk.Context, netuid uint16, validatorEpochsPerReset uint16) bool {
	validatorEpochCount := k.GetValidatorEpochCount(ctx, netuid)
	return validatorEpochCount >= validatorEpochsPerReset
}

// resetValidators 重置验证器
func (k Keeper) resetValidators(ctx sdk.Context, netuid uint16, validators []types.NeuronInfo) {
	// 重置验证器状态
	for _, validator := range validators {
		validator.Trust = sdk.ZeroDec()
		validator.Consensus = sdk.ZeroDec()
		validator.Rank = sdk.ZeroDec()
		k.SetNeuronInfo(ctx, netuid, validator.Uid, validator)
	}

	// 重置epoch计数
	k.SetValidatorEpochCount(ctx, netuid, 0)
}

// 辅助存储函数

// GetLastValidatorEpoch 获取最后验证器epoch
func (k Keeper) GetLastValidatorEpoch(ctx sdk.Context, netuid uint16) uint16 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastValidatorEpochKey(netuid)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	var lastEpoch uint16
	k.cdc.MustUnmarshal(bz, &lastEpoch)
	return lastEpoch
}

// SetLastValidatorEpoch 设置最后验证器epoch
func (k Keeper) SetLastValidatorEpoch(ctx sdk.Context, netuid uint16, blockHeight uint16) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastValidatorEpochKey(netuid)

	bz := k.cdc.MustMarshal(&blockHeight)
	store.Set(key, bz)
}

// GetValidatorEpochCount 获取验证器epoch计数
func (k Keeper) GetValidatorEpochCount(ctx sdk.Context, netuid uint16) uint16 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorEpochCountKey(netuid)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	var count uint16
	k.cdc.MustUnmarshal(bz, &count)
	return count
}

// SetValidatorEpochCount 设置验证器epoch计数
func (k Keeper) SetValidatorEpochCount(ctx sdk.Context, netuid uint16, count uint16) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorEpochCountKey(netuid)

	bz := k.cdc.MustMarshal(&count)
	store.Set(key, bz)
}

// getHistoricalLogits 获取历史logits
func (k Keeper) getHistoricalLogits(ctx sdk.Context, netuid uint16, uid uint16) []sdk.Dec {
	// 实现获取历史logits的逻辑
	// 暂时返回空数组
	return []sdk.Dec{}
}

// getCurrentLogits 获取当前logits
func (k Keeper) getCurrentLogits(ctx sdk.Context, netuid uint16, uid uint16) []sdk.Dec {
	// 实现获取当前logits的逻辑
	// 暂时返回空数组
	return []sdk.Dec{}
}

// calculateKLDivergence 计算KL散度
func (k Keeper) calculateKLDivergence(p, q []sdk.Dec) sdk.Dec {
	if len(p) != len(q) || len(p) == 0 {
		return sdk.ZeroDec()
	}

	divergence := sdk.ZeroDec()
	for i := 0; i < len(p); i++ {
		if p[i].GT(sdk.ZeroDec()) && q[i].GT(sdk.ZeroDec()) {
			// KL(P||Q) = sum(P * log(P/Q))
			ratio := p[i].Quo(q[i])
			logRatio := k.logarithm(ratio)
			divergence = divergence.Add(p[i].Mul(logRatio))
		}
	}

	return divergence
}

// logarithm 对数函数的近似实现
func (k Keeper) logarithm(x sdk.Dec) sdk.Dec {
	if x.LTE(sdk.ZeroDec()) {
		return sdk.ZeroDec()
	}

	// 使用泰勒级数近似 ln(x)
	// 这里可以实现更精确的算法
	return x.Sub(sdk.OneDec())
}
