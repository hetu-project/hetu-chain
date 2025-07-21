package types

import (
	"cosmossdk.io/math"
)

// SubnetEmissionData represents emission data for a subnet
type SubnetEmissionData struct {
	Netuid                 uint16   `json:"netuid" yaml:"netuid"`
	TaoIn                  math.Int `json:"tao_in" yaml:"tao_in"`
	AlphaIn                math.Int `json:"alpha_in" yaml:"alpha_in"`
	AlphaOut               math.Int `json:"alpha_out" yaml:"alpha_out"`
	OwnerCut               math.Int `json:"owner_cut" yaml:"owner_cut"`
	RootDivs               math.Int `json:"root_divs" yaml:"root_divs"`
	SubnetAlphaInEmission  math.Int `json:"subnet_alpha_in_emission" yaml:"subnet_alpha_in_emission"`   // 累计Alpha流入排放
	SubnetAlphaOutEmission math.Int `json:"subnet_alpha_out_emission" yaml:"subnet_alpha_out_emission"` // 累计Alpha流出排放
	SubnetTaoInEmission    math.Int `json:"subnet_tao_in_emission" yaml:"subnet_tao_in_emission"`       // 累计TAO流入排放
}
