package types

import (
	"encoding/json"

	"cosmossdk.io/math"
)

// Subnet represents a subnet in the network
type Subnet struct {
	Netuid                uint16            `json:"netuid" yaml:"netuid"`
	Owner                 string            `json:"owner" yaml:"owner"`
	LockAmount            string            `json:"lock_amount" yaml:"lock_amount"`
	BurnedTao             string            `json:"burned_tao" yaml:"burned_tao"`
	Pool                  string            `json:"pool" yaml:"pool"`
	Params                map[string]string `json:"params" yaml:"params"`
	FirstEmissionBlock    uint64            `json:"first_emission_block" yaml:"first_emission_block"`         // 首次排放区块号
	Mechanism             uint8             `json:"mechanism" yaml:"mechanism"`                               // 子网机制 (0=稳定, 1=动态)
	EMAPriceHalvingBlocks uint64            `json:"ema_price_halving_blocks" yaml:"ema_price_halving_blocks"` // EMA价格减半区块数 (默认201600=4周)
}

// MarshalJSON implements json.Marshaler
func (s Subnet) MarshalJSON() ([]byte, error) {
	type Alias Subnet
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&s),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Subnet) UnmarshalJSON(data []byte) error {
	type Alias Subnet
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

// SubnetPriceData represents price-related data for a subnet
type SubnetPriceData struct {
	Netuid         uint16         `json:"netuid" yaml:"netuid"`
	MovingPrice    math.LegacyDec `json:"moving_price" yaml:"moving_price"`         // 移动价格
	AlphaPrice     math.LegacyDec `json:"alpha_price" yaml:"alpha_price"`           // Alpha价格
	SubnetTAO      math.Int       `json:"subnet_tao" yaml:"subnet_tao"`             // 子网中的TAO数量
	SubnetAlphaIn  math.Int       `json:"subnet_alpha_in" yaml:"subnet_alpha_in"`   // 池中的Alpha数量
	SubnetAlphaOut math.Int       `json:"subnet_alpha_out" yaml:"subnet_alpha_out"` // 子网中的Alpha数量
	Volume         math.Int       `json:"volume" yaml:"volume"`                     // 总交易量
}

// SubnetEmissionData represents emission data for a subnet
type SubnetEmissionData struct {
	Netuid           uint16   `json:"netuid" yaml:"netuid"`
	TaoInEmission    math.Int `json:"tao_in_emission" yaml:"tao_in_emission"`       // TAO输入排放
	AlphaInEmission  math.Int `json:"alpha_in_emission" yaml:"alpha_in_emission"`   // Alpha输入排放
	AlphaOutEmission math.Int `json:"alpha_out_emission" yaml:"alpha_out_emission"` // Alpha输出排放
}
