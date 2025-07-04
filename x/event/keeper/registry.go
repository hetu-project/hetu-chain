// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"encoding/json"
	"strconv"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/x/event/types"
)

// 子网注册表
func (k Keeper) SetSubnetInfo(ctx sdk.Context, info types.SubnetInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_info"))
	bz, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}
	store.Set([]byte(strconv.FormatUint(uint64(info.Netuid), 10)), bz)
}

func (k Keeper) GetSubnetInfo(ctx sdk.Context, netuid uint16) (types.SubnetInfo, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_info"))
	bz := store.Get([]byte(strconv.FormatUint(uint64(netuid), 10)))
	if len(bz) == 0 {
		return types.SubnetInfo{}, false
	}
	var info types.SubnetInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		return types.SubnetInfo{}, false
	}
	return info, true
}

// 更新参数
func (k Keeper) UpdateSubnetParam(ctx sdk.Context, netuid uint16, param string, value string) {
	info, found := k.GetSubnetInfo(ctx, netuid)
	if !found {
		return
	}
	if info.Params == nil {
		info.Params = make(map[string]string)
	}
	info.Params[param] = value
	k.SetSubnetInfo(ctx, info)
}

func (k Keeper) GetSubnetParam(ctx sdk.Context, netuid uint16, param string) (string, bool) {
	info, found := k.GetSubnetInfo(ctx, netuid)
	if !found {
		return "", false
	}
	value, ok := info.Params[param]
	return value, ok
}

// 质押表
func (k Keeper) AddStake(ctx sdk.Context, netuid uint16, staker string, amount string) {
	old, _ := k.GetStake(ctx, netuid, staker)
	newAmount := types.AddBigIntString(old, amount)
	k.SetStake(ctx, netuid, staker, newAmount)
}

func (k Keeper) SubStake(ctx sdk.Context, netuid uint16, staker string, amount string) {
	old, found := k.GetStake(ctx, netuid, staker)
	if !found {
		return
	}
	newAmount := types.SubBigIntString(old, amount)
	k.SetStake(ctx, netuid, staker, newAmount)
}

func (k Keeper) SetStake(ctx sdk.Context, netuid uint16, staker string, amount string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"+strconv.FormatUint(uint64(netuid), 10)))
	store.Set([]byte(staker), []byte(amount))
}

func (k Keeper) GetStake(ctx sdk.Context, netuid uint16, staker string) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"+strconv.FormatUint(uint64(netuid), 10)))
	bz := store.Get([]byte(staker))
	if len(bz) == 0 {
		return "", false
	}
	return string(bz), true
}
