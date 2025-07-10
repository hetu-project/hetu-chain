package types

// DefaultParamsMap 返回默认参数映射
// 当合约事件中的 param 字段没有指定某个参数时，使用这些默认值
func DefaultParamsMap() map[string]string {
	return map[string]string{
		"epoch":                   "100",    // 默认 epoch 长度
		"kappa":                   "0.5",    // 默认 kappa 值
		"bonds_penalty":           "0.1",    // 默认质押惩罚
		"bonds_moving_average":    "0.9",    // 默认移动平均
		"alpha_low":               "0.1",    // 默认 alpha 下限
		"alpha_high":              "0.9",    // 默认 alpha 上限
		"alpha_sigmoid_steepness": "1.0",    // 默认 sigmoid 陡峭度
		"min_stake":               "1000",   // 默认最小质押量
		"max_stake":               "100000", // 默认最大质押量
	}
}
