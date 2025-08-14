package keeper

import (
	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

func getSubnetInfo(subnet eventtypes.Subnet) (types.SubnetInfo, bool) {
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
