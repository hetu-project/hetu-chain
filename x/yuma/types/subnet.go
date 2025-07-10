package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SubnetInfo 子网信息结构体
type SubnetInfo struct {
	Netuid                uint16  `json:"netuid"`
	N                     uint16  `json:"n"`                       // 神经元数量
	Modality              uint16  `json:"modality"`                // 模式
	Tempo                 uint16  `json:"tempo"`                   // 节拍
	LastMechanism         uint64  `json:"last_mechanism"`          // 上次机制步骤
	LastEpochLength       uint16  `json:"last_epoch_length"`       // 上次时期长度
	LastEpoch             uint64  `json:"last_epoch"`              // 上次时期
	NetworkLastRegistered uint64  `json:"network_last_registered"` // 网络上次注册
	EmissionValues        sdk.Dec `json:"emission_values"`         // 发行值
	Burn                  sdk.Int `json:"burn"`                    // 燃烧值
	Owner                 string  `json:"owner"`                   // 拥有者
	DaoTreasuryAddress    string  `json:"dao_treasury_address"`    // DAO金库地址
}

// NeuronInfo 神经元信息结构体
type NeuronInfo struct {
	Hotkey              string         `json:"hotkey"`                // 热键
	Coldkey             string         `json:"coldkey"`               // 冷键
	Uid                 uint16         `json:"uid"`                   // 唯一标识符
	Netuid              uint16         `json:"netuid"`                // 网络唯一标识符
	Active              bool           `json:"active"`                // 是否活跃
	Axon                AxonInfo       `json:"axon"`                  // 轴突信息
	Prometheus          PrometheusInfo `json:"prometheus"`            // 普罗米修斯信息
	Stake               sdk.Int        `json:"stake"`                 // 质押
	Rank                sdk.Dec        `json:"rank"`                  // 排名
	Emission            sdk.Dec        `json:"emission"`              // 发行
	Incentive           sdk.Dec        `json:"incentive"`             // 激励
	Consensus           sdk.Dec        `json:"consensus"`             // 共识
	Trust               sdk.Dec        `json:"trust"`                 // 信任
	ValidatorTrust      sdk.Dec        `json:"validator_trust"`       // 验证者信任
	Dividends           sdk.Dec        `json:"dividends"`             // 分红
	LastUpdate          uint16         `json:"last_update"`           // 上次更新
	BlockAtRegistration uint16         `json:"block_at_registration"` // 注册时的区块高度
	ValidatorPermit     bool           `json:"validator_permit"`      // 验证者许可
	Weights             [][]uint16     `json:"weights"`               // 权重
	Bonds               [][]uint16     `json:"bonds"`                 // 绑定
	PruningScore        sdk.Dec        `json:"pruning_score"`         // 修剪分数
}

// AxonInfo 轴突信息
type AxonInfo struct {
	Block             uint64 `json:"block"`               // 区块
	Version           uint32 `json:"version"`             // 版本
	Ip                uint32 `json:"ip"`                  // IP地址
	Port              uint16 `json:"port"`                // 端口
	IpType            uint8  `json:"ip_type"`             // IP类型
	Protocol          uint8  `json:"protocol"`            // 协议
	PlaceholderIPAddr string `json:"placeholder_ip_addr"` // 占位符IP地址
}

// PrometheusInfo 普罗米修斯信息
type PrometheusInfo struct {
	Block   uint64 `json:"block"`   // 区块
	Version uint32 `json:"version"` // 版本
	Ip      uint32 `json:"ip"`      // IP地址
	Port    uint16 `json:"port"`    // 端口
	IpType  uint8  `json:"ip_type"` // IP类型
}

// ValidateBasic 基本验证
func (s *SubnetInfo) ValidateBasic() error {
	if s.Netuid == 0 {
		return fmt.Errorf("无效的网络ID: %d", s.Netuid)
	}
	if s.Owner == "" {
		return fmt.Errorf("拥有者地址不能为空")
	}
	return nil
}

// ValidateBasic 基本验证
func (n *NeuronInfo) ValidateBasic() error {
	if n.Hotkey == "" {
		return fmt.Errorf("热键不能为空")
	}
	if n.Coldkey == "" {
		return fmt.Errorf("冷键不能为空")
	}
	if n.Netuid == 0 {
		return fmt.Errorf("无效的网络ID: %d", n.Netuid)
	}
	return nil
}
