package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EventKeeper 事件模块接口 - 用于从event模块获取数据
type EventKeeper interface {
	// 子网相关
	GetSubnet(ctx sdk.Context, netuid uint16) (Subnet, bool)
	GetAllSubnets(ctx sdk.Context) []Subnet

	// 质押相关
	GetValidatorStake(ctx sdk.Context, netuid uint16, validator string) (ValidatorStake, bool)
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []ValidatorStake

	// 权重相关
	GetValidatorWeight(ctx sdk.Context, netuid uint16, validator string) (ValidatorWeight, bool)
}

// 从event模块导入的类型定义
type Subnet struct {
	Netuid     uint16            `json:"netuid"`
	Owner      string            `json:"owner"`
	LockAmount string            `json:"lock_amount"`
	BurnedTao  string            `json:"burned_tao"`
	Pool       string            `json:"pool"`
	Params     map[string]string `json:"params"`
}

type ValidatorStake struct {
	Netuid    uint16 `json:"netuid"`
	Validator string `json:"validator"`
	Amount    string `json:"amount"`
}

type ValidatorWeight struct {
	Netuid    uint16            `json:"netuid"`
	Validator string            `json:"validator"`
	Weights   map[string]uint64 `json:"weights"`
}
