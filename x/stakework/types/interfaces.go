package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

// EventKeeper 事件模块接口 - 用于从event模块获取数据
type EventKeeper interface {
	// 子网相关
	GetSubnet(ctx sdk.Context, netuid uint16) (eventtypes.Subnet, bool)
	GetAllSubnets(ctx sdk.Context) []eventtypes.Subnet

	// 质押相关
	GetValidatorStake(ctx sdk.Context, netuid uint16, validator string) (eventtypes.ValidatorStake, bool)
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.ValidatorStake

	// 权重相关
	GetValidatorWeight(ctx sdk.Context, netuid uint16, validator string) (eventtypes.ValidatorWeight, bool)

	// 新增：神经元信息接口（可选，用于未来优化）
	GetActiveNeuronInfosByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.NeuronInfo
	GetValidatorInfosByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.NeuronInfo
}
