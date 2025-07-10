package types

import (
	"strconv"
)

// EpochResult epoch 计算结果
type EpochResult struct {
	Netuid    uint16      `json:"netuid"`
	Accounts  []string    `json:"accounts"`
	Emission  []uint64    `json:"emission"`
	Dividend  []uint64    `json:"dividend"`
	Bonds     [][]float64 `json:"bonds"`
	Consensus []float64   `json:"consensus"`
}

// EpochParams epoch 参数（从 event 模块的 Subnet.Params 解析）
type EpochParams struct {
	// 核心参数
	Kappa float64 `json:"kappa"` // 多数阈值 (0.5)
	Alpha float64 `json:"alpha"` // EMA 参数 (0.1-0.9)
	Delta float64 `json:"delta"` // 权重裁剪范围 (1.0)

	// 活跃性参数
	ActivityCutoff uint64 `json:"activity_cutoff"` // 活跃截止时间
	ImmunityPeriod uint64 `json:"immunity_period"` // 免疫期

	// 权重参数
	MaxWeightsLimit     uint64 `json:"max_weights_limit"`      // 最大权重数量
	MinAllowedWeights   uint64 `json:"min_allowed_weights"`    // 最小权重数量
	WeightsSetRateLimit uint64 `json:"weights_set_rate_limit"` // 权重设置速率限制

	// 其他参数
	Tempo              uint64  `json:"tempo"`                // epoch 运行频率
	BondsPenalty       float64 `json:"bonds_penalty"`        // bonds 惩罚
	BondsMovingAverage float64 `json:"bonds_moving_average"` // bonds 移动平均
}

// DefaultEpochParams 默认参数
func DefaultEpochParams() EpochParams {
	return EpochParams{
		Kappa:               0.5,
		Alpha:               0.1,
		Delta:               1.0,
		ActivityCutoff:      5000,
		ImmunityPeriod:      4096,
		MaxWeightsLimit:     1000,
		MinAllowedWeights:   8,
		WeightsSetRateLimit: 100,
		Tempo:               100,
		BondsPenalty:        0.1,
		BondsMovingAverage:  0.9,
	}
}

// ParseEpochParams 从 event 模块的 Subnet.Params 解析参数
func ParseEpochParams(params map[string]string) EpochParams {
	epochParams := DefaultEpochParams()

	// 解析参数（如果存在的话）
	if val, ok := params["kappa"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			epochParams.Kappa = f
		}
	}

	if val, ok := params["alpha"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			epochParams.Alpha = f
		}
	}

	if val, ok := params["delta"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			epochParams.Delta = f
		}
	}

	if val, ok := params["activity_cutoff"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.ActivityCutoff = f
		}
	}

	if val, ok := params["immunity_period"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.ImmunityPeriod = f
		}
	}

	if val, ok := params["max_weights_limit"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.MaxWeightsLimit = f
		}
	}

	if val, ok := params["min_allowed_weights"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.MinAllowedWeights = f
		}
	}

	if val, ok := params["weights_set_rate_limit"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.WeightsSetRateLimit = f
		}
	}

	if val, ok := params["tempo"]; ok {
		if f, err := strconv.ParseUint(val, 10, 64); err == nil {
			epochParams.Tempo = f
		}
	}

	if val, ok := params["bonds_penalty"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			epochParams.BondsPenalty = f
		}
	}

	if val, ok := params["bonds_moving_average"]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			epochParams.BondsMovingAverage = f
		}
	}

	return epochParams
}

// ValidatorInfo 验证者信息（从 event 模块获取）
type ValidatorInfo struct {
	Address string   `json:"address"`
	Stake   float64  `json:"stake"`
	Weights []uint64 `json:"weights"`
	Active  bool     `json:"active"`
}

// SubnetEpochData 子网 epoch 数据
type SubnetEpochData struct {
	Netuid     uint16          `json:"netuid"`
	Validators []ValidatorInfo `json:"validators"`
	Params     EpochParams     `json:"params"`
	Emission   uint64          `json:"emission"`
}
