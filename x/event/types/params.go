package types

// DefaultParamsMap 返回默认参数映射
// 当合约事件中的 param 字段没有指定某个参数时，使用这些默认值
func DefaultParamsMap() map[string]string {
	return map[string]string{
		"rho":                         "0.5",   // 默认 rho 值
		"kappa":                       "32767", // 默认 kappa 值
		"max_allowed_uids":            "4096",  // 默认最大允许的子网数量
		"immunity_period":             "4096",  // 默认免疫周期
		"activity_cutoff":             "5000",  // 默认活动截止
		"max_weights_limit":           "1000",  // 默认最大权重限制
		"weights_version_key":         "0",     // 默认权重版本
		"min_allowed_weights":         "8",     // 默认最小允许权重
		"max_allowed_validators":      "128",   // 默认最大允许验证者数量
		"tempo":                       "100",   // 默认 tempo 值
		"adjustment_interval":         "112",   // 默认调整间隔
		"adjustment_alpha":            "58982", // 默认调整 alpha 值
		"bonds_moving_average":        "0.9",   // 默认移动平均
		"weights_set_rate_limit":      "1000",  // 默认权重设置速率限制
		"validator_prune_len":         "100",   // 默认验证者修剪长度
		"validator_logits_divergence": "0.1",   // 默认验证者 logits 分歧
		"validator_sequence_length":   "100",   // 默认验证者序列长度
		"validator_epoch_length":      "100",   // 默认验证者 epoch 长度
		"validator_epochs_per_reset":  "100",   // 默认验证者 epoch 重置间隔
		"liquid_alpha_enabled":        "true",  // 默认 liquid alpha 启用
		"alpha_enabled":               "true",  // 默认 alpha 启用
		"alpha_high":                  "0.9",   // 默认 alpha 上限
		"alpha_low":                   "0.1",   // 默认 alpha 下限
		"bonds_penalty":               "0.1",   // 默认质押惩罚
	}
}
