// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
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
		Netuid:                event.Netuid,
		Owner:                 event.Owner.Hex(),
		LockAmount:            event.LockAmount.String(),
		BurnedTao:             event.BurnedTao.String(),
		Pool:                  event.Pool.Hex(),
		Params:                params,
		Mechanism:             1,      // 默认动态机制
		EMAPriceHalvingBlocks: 201600, // 默认4周 (201600块)
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
	// Add debug information
	if k.storeKey == nil {
		k.Logger(ctx).Error("GetAllSubnets: storeKey is nil")
		return []types.Subnet{}
	}

	store := ctx.KVStore(k.storeKey)
	if store == nil {
		k.Logger(ctx).Error("GetAllSubnets: KVStore is nil")
		return []types.Subnet{}
	}

	prefixStore := prefix.NewStore(store, []byte("subnet:"))
	iterator := storetypes.KVStorePrefixIterator(prefixStore, nil)
	defer iterator.Close()
	var subnets []types.Subnet
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.Subnet
		err := json.Unmarshal(iterator.Value(), &subnet)
		if err != nil {
			k.Logger(ctx).Error("GetAllSubnets: failed to unmarshal subnet", "error", err, "value", string(iterator.Value()))
			continue
		}
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
	// Add debug information
	if k.storeKey == nil {
		k.Logger(ctx).Error("GetSubnetCount: storeKey is nil")
		return 0
	}

	store := ctx.KVStore(k.storeKey)
	if store == nil {
		k.Logger(ctx).Error("GetSubnetCount: KVStore is nil")
		return 0
	}

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

// ---------------- 价格相关存储 ----------------

// SetSubnetMovingPrice sets the moving price for a subnet
func (k Keeper) SetSubnetMovingPrice(ctx sdk.Context, netuid uint16, price math.LegacyDec) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("moving_price:"))
	priceBytes := []byte(price.String())
	store.Set(uint16ToBytes(netuid), priceBytes)
}

// GetSubnetMovingPrice gets the moving price for a subnet
func (k Keeper) GetSubnetMovingPrice(ctx sdk.Context, netuid uint16) math.LegacyDec {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("moving_price:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.LegacyNewDec(1) // Default to 1.0
	}
	price, err := math.LegacyNewDecFromStr(string(bz))
	if err != nil {
		return math.LegacyNewDec(1) // Default to 1.0 on error
	}
	return price
}

