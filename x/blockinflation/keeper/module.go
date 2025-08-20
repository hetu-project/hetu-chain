package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessBeginBlockEvents 处理BeginBlock事件，包括子网注册事件
func (k Keeper) ProcessBeginBlockEvents(ctx sdk.Context) {
	// 处理事件
	for _, event := range ctx.EventManager().Events() {
		// 检查是否是子网注册事件
		if event.Type == "subnet_registered_sync_amm" {
			k.handleSubnetRegisteredEvent(ctx, event)
		}
	}
}

// 处理子网注册事件
func (k Keeper) handleSubnetRegisteredEvent(ctx sdk.Context, event sdk.Event) {
	var netuid uint16
	var ammPoolAddress string

	// 从事件属性中提取信息
	for _, attr := range event.Attributes {
		if string(attr.Key) == "netuid" {
			netuIdStr := string(attr.Value)
			netuIdInt, err := strconv.ParseUint(netuIdStr, 10, 16)
			if err != nil {
				k.Logger(ctx).Error("Failed to parse netuid from event",
					"netuid_str", netuIdStr,
					"error", err)
				continue
			}
			netuid = uint16(netuIdInt)
		} else if string(attr.Key) == "amm_pool" {
			ammPoolAddress = string(attr.Value)
		}
	}

	// 如果提取到了有效的netuid和ammPoolAddress，立即同步AMM池状态
	if netuid > 0 && ammPoolAddress != "" {
		k.Logger(ctx).Info("Detected subnet registration, triggering AMM pool sync",
			"netuid", netuid,
			"amm_pool_address", ammPoolAddress)

		// 调用处理函数
		k.HandleSubnetRegisteredEvent(ctx, netuid, ammPoolAddress)
	}
}
