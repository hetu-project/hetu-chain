// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"encoding/json"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/event/types"
)

// SetSubnetRegisteredEvent stores a SubnetRegisteredEvent
func (k Keeper) SetSubnetRegisteredEvent(ctx sdk.Context, event types.SubnetRegisteredEvent) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSubnetRegistered)
	bz, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	store.Set([]byte(event.TxHash), bz)
}

// GetSubnetRegisteredEvent retrieves a SubnetRegisteredEvent by tx hash
func (k Keeper) GetSubnetRegisteredEvent(ctx sdk.Context, txHash string) (types.SubnetRegisteredEvent, bool) {
	event := types.SubnetRegisteredEvent{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSubnetRegistered)
	bz := store.Get([]byte(txHash))
	if len(bz) == 0 {
		return event, false
	}
	if err := json.Unmarshal(bz, &event); err != nil {
		return event, false
	}
	return event, true
}

// IterateSubnetRegisteredEvents iterates all SubnetRegisteredEvents
func (k Keeper) IterateSubnetRegisteredEvents(ctx sdk.Context, fn func(index int64, event types.SubnetRegisteredEvent) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSubnetRegistered)
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		event := types.SubnetRegisteredEvent{}
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			continue
		}
		if fn(i, event) {
			break
		}
		i++
	}
}

// AllSubnetRegisteredEvents returns all SubnetRegisteredEvents
func (k Keeper) AllSubnetRegisteredEvents(ctx sdk.Context) []types.SubnetRegisteredEvent {
	events := []types.SubnetRegisteredEvent{}
	k.IterateSubnetRegisteredEvents(ctx, func(_ int64, event types.SubnetRegisteredEvent) (stop bool) {
		events = append(events, event)
		return false
	})
	return events
}

// 其余三类事件（SubnetMultiParamUpdated、TaoStaked、TaoUnstaked）可用类似方法实现