// SetSubnetTAO sets the TAO amount for a subnet
func (k Keeper) SetSubnetTAO(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetTAO gets the TAO amount for a subnet
func (k Keeper) GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// SetSubnetAlphaIn sets the Alpha in amount for a subnet
func (k Keeper) SetSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetAlphaIn gets the Alpha in amount for a subnet
func (k Keeper) GetSubnetAlphaIn(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// SetSubnetAlphaOut sets the Alpha out amount for a subnet
func (k Keeper) SetSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_out:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetAlphaOut gets the Alpha out amount for a subnet
func (k Keeper) GetSubnetAlphaOut(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_out:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// SetSubnetVolume sets the volume for a subnet
func (k Keeper) SetSubnetVolume(ctx sdk.Context, netuid uint16, volume math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_volume:"))
	volumeBytes := []byte(volume.String())
	store.Set(uint16ToBytes(netuid), volumeBytes)
}

// GetSubnetVolume gets the volume for a subnet
func (k Keeper) GetSubnetVolume(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_volume:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	volume, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return volume
}

// SetSubnetEmissionData sets emission data for a subnet
func (k Keeper) SetSubnetEmissionData(ctx sdk.Context, netuid uint16, data types.SubnetEmissionData) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_emission:"))
	bz, _ := json.Marshal(data)
	store.Set(uint16ToBytes(netuid), bz)
}

// GetSubnetEmissionData gets emission data for a subnet
func (k Keeper) GetSubnetEmissionData(ctx sdk.Context, netuid uint16) (types.SubnetEmissionData, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_emission:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return types.SubnetEmissionData{}, false
	}
	var data types.SubnetEmissionData
	_ = json.Unmarshal(bz, &data)
	return data, true
}

// ---------------- 价格计算函数 ----------------

// GetAlphaPrice calculates the alpha price for a subnet
func (k Keeper) GetAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec {
	// Root subnet always has price 1.0
	if netuid == 0 {
		return math.LegacyNewDec(1)
	}

	// Get subnet mechanism
	subnet, exists := k.GetSubnet(ctx, netuid)
	if !exists {
		return math.LegacyNewDec(1) // Default to 1.0
	}

	// Stable mechanism (0) always has price 1.0
	if subnet.Mechanism == 0 {
		return math.LegacyNewDec(1)
	}

	// Dynamic mechanism: price = TAO / AlphaIn
	subnetTAO := k.GetSubnetTAO(ctx, netuid)
	subnetAlphaIn := k.GetSubnetAlphaIn(ctx, netuid)

	if subnetAlphaIn.IsZero() {
		return math.LegacyZeroDec()
	}

	// Convert to LegacyDec for division
	taoDec := math.LegacyNewDecFromInt(subnetTAO)
	alphaInDec := math.LegacyNewDecFromInt(subnetAlphaIn)

	return taoDec.Quo(alphaInDec)
}

// GetMovingAlphaPrice gets the moving alpha price for a subnet
func (k Keeper) GetMovingAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec {
	// Root subnet always has price 1.0
	if netuid == 0 {
		return math.LegacyNewDec(1)
	}

	// Get subnet mechanism
	subnet, exists := k.GetSubnet(ctx, netuid)
	if !exists {
		return math.LegacyNewDec(1) // Default to 1.0
	}

	// Stable mechanism (0) always has price 1.0
	if subnet.Mechanism == 0 {
		return math.LegacyNewDec(1)
	}

	// Dynamic mechanism: return moving price
	return k.GetSubnetMovingPrice(ctx, netuid)
}

// UpdateMovingPrice updates the moving price for a subnet
func (k Keeper) UpdateMovingPrice(ctx sdk.Context, netuid uint16, movingAlpha math.LegacyDec, halvingBlocks uint64) {
	// Get first emission block
	firstEmissionBlock, exists := k.GetSubnetFirstEmissionBlock(ctx, netuid)
	if !exists {
		k.Logger(ctx).Error("subnet first emission block not found", "netuid", netuid)
		return
	}

	currentBlock := ctx.BlockHeight()

	// Calculate blocks since start call
	startCallBlock := int64(firstEmissionBlock) - 1
	if startCallBlock < 0 {
		startCallBlock = 0
	}

	blocksSinceStartCall := currentBlock - startCallBlock
	if blocksSinceStartCall < 0 {
		blocksSinceStartCall = 0
	}

	// Convert to LegacyDec
	blocksSinceStartCallDec := math.LegacyNewDec(blocksSinceStartCall)
	halvingTimeDec := math.LegacyNewDec(int64(halvingBlocks))

	// Calculate alpha: alpha = moving_alpha * (blocks_since_start / (blocks_since_start + halving_time))
	denominator := blocksSinceStartCallDec.Add(halvingTimeDec)
	if denominator.IsZero() {
		denominator = math.LegacyNewDec(1)
	}

	alpha := movingAlpha.Mul(blocksSinceStartCallDec.Quo(denominator))

	// Calculate 1 - alpha
	oneMinusAlpha := math.LegacyNewDec(1).Sub(alpha)

	// Get current price (capped at 1.0)
	currentPrice := k.GetAlphaPrice(ctx, netuid)
	if currentPrice.GT(math.LegacyNewDec(1)) {
		currentPrice = math.LegacyNewDec(1)
	}

	// Get current moving price
	currentMoving := k.GetMovingAlphaPrice(ctx, netuid)

	// Calculate new moving price: new_moving = alpha * current_price + (1 - alpha) * current_moving
	newMoving := alpha.Mul(currentPrice).Add(oneMinusAlpha.Mul(currentMoving))

	// Set new moving price
	k.SetSubnetMovingPrice(ctx, netuid, newMoving)

	k.Logger(ctx).Debug("Updated moving price",
		"netuid", netuid,
		"current_price", currentPrice.String(),
		"current_moving", currentMoving.String(),
		"new_moving", newMoving.String(),
		"alpha", alpha.String(),
		"moving_alpha", movingAlpha.String(),
		"halving_blocks", halvingBlocks,
	)
}

// AddSubnetAlphaIn adds amount to the Alpha in amount for a subnet
func (k Keeper) AddSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetAlphaIn(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetAlphaIn(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet Alpha in",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// AddSubnetAlphaOut adds amount to the Alpha out amount for a subnet
func (k Keeper) AddSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetAlphaOut(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetAlphaOut(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet Alpha out",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// AddSubnetTAO adds amount to the TAO amount for a subnet
func (k Keeper) AddSubnetTAO(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetTAO(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetTAO(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet TAO",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetSubnetAlphaInEmission sets the cumulative Alpha in emission for a subnet
func (k Keeper) SetSubnetAlphaInEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in_emission:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetAlphaInEmission gets the cumulative Alpha in emission for a subnet
func (k Keeper) GetSubnetAlphaInEmission(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_in_emission:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// AddSubnetAlphaInEmission adds to the cumulative Alpha in emission for a subnet
func (k Keeper) AddSubnetAlphaInEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetAlphaInEmission(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetAlphaInEmission(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet Alpha in emission",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetSubnetAlphaOutEmission sets the cumulative Alpha out emission for a subnet
func (k Keeper) SetSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_out_emission:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetAlphaOutEmission gets the cumulative Alpha out emission for a subnet
func (k Keeper) GetSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_alpha_out_emission:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// AddSubnetAlphaOutEmission adds to the cumulative Alpha out emission for a subnet
func (k Keeper) AddSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetAlphaOutEmission(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetAlphaOutEmission(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet Alpha out emission",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetSubnetTaoInEmission sets the cumulative TAO in emission for a subnet
func (k Keeper) SetSubnetTaoInEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao_in_emission:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetSubnetTaoInEmission gets the cumulative TAO in emission for a subnet
func (k Keeper) GetSubnetTaoInEmission(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnet_tao_in_emission:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// AddSubnetTaoInEmission adds to the cumulative TAO in emission for a subnet
func (k Keeper) AddSubnetTaoInEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetSubnetTaoInEmission(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetTaoInEmission(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to subnet TAO in emission",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetPendingOwnerCut sets the pending owner cut for a subnet
func (k Keeper) SetPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("pending_owner_cut:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetPendingOwnerCut gets the pending owner cut for a subnet
func (k Keeper) GetPendingOwnerCut(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("pending_owner_cut:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// AddPendingOwnerCut adds to the pending owner cut for a subnet
func (k Keeper) AddPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetPendingOwnerCut(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetPendingOwnerCut(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to pending owner cut",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetPendingEmission sets the pending emission for a subnet
func (k Keeper) SetPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("pending_emission:"))
	amountBytes := []byte(amount.String())
	store.Set(uint16ToBytes(netuid), amountBytes)
}

// GetPendingEmission gets the pending emission for a subnet
func (k Keeper) GetPendingEmission(ctx sdk.Context, netuid uint16) math.Int {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("pending_emission:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return math.ZeroInt()
	}
	amount, ok := math.NewIntFromString(string(bz))
	if !ok {
		return math.ZeroInt()
	}
	return amount
}

// AddPendingEmission adds to the pending emission for a subnet
func (k Keeper) AddPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int) {
	currentAmount := k.GetPendingEmission(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetPendingEmission(ctx, netuid, newAmount)

	k.Logger(ctx).Debug("Added to pending emission",
		"netuid", netuid,
		"added_amount", amount.String(),
		"new_total", newAmount.String(),
	)
}

// SetBlocksSinceLastStep 设置自上次epoch以来的区块计数器
func (k Keeper) SetBlocksSinceLastStep(ctx sdk.Context, netuid uint16, value uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("blocks_since_last_step:"))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, value)
	store.Set(uint16ToBytes(netuid), bz)
}

// GetBlocksSinceLastStep 获取自上次epoch以来的区块计数器
func (k Keeper) GetBlocksSinceLastStep(ctx sdk.Context, netuid uint16) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("blocks_since_last_step:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil || len(bz) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// SetLastMechanismStepBlock 设置上次epoch运行的区块号
func (k Keeper) SetLastMechanismStepBlock(ctx sdk.Context, netuid uint16, blockHeight int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("last_mechanism_step_block:"))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))
	store.Set(uint16ToBytes(netuid), bz)
}

// GetLastMechanismStepBlock 获取上次epoch运行的区块号
func (k Keeper) GetLastMechanismStepBlock(ctx sdk.Context, netuid uint16) int64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("last_mechanism_step_block:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil || len(bz) != 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}
