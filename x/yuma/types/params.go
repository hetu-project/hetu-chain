package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// 参数键
var (
	KeyRho                       = []byte("Rho")
	KeyKappa                     = []byte("Kappa")
	KeyMaxAllowedUids            = []byte("MaxAllowedUids")
	KeyImmunityPeriod            = []byte("ImmunityPeriod")
	KeyActivityCutoff            = []byte("ActivityCutoff")
	KeyMaxWeightsLimit           = []byte("MaxWeightsLimit")
	KeyWeightsVersionKey         = []byte("WeightsVersionKey")
	KeyMinAllowedWeights         = []byte("MinAllowedWeights")
	KeyMaxAllowedValidators      = []byte("MaxAllowedValidators")
	KeyTempo                     = []byte("Tempo")
	KeyAdjustmentInterval        = []byte("AdjustmentInterval")
	KeyAdjustmentAlpha           = []byte("AdjustmentAlpha")
	KeyBondsMovingAverage        = []byte("BondsMovingAverage")
	KeyWeightsSetRateLimit       = []byte("WeightsSetRateLimit")
	KeyValidatorPruneLen         = []byte("ValidatorPruneLen")
	KeyValidatorLogitsDivergence = []byte("ValidatorLogitsDivergence")
	KeyValidatorSequenceLength   = []byte("ValidatorSequenceLength")
	KeyValidatorEpochLength      = []byte("ValidatorEpochLength")
	KeyValidatorEpochsPerReset   = []byte("ValidatorEpochsPerReset")
	KeyLiquidAlphaEnabled        = []byte("LiquidAlphaEnabled")
	KeyAlphaEnabled              = []byte("AlphaEnabled")
	KeyAlphaHigh                 = []byte("AlphaHigh")
	KeyAlphaLow                  = []byte("AlphaLow")
	KeyBondsPenalty              = []byte("BondsPenalty")
)

// 默认参数值
const (
	DefaultRho                       = "10"
	DefaultKappa                     = "32767"
	DefaultMaxAllowedUids            = uint16(4096)
	DefaultImmunityPeriod            = uint16(4096)
	DefaultActivityCutoff            = uint16(5000)
	DefaultMaxWeightsLimit           = uint16(1000)
	DefaultWeightsVersionKey         = uint64(0)
	DefaultMinAllowedWeights         = uint16(8)
	DefaultMaxAllowedValidators      = uint16(128)
	DefaultTempo                     = uint16(100) // 每100个区块运行一次共识
	DefaultAdjustmentInterval        = uint16(112) // 每112个区块调整参数
	DefaultAdjustmentAlpha           = "58982"
	DefaultBondsMovingAverage        = uint64(900000)
	DefaultWeightsSetRateLimit       = uint64(100)
	DefaultValidatorPruneLen         = uint64(1)
	DefaultValidatorLogitsDivergence = "1073741824"
	DefaultValidatorSequenceLength   = uint16(256)
	DefaultValidatorEpochLength      = uint16(100)
	DefaultValidatorEpochsPerReset   = uint16(60)
	DefaultLiquidAlphaEnabled        = true
	DefaultAlphaEnabled              = false
	DefaultAlphaHigh                 = "53000" // 高alpha阈值
	DefaultAlphaLow                  = "46000" // 低alpha阈值
	DefaultBondsPenalty              = "1000"  // bonds惩罚系数
)

// Params 定义模块参数
type Params struct {
	Rho                       sdk.Dec `json:"rho" yaml:"rho"`
	Kappa                     sdk.Dec `json:"kappa" yaml:"kappa"`
	MaxAllowedUids            uint16  `json:"max_allowed_uids" yaml:"max_allowed_uids"`
	ImmunityPeriod            uint16  `json:"immunity_period" yaml:"immunity_period"`
	ActivityCutoff            uint16  `json:"activity_cutoff" yaml:"activity_cutoff"`
	MaxWeightsLimit           uint16  `json:"max_weights_limit" yaml:"max_weights_limit"`
	WeightsVersionKey         uint64  `json:"weights_version_key" yaml:"weights_version_key"`
	MinAllowedWeights         uint16  `json:"min_allowed_weights" yaml:"min_allowed_weights"`
	MaxAllowedValidators      uint16  `json:"max_allowed_validators" yaml:"max_allowed_validators"`
	Tempo                     uint16  `json:"tempo" yaml:"tempo"`
	AdjustmentInterval        uint16  `json:"adjustment_interval" yaml:"adjustment_interval"`
	AdjustmentAlpha           sdk.Dec `json:"adjustment_alpha" yaml:"adjustment_alpha"`
	BondsMovingAverage        uint64  `json:"bonds_moving_average" yaml:"bonds_moving_average"`
	WeightsSetRateLimit       uint64  `json:"weights_set_rate_limit" yaml:"weights_set_rate_limit"`
	ValidatorPruneLen         uint64  `json:"validator_prune_len" yaml:"validator_prune_len"`
	ValidatorLogitsDivergence sdk.Dec `json:"validator_logits_divergence" yaml:"validator_logits_divergence"`
	ValidatorSequenceLength   uint16  `json:"validator_sequence_length" yaml:"validator_sequence_length"`
	ValidatorEpochLength      uint16  `json:"validator_epoch_length" yaml:"validator_epoch_length"`
	ValidatorEpochsPerReset   uint16  `json:"validator_epochs_per_reset" yaml:"validator_epochs_per_reset"`
	LiquidAlphaEnabled        bool    `json:"liquid_alpha_enabled" yaml:"liquid_alpha_enabled"`
	AlphaEnabled              bool    `json:"alpha_enabled" yaml:"alpha_enabled"`
	AlphaHigh                 sdk.Dec `json:"alpha_high" yaml:"alpha_high"`
	AlphaLow                  sdk.Dec `json:"alpha_low" yaml:"alpha_low"`
	BondsPenalty              sdk.Dec `json:"bonds_penalty" yaml:"bonds_penalty"`
}

