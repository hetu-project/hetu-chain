package keeper

import (
	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

// getSubnetInfo 从 Subnet 对象中提取 SubnetInfo
func getSubnetInfo(subnet eventtypes.Subnet) (types.SubnetInfo, bool) {
	// 假设 AlphaToken 信息存储在 subnet.Params 中
	alphaToken, ok := subnet.Params["alpha_token"]
	if !ok {
		return types.SubnetInfo{}, false
	}

	return types.SubnetInfo{
		Netuid:                subnet.Netuid,
		Owner:                 subnet.Owner,
		AlphaToken:            alphaToken,
		EMAPriceHalvingBlocks: subnet.EMAPriceHalvingBlocks,
		Params:                subnet.Params,
	}, true
}
