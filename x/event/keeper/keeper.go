// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"encoding/json"
	"fmt"
	"math/big"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/hetu-project/hetu/v1/x/event/types"
)

// ----------- 事件 topic 常量 -----------
var (
	SubnetRegisteredTopic  = crypto.Keccak256Hash([]byte("SubnetRegistered(address,uint16,uint256,uint256,address,string)")).Hex()
	StakedSelfTopic        = crypto.Keccak256Hash([]byte("Staked(uint16,address,uint256)")).Hex()
	UnstakedSelfTopic      = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,uint256)")).Hex()
	StakedDelegatedTopic   = crypto.Keccak256Hash([]byte("Staked(uint16,address,address,uint256)")).Hex()
	UnstakedDelegatedTopic = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,address,uint256)")).Hex()
	WeightsSetTopic        = crypto.Keccak256Hash([]byte("WeightsSet(uint16,address,(address,uint256)[])")).Hex()
)

// ----------- Keeper 结构体 -----------
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey

	// 合约 ABI
	subnetRegistryABI   abi.ABI
	stakingSelfABI      abi.ABI
	stakingDelegatedABI abi.ABI
	weightsABI          abi.ABI
}

// ----------- Keeper 初始化 -----------
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	subnetRegistryABI abi.ABI,
	stakingSelfABI abi.ABI,
	stakingDelegatedABI abi.ABI,
	weightsABI abi.ABI,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		subnetRegistryABI:   subnetRegistryABI,
		stakingSelfABI:      stakingSelfABI,
		stakingDelegatedABI: stakingDelegatedABI,
		weightsABI:          weightsABI,
	}
}

// ----------- Logger -----------
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/event")
}

// ----------- HandleEvmLogs 集成所有事件 -----------
func (k *Keeper) HandleEvmLogs(ctx sdk.Context, logs []ethTypes.Log) {
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		topic := log.Topics[0].Hex()
		switch topic {
		case SubnetRegisteredTopic:
			k.handleSubnetRegistered(ctx, log)
		case StakedSelfTopic:
			k.handleStaked(ctx, log)
		case UnstakedSelfTopic:
			k.handleUnstaked(ctx, log)
		case StakedDelegatedTopic:
			k.handleDelegatedStaked(ctx, log)
		case UnstakedDelegatedTopic:
			k.handleDelegatedUnstaked(ctx, log)
		case WeightsSetTopic:
			k.handleWeightsSet(ctx, log)
		default:
			fmt.Printf("未识别的EVM事件topic: %s\n", topic)
		}
	}
}

// ----------- 事件处理方法 -----------

// 子网注册
func (k Keeper) handleSubnetRegistered(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Owner      common.Address
		Netuid     uint16
		LockAmount *big.Int
		BurnedTao  *big.Int
		Pool       common.Address
		Param      string
	}
	if err := k.subnetRegistryABI.UnpackIntoInterface(&event, "SubnetRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("parse SubnetRegistered failed", "err", err)
		return
	}
	params := types.DefaultParamsMap() // 你需在 types 实现 DefaultParamsMap
	userParams := map[string]string{}
	_ = json.Unmarshal([]byte(event.Param), &userParams)
	for k, v := range userParams {
		params[k] = v
	}
	subnet := types.Subnet{
		Netuid:     event.Netuid,
		Owner:      event.Owner.Hex(),
		LockAmount: event.LockAmount.String(),
		BurnedTao:  event.BurnedTao.String(),
		Pool:       event.Pool.Hex(),
		Params:     params,
	}
	k.SetSubnet(ctx, subnet)
}

// validator自质押
func (k Keeper) handleStaked(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Staked", log.Data); err != nil {
		k.Logger(ctx).Error("parse Staked failed", "err", err)
		return
	}
	stake, _ := k.GetValidatorStake(ctx, event.Netuid, event.Validator.Hex())
	if stake.Netuid == 0 {
		stake = types.ValidatorStake{Netuid: event.Netuid, Validator: event.Validator.Hex(), Amount: "0"}
	}
	stake.Amount = types.AddBigIntString(stake.Amount, event.Amount.String())
	k.SetValidatorStake(ctx, stake)
}

// validator自解质押
func (k Keeper) handleUnstaked(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Unstaked", log.Data); err != nil {
		k.Logger(ctx).Error("parse Unstaked failed", "err", err)
		return
	}
	stake, _ := k.GetValidatorStake(ctx, event.Netuid, event.Validator.Hex())
	if stake.Netuid == 0 {
		return
	}
	stake.Amount = types.SubBigIntString(stake.Amount, event.Amount.String())
	k.SetValidatorStake(ctx, stake)
}