// NewParams 创建新的参数
func NewParams() Params {
	rho, _ := sdk.NewDecFromStr(DefaultRho)
	kappa, _ := sdk.NewDecFromStr(DefaultKappa)
	adjustmentAlpha, _ := sdk.NewDecFromStr(DefaultAdjustmentAlpha)
	validatorLogitsDivergence, _ := sdk.NewDecFromStr(DefaultValidatorLogitsDivergence)
	alphaHigh, _ := sdk.NewDecFromStr(DefaultAlphaHigh)
	alphaLow, _ := sdk.NewDecFromStr(DefaultAlphaLow)
	bondsPenalty, _ := sdk.NewDecFromStr(DefaultBondsPenalty)

	return Params{
		Rho:                       rho,
		Kappa:                     kappa,
		MaxAllowedUids:            DefaultMaxAllowedUids,
		ImmunityPeriod:            DefaultImmunityPeriod,
		ActivityCutoff:            DefaultActivityCutoff,
		MaxWeightsLimit:           DefaultMaxWeightsLimit,
		WeightsVersionKey:         DefaultWeightsVersionKey,
		MinAllowedWeights:         DefaultMinAllowedWeights,
		MaxAllowedValidators:      DefaultMaxAllowedValidators,
		Tempo:                     DefaultTempo,
		AdjustmentInterval:        DefaultAdjustmentInterval,
		AdjustmentAlpha:           adjustmentAlpha,
		BondsMovingAverage:        DefaultBondsMovingAverage,
		WeightsSetRateLimit:       DefaultWeightsSetRateLimit,
		ValidatorPruneLen:         DefaultValidatorPruneLen,
		ValidatorLogitsDivergence: validatorLogitsDivergence,
		ValidatorSequenceLength:   DefaultValidatorSequenceLength,
		ValidatorEpochLength:      DefaultValidatorEpochLength,
		ValidatorEpochsPerReset:   DefaultValidatorEpochsPerReset,
		LiquidAlphaEnabled:        DefaultLiquidAlphaEnabled,
		AlphaEnabled:              DefaultAlphaEnabled,
		AlphaHigh:                 alphaHigh,
		AlphaLow:                  alphaLow,
		BondsPenalty:              bondsPenalty,
	}
}

// DefaultParams 返回默认参数
func DefaultParams() Params {
	return NewParams()
}

