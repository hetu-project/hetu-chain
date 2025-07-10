package types

import "fmt"

// GenesisState 定义yuma模块的创世状态
type GenesisState struct {
	Params  Params       `json:"params" yaml:"params"`
	Subnets []SubnetInfo `json:"subnets" yaml:"subnets"`
	Neurons []NeuronInfo `json:"neurons" yaml:"neurons"`
	Weights []WeightData `json:"weights" yaml:"weights"`
	Bonds   []BondData   `json:"bonds" yaml:"bonds"`
}

// WeightData 权重数据
type WeightData struct {
	Netuid  uint16   `json:"netuid"`
	Uid     uint16   `json:"uid"`
	Weights []uint16 `json:"weights"`
}

// BondData 绑定数据
type BondData struct {
	Netuid uint16   `json:"netuid"`
	Uid    uint16   `json:"uid"`
	Bonds  []uint16 `json:"bonds"`
}

// DefaultGenesis 返回默认的创世状态
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:  DefaultParams(),
		Subnets: []SubnetInfo{},
		Neurons: []NeuronInfo{},
		Weights: []WeightData{},
		Bonds:   []BondData{},
	}
}

// ValidateGenesis 验证创世状态
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.ValidateBasic(); err != nil {
		return fmt.Errorf("无效的参数: %w", err)
	}

	// 验证子网
	seenNetuids := make(map[uint16]bool)
	for _, subnet := range data.Subnets {
		if err := subnet.ValidateBasic(); err != nil {
			return fmt.Errorf("无效的子网信息: %w", err)
		}
		if seenNetuids[subnet.Netuid] {
			return fmt.Errorf("重复的网络ID: %d", subnet.Netuid)
		}
		seenNetuids[subnet.Netuid] = true
	}

	// 验证神经元
	seenNeurons := make(map[string]bool)
	for _, neuron := range data.Neurons {
		if err := neuron.ValidateBasic(); err != nil {
			return fmt.Errorf("无效的神经元信息: %w", err)
		}
		key := fmt.Sprintf("%d-%s", neuron.Netuid, neuron.Hotkey)
		if seenNeurons[key] {
			return fmt.Errorf("重复的神经元: netuid=%d, hotkey=%s", neuron.Netuid, neuron.Hotkey)
		}
		seenNeurons[key] = true

		// 检查神经元是否属于已定义的子网
		if !seenNetuids[neuron.Netuid] {
			return fmt.Errorf("神经元引用了不存在的子网: %d", neuron.Netuid)
		}
	}

	return nil
}