// 委托质押
func (k Keeper) handleDelegatedStaked(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid    uint16
		Validator common.Address
		Staker    common.Address
		Amount    *big.Int
	}
	if err := k.stakingDelegatedABI.UnpackIntoInterface(&event, "Staked", log.Data); err != nil {
		k.Logger(ctx).Error("parse DelegatedStaked failed", "err", err)
		return
	}
	deleg, _ := k.GetDelegation(ctx, event.Netuid, event.Validator.Hex(), event.Staker.Hex())
	if deleg.Netuid == 0 {
		deleg = types.Delegation{Netuid: event.Netuid, Validator: event.Validator.Hex(), Staker: event.Staker.Hex(), Amount: "0"}
	}
	deleg.Amount = types.AddBigIntString(deleg.Amount, event.Amount.String())
	k.SetDelegation(ctx, deleg)
}

// 委托解质押
func (k Keeper) handleDelegatedUnstaked(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid    uint16
		Validator common.Address
		Staker    common.Address
		Amount    *big.Int
	}
	if err := k.stakingDelegatedABI.UnpackIntoInterface(&event, "Unstaked", log.Data); err != nil {
		k.Logger(ctx).Error("parse DelegatedUnstaked failed", "err", err)
		return
	}
	deleg, _ := k.GetDelegation(ctx, event.Netuid, event.Validator.Hex(), event.Staker.Hex())
	if deleg.Netuid == 0 {
		return
	}
	deleg.Amount = types.SubBigIntString(deleg.Amount, event.Amount.String())
	k.SetDelegation(ctx, deleg)
}

// 权重矩阵
func (k Keeper) handleWeightsSet(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid    uint16
		Validator common.Address
		Weights   []struct {
			Dest   common.Address
			Weight *big.Int
		}
	}
	if err := k.weightsABI.UnpackIntoInterface(&event, "WeightsSet", log.Data); err != nil {
		k.Logger(ctx).Error("parse WeightsSet failed", "err", err)
		return
	}
	weights := make(map[string]uint64)
	for _, w := range event.Weights {
		weights[w.Dest.Hex()] = w.Weight.Uint64()
	}
	k.SetValidatorWeight(ctx, event.Netuid, event.Validator.Hex(), weights)
}

// ----------- 工具函数 -----------
func uint16ToBytes(u uint16) []byte {
	return []byte{byte(u >> 8), byte(u)}
}

// ----------- 存储/查询方法 -----------

// ---------------- 子网 ----------------
func (k Keeper) SetSubnet(ctx sdk.Context, subnet types.Subnet) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet:"))
	bz, _ := json.Marshal(subnet)
	store.Set(uint16ToBytes(subnet.Netuid), bz)
}

func (k Keeper) GetSubnet(ctx sdk.Context, netuid uint16) (types.Subnet, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return types.Subnet{}, false
	}
	var subnet types.Subnet
	_ = json.Unmarshal(bz, &subnet)
	return subnet, true
}

func (k Keeper) GetAllSubnets(ctx sdk.Context) []types.Subnet {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet:"))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	var subnets []types.Subnet
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.Subnet
		_ = json.Unmarshal(iterator.Value(), &subnet)
		subnets = append(subnets, subnet)
	}
	return subnets
}

// ---------------- 质押 ----------------
func (k Keeper) SetValidatorStake(ctx sdk.Context, stake types.ValidatorStake) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"))
	key := append(uint16ToBytes(stake.Netuid), []byte(":"+stake.Validator)...)
	bz, _ := json.Marshal(stake)
	store.Set(key, bz)
}

func (k Keeper) GetValidatorStake(ctx sdk.Context, netuid uint16, validator string) (types.ValidatorStake, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"))
	key := append(uint16ToBytes(netuid), []byte(":"+validator)...)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorStake{}, false
	}
	var stake types.ValidatorStake
	_ = json.Unmarshal(bz, &stake)
	return stake, true
}

func (k Keeper) GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []types.ValidatorStake {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"))
	prefixKey := uint16ToBytes(netuid)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()
	var stakes []types.ValidatorStake
	for ; iterator.Valid(); iterator.Next() {
		var stake types.ValidatorStake
		_ = json.Unmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}
	return stakes
}

func (k Keeper) GetAllValidatorStakesByValidator(ctx sdk.Context, validator string) []types.ValidatorStake {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	var stakes []types.ValidatorStake
	for ; iterator.Valid(); iterator.Next() {
		var stake types.ValidatorStake
		_ = json.Unmarshal(iterator.Value(), &stake)
		if stake.Validator == validator {
			stakes = append(stakes, stake)
		}
	}
	return stakes
}

// ---------------- 委托 ----------------
func (k Keeper) SetDelegation(ctx sdk.Context, deleg types.Delegation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("delegation:"))
	key := append(uint16ToBytes(deleg.Netuid), []byte(":"+deleg.Validator+":"+deleg.Staker)...)
	bz, _ := json.Marshal(deleg)
	store.Set(key, bz)
}