// ParamSetPairs 获取参数设置对
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyRho, &p.Rho, validateRho),
		paramtypes.NewParamSetPair(KeyKappa, &p.Kappa, validateKappa),
		paramtypes.NewParamSetPair(KeyMaxAllowedUids, &p.MaxAllowedUids, validateMaxAllowedUids),
		paramtypes.NewParamSetPair(KeyImmunityPeriod, &p.ImmunityPeriod, validateImmunityPeriod),
		paramtypes.NewParamSetPair(KeyActivityCutoff, &p.ActivityCutoff, validateActivityCutoff),
		paramtypes.NewParamSetPair(KeyMaxWeightsLimit, &p.MaxWeightsLimit, validateMaxWeightsLimit),
		paramtypes.NewParamSetPair(KeyMinAllowedWeights, &p.MinAllowedWeights, validateMinAllowedWeights),
		paramtypes.NewParamSetPair(KeyMaxAllowedValidators, &p.MaxAllowedValidators, validateMaxAllowedValidators),
		paramtypes.NewParamSetPair(KeyTempo, &p.Tempo, validateTempo),
		paramtypes.NewParamSetPair(KeyAdjustmentInterval, &p.AdjustmentInterval, validateAdjustmentInterval),
		paramtypes.NewParamSetPair(KeyAdjustmentAlpha, &p.AdjustmentAlpha, validateAdjustmentAlpha),
		paramtypes.NewParamSetPair(KeyBondsMovingAverage, &p.BondsMovingAverage, validateBondsMovingAverage),
		paramtypes.NewParamSetPair(KeyWeightsSetRateLimit, &p.WeightsSetRateLimit, validateWeightsSetRateLimit),
		paramtypes.NewParamSetPair(KeyValidatorPruneLen, &p.ValidatorPruneLen, validateValidatorPruneLen),
		paramtypes.NewParamSetPair(KeyValidatorLogitsDivergence, &p.ValidatorLogitsDivergence, validateValidatorLogitsDivergence),
		paramtypes.NewParamSetPair(KeyValidatorSequenceLength, &p.ValidatorSequenceLength, validateValidatorSequenceLength),
		paramtypes.NewParamSetPair(KeyValidatorEpochLength, &p.ValidatorEpochLength, validateValidatorEpochLength),
		paramtypes.NewParamSetPair(KeyValidatorEpochsPerReset, &p.ValidatorEpochsPerReset, validateValidatorEpochsPerReset),
		paramtypes.NewParamSetPair(KeyLiquidAlphaEnabled, &p.LiquidAlphaEnabled, validateLiquidAlphaEnabled),
		paramtypes.NewParamSetPair(KeyAlphaEnabled, &p.AlphaEnabled, validateAlphaEnabled),
		paramtypes.NewParamSetPair(KeyAlphaHigh, &p.AlphaHigh, validateAlphaHigh),
		paramtypes.NewParamSetPair(KeyAlphaLow, &p.AlphaLow, validateAlphaLow),
		paramtypes.NewParamSetPair(KeyBondsPenalty, &p.BondsPenalty, validateBondsPenalty),
	}
}

// ValidateBasic 基本验证
func (p Params) ValidateBasic() error {
	if p.Rho.IsNegative() {
		return fmt.Errorf("rho 必须为非负数: %s", p.Rho)
	}
	if p.Kappa.IsNegative() {
		return fmt.Errorf("kappa 必须为非负数: %s", p.Kappa)
	}
	if p.MaxAllowedUids == 0 {
		return fmt.Errorf("max_allowed_uids 必须大于0")
	}
	return nil
}

// 验证函数
func validateRho(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("rho 必须为非负数: %s", v)
	}
	return nil
}

func validateKappa(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("kappa 必须为非负数: %s", v)
	}
	return nil
}

func validateMaxAllowedUids(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max_allowed_uids 必须大于0")
	}
	return nil
}

// 添加其他参数验证函数
func validateImmunityPeriod(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("immunity_period 必须大于0")
	}
	return nil
}

func validateActivityCutoff(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("activity_cutoff 必须大于0")
	}
	return nil
}

func validateMaxWeightsLimit(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max_weights_limit 必须大于0")
	}
	return nil
}

func validateMinAllowedWeights(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	return nil
}

func validateMaxAllowedValidators(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("max_allowed_validators 必须大于0")
	}
	return nil
}

func validateTempo(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("tempo 必须大于0")
	}
	return nil
}

func validateAdjustmentInterval(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("adjustment_interval 必须大于0")
	}
	return nil
}

func validateAdjustmentAlpha(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("adjustment_alpha 必须为非负数: %s", v)
	}
	return nil
}

func validateBondsMovingAverage(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("bonds_moving_average 必须大于0")
	}
	return nil
}

func validateWeightsSetRateLimit(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("weights_set_rate_limit 必须大于0")
	}
	return nil
}

func validateValidatorPruneLen(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	return nil
}

func validateValidatorLogitsDivergence(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("validator_logits_divergence 必须为非负数: %s", v)
	}
	return nil
}

func validateValidatorSequenceLength(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("validator_sequence_length 必须大于0")
	}
	return nil
}

func validateValidatorEpochLength(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("validator_epoch_length 必须大于0")
	}
	return nil
}

func validateValidatorEpochsPerReset(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("validator_epochs_per_reset 必须大于0")
	}
	return nil
}

func validateLiquidAlphaEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	return nil
}

func validateAlphaEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	return nil
}

func validateAlphaHigh(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("alpha_high 必须为非负数: %s", v)
	}
	return nil
}

func validateAlphaLow(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("alpha_low 必须为非负数: %s", v)
	}
	return nil
}

func validateBondsPenalty(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("无效的参数类型: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("bonds_penalty 必须为非负数: %s", v)
	}
	return nil
}
