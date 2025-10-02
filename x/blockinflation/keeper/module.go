package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessBeginBlockEvents Handling BeginBlock events, including subnet registration events
func (k Keeper) ProcessBeginBlockEvents(ctx sdk.Context) {
	// Process events
	for _, event := range ctx.EventManager().Events() {
		// Check if it is a subnet registration event
		if event.Type == "subnet_registered_sync_amm" {
			k.handleSubnetRegisteredEvent(ctx, event)
		}
	}
}

// Handle the subnet registration event
func (k Keeper) handleSubnetRegisteredEvent(ctx sdk.Context, event sdk.Event) {
	var netuid uint16
	var ammPoolAddress string

	// Extract information from the event attributes
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

	// If a valid netuid and ammPoolAddress are extracted, immediately sync the AMM pool state
	if netuid > 0 && ammPoolAddress != "" {
		k.Logger(ctx).Info("Detected subnet registration, triggering AMM pool sync",
			"netuid", netuid,
			"amm_pool_address", ammPoolAddress)

		// Call the handling function
		k.HandleSubnetRegisteredEvent(ctx, netuid, ammPoolAddress)
	}
}