func (k Keeper) GetDelegation(ctx sdk.Context, netuid uint16, validator, staker string) (types.Delegation, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("delegation:"))
	key := append(uint16ToBytes(netuid), []byte(":"+validator+":"+staker)...)
	bz := store.Get(key)
	if bz == nil {
		return types.Delegation{}, false
	}
	var deleg types.Delegation
	_ = json.Unmarshal(bz, &deleg)
	return deleg, true
}

func (k Keeper) GetDelegationsByStaker(ctx sdk.Context, staker string) []types.Delegation {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("delegation:"))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	var delegs []types.Delegation
	for ; iterator.Valid(); iterator.Next() {
		var deleg types.Delegation
		_ = json.Unmarshal(iterator.Value(), &deleg)
		if deleg.Staker == staker {
			delegs = append(delegs, deleg)
		}
	}
	return delegs
}

func (k Keeper) GetDelegationsByValidator(ctx sdk.Context, netuid uint16, validator string) []types.Delegation {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("delegation:"))
	prefixKey := append(uint16ToBytes(netuid), []byte(":"+validator)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()
	var delegs []types.Delegation
	for ; iterator.Valid(); iterator.Next() {
		var deleg types.Delegation
		_ = json.Unmarshal(iterator.Value(), &deleg)
		delegs = append(delegs, deleg)
	}
	return delegs
}

// ---------------- 权重 ----------------
func (k Keeper) SetValidatorWeight(ctx sdk.Context, netuid uint16, validator string, weights map[string]uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("weight:"))
	key := append(uint16ToBytes(netuid), []byte(":"+validator)...)
	valWeight := types.ValidatorWeight{
		Netuid:    netuid,
		Validator: validator,
		Weights:   weights,
	}
	bz, _ := json.Marshal(valWeight)
	store.Set(key, bz)
}

func (k Keeper) GetValidatorWeight(ctx sdk.Context, netuid uint16, validator string) (types.ValidatorWeight, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("weight:"))
	key := append(uint16ToBytes(netuid), []byte(":"+validator)...)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorWeight{}, false
	}
	var valWeight types.ValidatorWeight
	_ = json.Unmarshal(bz, &valWeight)
	return valWeight, true
}

// ---------------- 质押量聚合 ----------------
func (k Keeper) GetAllValidatorStakesAmount(ctx sdk.Context, netuid uint16) map[string]string {
	result := make(map[string]string)
	for _, stake := range k.GetAllValidatorStakesByNetuid(ctx, netuid) {
		result[stake.Validator] = stake.Amount
	}
	for _, stake := range k.GetAllValidatorStakesByNetuid(ctx, netuid) {
		for _, deleg := range k.GetDelegationsByValidator(ctx, netuid, stake.Validator) {
			result[stake.Validator] = types.AddBigIntString(result[stake.Validator], deleg.Amount)
		}
	}
	return result
}

// GetSubnetCount returns the total number of subnets
func (k Keeper) GetSubnetCount(ctx sdk.Context) uint64 {
	subnets := k.GetAllSubnets(ctx)
	return uint64(len(subnets))
}

// GetAllSubnetNetuids returns all subnet netuids (excluding root subnet 0)
func (k Keeper) GetAllSubnetNetuids(ctx sdk.Context) []uint16 {
	subnets := k.GetAllSubnets(ctx)
	var netuids []uint16
	for _, subnet := range subnets {
		if subnet.Netuid != 0 { // 过滤掉根子网
			netuids = append(netuids, subnet.Netuid)
		}
	}
	return netuids
}

// GetSubnetsToEmitTo returns subnets that have first emission block number set
func (k Keeper) GetSubnetsToEmitTo(ctx sdk.Context) []uint16 {
	subnets := k.GetAllSubnets(ctx)
	var emitSubnets []uint16
	for _, subnet := range subnets {
		if subnet.Netuid != 0 && subnet.FirstEmissionBlock > 0 { // 过滤掉根子网和未设置首次排放区块的子网
			emitSubnets = append(emitSubnets, subnet.Netuid)
		}
	}
	return emitSubnets
}

// SetSubnetFirstEmissionBlock sets the first emission block number for a subnet
func (k Keeper) SetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16, blockNumber uint64) {
	subnet, exists := k.GetSubnet(ctx, netuid)
	if !exists {
		k.Logger(ctx).Error("subnet not found", "netuid", netuid)
		return
	}
	subnet.FirstEmissionBlock = blockNumber
	k.SetSubnet(ctx, subnet)
}

// GetSubnetFirstEmissionBlock gets the first emission block number for a subnet
func (k Keeper) GetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16) (uint64, bool) {
	subnet, exists := k.GetSubnet(ctx, netuid)
	if !exists {
		return 0, false
	}
	return subnet.FirstEmissionBlock, true
}
