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

// GetSubnetMultiParamUpdatedEvent retrieves a SubnetMultiParamUpdatedEvent by tx hash
func (k Keeper) GetSubnetMultiParamUpdatedEvent(ctx sdk.Context, txHash string) (types.SubnetMultiParamUpdatedEvent, bool) {
	event := types.SubnetMultiParamUpdatedEvent{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSubnetMultiParamUpdated)
	bz := store.Get([]byte(txHash))
	if len(bz) == 0 {
		return event, false
	}
	if err := json.Unmarshal(bz, &event); err != nil {
		return event, false
	}
	return event, true
}

// IterateSubnetMultiParamUpdatedEvents iterates all SubnetMultiParamUpdatedEvents
func (k Keeper) IterateSubnetMultiParamUpdatedEvents(ctx sdk.Context, fn func(index int64, event types.SubnetMultiParamUpdatedEvent) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSubnetMultiParamUpdated)
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		event := types.SubnetMultiParamUpdatedEvent{}
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			continue
		}
		if fn(i, event) {
			break
		}
		i++
	}
}

// AllSubnetMultiParamUpdatedEvents returns all SubnetMultiParamUpdatedEvents
func (k Keeper) AllSubnetMultiParamUpdatedEvents(ctx sdk.Context) []types.SubnetMultiParamUpdatedEvent {
	events := []types.SubnetMultiParamUpdatedEvent{}
	k.IterateSubnetMultiParamUpdatedEvents(ctx, func(_ int64, event types.SubnetMultiParamUpdatedEvent) (stop bool) {
		events = append(events, event)
		return false
	})
	return events
}

// GetTaoStakedEvent retrieves a TaoStakedEvent by tx hash
func (k Keeper) GetTaoStakedEvent(ctx sdk.Context, txHash string) (types.TaoStakedEvent, bool) {
	event := types.TaoStakedEvent{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTaoStaked)
	bz := store.Get([]byte(txHash))
	if len(bz) == 0 {
		return event, false
	}
	if err := json.Unmarshal(bz, &event); err != nil {
		return event, false
	}
	return event, true
}

// IterateTaoStakedEvents iterates all TaoStakedEvents
func (k Keeper) IterateTaoStakedEvents(ctx sdk.Context, fn func(index int64, event types.TaoStakedEvent) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTaoStaked)
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		event := types.TaoStakedEvent{}
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			continue
		}
		if fn(i, event) {
			break
		}
		i++
	}
}

// AllTaoStakedEvents returns all TaoStakedEvents
func (k Keeper) AllTaoStakedEvents(ctx sdk.Context) []types.TaoStakedEvent {
	events := []types.TaoStakedEvent{}
	k.IterateTaoStakedEvents(ctx, func(_ int64, event types.TaoStakedEvent) (stop bool) {
		events = append(events, event)
		return false
	})
	return events
}

// GetTaoUnstakedEvent retrieves a TaoUnstakedEvent by tx hash
func (k Keeper) GetTaoUnstakedEvent(ctx sdk.Context, txHash string) (types.TaoUnstakedEvent, bool) {
	event := types.TaoUnstakedEvent{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTaoUnstaked)
	bz := store.Get([]byte(txHash))
	if len(bz) == 0 {
		return event, false
	}
	if err := json.Unmarshal(bz, &event); err != nil {
		return event, false
	}
	return event, true
}

// IterateTaoUnstakedEvents iterates all TaoUnstakedEvents
func (k Keeper) IterateTaoUnstakedEvents(ctx sdk.Context, fn func(index int64, event types.TaoUnstakedEvent) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixTaoUnstaked)
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		event := types.TaoUnstakedEvent{}
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			continue
		}
		if fn(i, event) {
			break
		}
		i++
	}
}

// AllTaoUnstakedEvents returns all TaoUnstakedEvents
func (k Keeper) AllTaoUnstakedEvents(ctx sdk.Context) []types.TaoUnstakedEvent {
	events := []types.TaoUnstakedEvent{}
	k.IterateTaoUnstakedEvents(ctx, func(_ int64, event types.TaoUnstakedEvent) (stop bool) {
		events = append(events, event)
		return false
	})
	return events
}
