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
	// 旧事件保持不变 (向后兼容)
	SubnetRegisteredTopic  = crypto.Keccak256Hash([]byte("SubnetRegistered(address,uint16,uint256,uint256,address,string)")).Hex()
	StakedSelfTopic        = crypto.Keccak256Hash([]byte("Staked(uint16,address,uint256)")).Hex()
	UnstakedSelfTopic      = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,uint256)")).Hex()
	StakedDelegatedTopic   = crypto.Keccak256Hash([]byte("Staked(uint16,address,address,uint256)")).Hex()
	UnstakedDelegatedTopic = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,address,uint256)")).Hex()
	WeightsSetTopic        = crypto.Keccak256Hash([]byte("WeightsSet(uint16,address,(address,uint256)[])")).Hex()

	// 新事件 Topics
	NetworkRegisteredTopic       = crypto.Keccak256Hash([]byte("NetworkRegistered(uint16,address,address,address,uint256,uint256,uint256,string,(uint16,uint16,uint16,uint16,uint16,uint16,uint16,uint16,uint16,uint16,uint256,uint64,uint16,uint16,uint64,bool,bool,uint64,uint64,uint256,uint256))")).Hex()
	SubnetActivatedTopic         = crypto.Keccak256Hash([]byte("SubnetActivated(uint16,address,uint256,uint256)")).Hex()
	NetworkConfigUpdatedTopic    = crypto.Keccak256Hash([]byte("NetworkConfigUpdated(string,uint256,uint256,address)")).Hex()
	NeuronRegisteredTopic        = crypto.Keccak256Hash([]byte("NeuronRegistered(uint16,address,uint256,bool,bool,string,uint32,string,uint32,uint256)")).Hex()
	NeuronDeregisteredTopic      = crypto.Keccak256Hash([]byte("NeuronDeregistered(uint16,address,uint256)")).Hex()
	NeuronStakeChangedTopic      = crypto.Keccak256Hash([]byte("StakeAllocationChanged(uint16,address,uint256,uint256,uint256)")).Hex()
	ServiceUpdatedTopic          = crypto.Keccak256Hash([]byte("ServiceUpdated(uint16,address,string,uint32,string,uint32,uint256)")).Hex()
	SubnetAllocationChangedTopic = crypto.Keccak256Hash([]byte("SubnetAllocationChanged(address,uint16,uint256,uint256)")).Hex()
	DeallocatedFromSubnetTopic   = crypto.Keccak256Hash([]byte("DeallocatedFromSubnet(address,uint16,uint256)")).Hex()
)

// ----------- Keeper 结构体 -----------
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey

	// 原有合约 ABI (保持兼容)
	subnetRegistryABI   abi.ABI
	stakingSelfABI      abi.ABI
	stakingDelegatedABI abi.ABI
	weightsABI          abi.ABI

	// 新增合约 ABI
	subnetManagerABI abi.ABI
	neuronManagerABI abi.ABI
	globalStakingABI abi.ABI
}

// ----------- Keeper 初始化 -----------
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	subnetRegistryABI abi.ABI,
	stakingSelfABI abi.ABI,
	stakingDelegatedABI abi.ABI,
	weightsABI abi.ABI,
	subnetManagerABI abi.ABI,
	neuronManagerABI abi.ABI,
	globalStakingABI abi.ABI,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		subnetRegistryABI:   subnetRegistryABI,
		stakingSelfABI:      stakingSelfABI,
		stakingDelegatedABI: stakingDelegatedABI,
		weightsABI:          weightsABI,
		subnetManagerABI:    subnetManagerABI,
		neuronManagerABI:    neuronManagerABI,
		globalStakingABI:    globalStakingABI,
	}
}

// ----------- Logger -----------
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/event")
}

// ----------- HandleEvmLogs 集成所有事件 -----------
func (k *Keeper) HandleEvmLogs(ctx sdk.Context, logs []ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理EVM事件", "事件数量", len(logs))

	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		topic := log.Topics[0].Hex()
		k.Logger(ctx).Debug("处理EVM事件", "topic", topic, "合约地址", log.Address.Hex(), "区块高度", log.BlockNumber)

		switch topic {
		// 旧事件 (保持兼容)
		case SubnetRegisteredTopic:
			k.Logger(ctx).Debug("识别到SubnetRegistered事件")
			k.handleSubnetRegistered(ctx, log)
		case StakedSelfTopic:
			k.Logger(ctx).Debug("识别到StakedSelf事件")
			k.handleStaked(ctx, log)
		case UnstakedSelfTopic:
			k.Logger(ctx).Debug("识别到UnstakedSelf事件")
			k.handleUnstaked(ctx, log)
		case StakedDelegatedTopic:
			k.Logger(ctx).Debug("识别到StakedDelegated事件")
			k.handleDelegatedStaked(ctx, log)
		case UnstakedDelegatedTopic:
			k.Logger(ctx).Debug("识别到UnstakedDelegated事件")
			k.handleDelegatedUnstaked(ctx, log)
		case WeightsSetTopic:
			k.Logger(ctx).Debug("识别到WeightsSet事件")
			k.handleWeightsSet(ctx, log)

		// 新事件
		case NetworkRegisteredTopic:
			k.Logger(ctx).Debug("识别到NetworkRegistered事件")
			k.handleNetworkRegistered(ctx, log)
		case SubnetActivatedTopic:
			k.Logger(ctx).Debug("识别到SubnetActivated事件")
			k.handleSubnetActivated(ctx, log)
		case NetworkConfigUpdatedTopic:
			k.Logger(ctx).Debug("识别到NetworkConfigUpdated事件")
			k.handleNetworkConfigUpdated(ctx, log)
		case NeuronRegisteredTopic:
			k.Logger(ctx).Debug("识别到NeuronRegistered事件")
			k.handleNeuronRegistered(ctx, log)
		case NeuronDeregisteredTopic:
			k.Logger(ctx).Debug("识别到NeuronDeregistered事件")
			k.handleNeuronDeregistered(ctx, log)
		case NeuronStakeChangedTopic:
			k.Logger(ctx).Debug("识别到NeuronStakeChanged事件")
			k.handleNeuronStakeChanged(ctx, log)
		case ServiceUpdatedTopic:
			k.Logger(ctx).Debug("识别到ServiceUpdated事件")
			k.handleServiceUpdated(ctx, log)
		case SubnetAllocationChangedTopic:
			k.Logger(ctx).Debug("识别到SubnetAllocationChanged事件")
			k.handleSubnetAllocationChanged(ctx, log)
		case DeallocatedFromSubnetTopic:
			k.Logger(ctx).Debug("识别到DeallocatedFromSubnet事件")
			k.handleDeallocatedFromSubnet(ctx, log)
		default:
			k.Logger(ctx).Debug("未识别的EVM事件topic", "topic", topic, "合约地址", log.Address.Hex())
		}
	}

	k.Logger(ctx).Debug("完成处理EVM事件", "事件数量", len(logs))
}

// ----------- 事件处理方法 -----------

// 子网注册
func (k Keeper) handleSubnetRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理SubnetRegistered事件", "合约地址", log.Address.Hex(), "区块高度", log.BlockNumber)

	var event struct {
		Owner      common.Address
		Netuid     uint16
		LockAmount *big.Int
		BurnedTao  *big.Int
		Pool       common.Address
		Param      string
	}
	if err := k.subnetRegistryABI.UnpackIntoInterface(&event, "SubnetRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("解析SubnetRegistered事件失败", "error", err, "合约地址", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("成功解析SubnetRegistered事件",
		"netuid", event.Netuid,
		"owner", event.Owner.Hex(),
		"lockAmount", event.LockAmount.String(),
		"burnedTao", event.BurnedTao.String(),
		"pool", event.Pool.Hex(),
		"param", event.Param)

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
	k.Logger(ctx).Debug("成功存储子网信息", "netuid", event.Netuid, "owner", event.Owner.Hex())
}

// validator自质押
func (k Keeper) handleStaked(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理Staked事件", "合约地址", log.Address.Hex())

	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Staked", log.Data); err != nil {
		k.Logger(ctx).Error("解析Staked事件失败", "error", err, "合约地址", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("成功解析Staked事件", "netuid", event.Netuid, "validator", event.Validator.Hex(), "amount", event.Amount.String())

	stake, _ := k.GetValidatorStake(ctx, event.Netuid, event.Validator.Hex())
	if stake.Netuid == 0 {
		stake = types.ValidatorStake{Netuid: event.Netuid, Validator: event.Validator.Hex(), Amount: "0"}
	}
	stake.Amount = types.AddBigIntString(stake.Amount, event.Amount.String())
	k.SetValidatorStake(ctx, stake)

	k.Logger(ctx).Debug("成功更新验证者质押", "netuid", event.Netuid, "validator", event.Validator.Hex(), "newAmount", stake.Amount)
}

// validator自解质押
func (k Keeper) handleUnstaked(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理Unstaked事件", "合约地址", log.Address.Hex())

	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Unstaked", log.Data); err != nil {
		k.Logger(ctx).Error("解析Unstaked事件失败", "error", err, "合约地址", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("成功解析Unstaked事件", "netuid", event.Netuid, "validator", event.Validator.Hex(), "amount", event.Amount.String())
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

// ----------- 新事件处理方法 -----------

// NetworkRegistered - 网络注册事件
func (k Keeper) handleNetworkRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理NetworkRegistered事件", "合约地址", log.Address.Hex(), "区块高度", log.BlockNumber)

	var event struct {
		Netuid       uint16
		Owner        common.Address
		AlphaToken   common.Address
		AmmPool      common.Address
		LockedAmount *big.Int
		PoolAmount   *big.Int
		BurnedAmount *big.Int
		Name         string
		Hyperparams  struct {
			// Core network parameters
			Rho                  uint16
			Kappa                uint16
			ImmunityPeriod       uint16
			Tempo                uint16
			MaxValidators        uint16
			ActivityCutoff       uint16
			MaxAllowedUids       uint16
			MaxAllowedValidators uint16
			MinAllowedWeights    uint16
			MaxWeightsLimit      uint16
			// Economic parameters
			BaseNeuronCost        *big.Int
			CurrentDifficulty     uint64
			TargetRegsPerInterval uint16
			MaxRegsPerBlock       uint16
			WeightsRateLimit      uint64
			// Governance parameters
			RegistrationAllowed bool
			CommitRevealEnabled bool
			CommitRevealPeriod  uint64
			ServingRateLimit    uint64
			ValidatorThreshold  *big.Int
			NeuronThreshold     *big.Int
		}
	}

	if err := k.subnetManagerABI.UnpackIntoInterface(&event, "NetworkRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("解析NetworkRegistered事件失败", "error", err, "合约地址", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("成功解析NetworkRegistered事件",
		"netuid", event.Netuid,
		"owner", event.Owner.Hex(),
		"alphaToken", event.AlphaToken.Hex(),
		"ammPool", event.AmmPool.Hex(),
		"lockedAmount", event.LockedAmount.String(),
		"poolAmount", event.PoolAmount.String(),
		"burnedAmount", event.BurnedAmount.String(),
		"name", event.Name,
		"rho", event.Hyperparams.Rho,
		"kappa", event.Hyperparams.Kappa,
		"immunityPeriod", event.Hyperparams.ImmunityPeriod,
		"tempo", event.Hyperparams.Tempo)

	// 处理参数 - Core network parameters 与默认值合并
	params := types.DefaultParamsMap()

	// 覆盖 Core network parameters (如果事件中有值)
	if event.Hyperparams.Rho != 0 {
		params["rho"] = fmt.Sprintf("%.1f", float64(event.Hyperparams.Rho)/10000.0) // uint16 转 float64，假设原值是放大10000倍的
	}
	if event.Hyperparams.Kappa != 0 {
		params["kappa"] = fmt.Sprintf("%.1f", float64(event.Hyperparams.Kappa)/10000.0) // uint16 转 float64，假设原值是放大10000倍的
	}
	if event.Hyperparams.ImmunityPeriod != 0 {
		params["immunity_period"] = fmt.Sprintf("%d", event.Hyperparams.ImmunityPeriod)
	}
	if event.Hyperparams.Tempo != 0 {
		params["tempo"] = fmt.Sprintf("%d", event.Hyperparams.Tempo)
	}
	if event.Hyperparams.MaxValidators != 0 {
		params["max_validators"] = fmt.Sprintf("%d", event.Hyperparams.MaxValidators)
	}
	if event.Hyperparams.ActivityCutoff != 0 {
		params["activity_cutoff"] = fmt.Sprintf("%d", event.Hyperparams.ActivityCutoff)
	}
	if event.Hyperparams.MaxAllowedUids != 0 {
		params["max_allowed_uids"] = fmt.Sprintf("%d", event.Hyperparams.MaxAllowedUids)
	}
	if event.Hyperparams.MaxAllowedValidators != 0 {
		params["max_allowed_validators"] = fmt.Sprintf("%d", event.Hyperparams.MaxAllowedValidators)
	}
	if event.Hyperparams.MinAllowedWeights != 0 {
		params["min_allowed_weights"] = fmt.Sprintf("%d", event.Hyperparams.MinAllowedWeights)
	}
	if event.Hyperparams.MaxWeightsLimit != 0 {
		params["max_weights_limit"] = fmt.Sprintf("%d", event.Hyperparams.MaxWeightsLimit)
	}

	// 直接添加 Economic & Governance parameters
	params["base_neuron_cost"] = event.Hyperparams.BaseNeuronCost.String()
	params["current_difficulty"] = fmt.Sprintf("%d", event.Hyperparams.CurrentDifficulty)
	params["target_regs_per_interval"] = fmt.Sprintf("%d", event.Hyperparams.TargetRegsPerInterval)
	params["max_regs_per_block"] = fmt.Sprintf("%d", event.Hyperparams.MaxRegsPerBlock)
	params["weights_rate_limit"] = fmt.Sprintf("%d", event.Hyperparams.WeightsRateLimit)
	params["weights_set_rate_limit"] = fmt.Sprintf("%d", event.Hyperparams.WeightsRateLimit) // 兼容 stakework
	params["registration_allowed"] = fmt.Sprintf("%t", event.Hyperparams.RegistrationAllowed)
	params["commit_reveal_enabled"] = fmt.Sprintf("%t", event.Hyperparams.CommitRevealEnabled)
	params["commit_reveal_period"] = fmt.Sprintf("%d", event.Hyperparams.CommitRevealPeriod)
	params["serving_rate_limit"] = fmt.Sprintf("%d", event.Hyperparams.ServingRateLimit)
	params["validator_threshold"] = event.Hyperparams.ValidatorThreshold.String()
	params["neuron_threshold"] = event.Hyperparams.NeuronThreshold.String()

	// 创建子网结构
	subnet := types.Subnet{
		Netuid:                event.Netuid,
		Owner:                 event.Owner.Hex(),
		LockAmount:            event.LockedAmount.String(),
		BurnedTao:             event.BurnedAmount.String(),
		Pool:                  event.AmmPool.Hex(),
		Params:                params,
		Mechanism:             1,      // 默认动态机制
		EMAPriceHalvingBlocks: 201600, // 默认4周 (201600块)
	}

	// 存储子网信息
	k.SetSubnet(ctx, subnet)

	// 存储详细的子网信息
	subnetInfo := types.SubnetInfo{
		Netuid:         event.Netuid,
		Owner:          event.Owner.Hex(),
		AlphaToken:     event.AlphaToken.Hex(),
		AmmPool:        event.AmmPool.Hex(),
		LockedAmount:   event.LockedAmount.String(),
		PoolInitialTao: event.PoolAmount.String(),
		BurnedAmount:   event.BurnedAmount.String(),
		CreatedAt:      uint64(ctx.BlockTime().Unix()),
		IsActive:       false, // 初始状态为未激活
		Name:           event.Name,
		Description:    "", // NetworkRegistered 事件中没有 description
	}
	k.SetSubnetInfo(ctx, subnetInfo)
}

// SubnetActivated - 子网激活事件
func (k Keeper) handleSubnetActivated(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid      uint16
		Owner       common.Address
		Timestamp   *big.Int
		BlockNumber *big.Int
	}

	if err := k.subnetManagerABI.UnpackIntoInterface(&event, "SubnetActivated", log.Data); err != nil {
		k.Logger(ctx).Error("parse SubnetActivated failed", "err", err)
		return
	}

	// 更新子网激活状态
	subnet, found := k.GetSubnet(ctx, event.Netuid)
	if found {
		subnet.FirstEmissionBlock = event.BlockNumber.Uint64()
		k.SetSubnet(ctx, subnet)
	}

	// 更新详细子网信息
	subnetInfo, found := k.GetSubnetInfo(ctx, event.Netuid)
	if found {
		subnetInfo.IsActive = true
		subnetInfo.ActivatedAt = event.Timestamp.Uint64()
		subnetInfo.ActivatedBlock = event.BlockNumber.Uint64()
		k.SetSubnetInfo(ctx, subnetInfo)
	}
}

// NetworkConfigUpdated - 网络配置更新事件
func (k Keeper) handleNetworkConfigUpdated(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		ParamName string
		OldValue  *big.Int
		NewValue  *big.Int
		Updater   common.Address
	}

	if err := k.subnetManagerABI.UnpackIntoInterface(&event, "NetworkConfigUpdated", log.Data); err != nil {
		k.Logger(ctx).Error("parse NetworkConfigUpdated failed", "err", err)
		return
	}

	// 这个事件主要用于日志记录，暂时不需要更新状态
	k.Logger(ctx).Info("Network config updated",
		"param", event.ParamName,
		"old_value", event.OldValue.String(),
		"new_value", event.NewValue.String(),
		"updater", event.Updater.Hex())
}

// NeuronRegistered - 神经元注册事件
func (k Keeper) handleNeuronRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("开始处理NeuronRegistered事件", "合约地址", log.Address.Hex(), "区块高度", log.BlockNumber)

	var event struct {
		Netuid                 uint16
		Account                common.Address
		Stake                  *big.Int
		IsValidator            bool
		RequestedValidatorRole bool
		AxonEndpoint           string
		AxonPort               uint32
		PrometheusEndpoint     string
		PrometheusPort         uint32
		BlockNumber            *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&event, "NeuronRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("解析NeuronRegistered事件失败", "error", err, "合约地址", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("成功解析NeuronRegistered事件",
		"netuid", event.Netuid,
		"account", event.Account.Hex(),
		"stake", event.Stake.String(),
		"isValidator", event.IsValidator,
		"requestedValidatorRole", event.RequestedValidatorRole,
		"axonEndpoint", event.AxonEndpoint,
		"axonPort", event.AxonPort,
		"prometheusEndpoint", event.PrometheusEndpoint,
		"prometheusPort", event.PrometheusPort,
		"blockNumber", event.BlockNumber.String())

	// 创建神经元信息
	neuronInfo := types.NeuronInfo{
		Account:                event.Account.Hex(),
		Netuid:                 event.Netuid,
		IsActive:               true,
		IsValidator:            event.IsValidator,
		RequestedValidatorRole: event.RequestedValidatorRole,
		Stake:                  event.Stake.String(),
		RegistrationBlock:      event.BlockNumber.Uint64(),
		LastUpdate:             uint64(ctx.BlockTime().Unix()),
		AxonEndpoint:           event.AxonEndpoint,
		AxonPort:               event.AxonPort,
		PrometheusEndpoint:     event.PrometheusEndpoint,
		PrometheusPort:         event.PrometheusPort,
	}

	k.SetNeuronInfo(ctx, neuronInfo)
	k.Logger(ctx).Debug("成功存储神经元信息", "netuid", event.Netuid, "account", event.Account.Hex(), "isValidator", event.IsValidator)
}

// NeuronDeregistered - 神经元注销事件
func (k Keeper) handleNeuronDeregistered(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid      uint16
		Account     common.Address
		BlockNumber *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&event, "NeuronDeregistered", log.Data); err != nil {
		k.Logger(ctx).Error("parse NeuronDeregistered failed", "err", err)
		return
	}

	// 更新神经元状态为非活跃而不删除
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.Account.Hex())
	if found {
		neuronInfo.IsActive = false
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// StakeAllocationChanged (NeuronManager) - 神经元质押分配变更事件
func (k Keeper) handleNeuronStakeChanged(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid      uint16
		Account     common.Address
		OldStake    *big.Int
		NewStake    *big.Int
		BlockNumber *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&event, "StakeAllocationChanged", log.Data); err != nil {
		k.Logger(ctx).Error("parse NeuronManager StakeAllocationChanged failed", "err", err)
		return
	}

	// 更新神经元质押信息
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.Account.Hex())
	if found {
		neuronInfo.Stake = event.NewStake.String()
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// ServiceUpdated - 服务信息更新事件
func (k Keeper) handleServiceUpdated(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		Netuid             uint16
		Account            common.Address
		AxonEndpoint       string
		AxonPort           uint32
		PrometheusEndpoint string
		PrometheusPort     uint32
		BlockNumber        *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&event, "ServiceUpdated", log.Data); err != nil {
		k.Logger(ctx).Error("parse ServiceUpdated failed", "err", err)
		return
	}

	// 更新神经元服务信息
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.Account.Hex())
	if found {
		neuronInfo.AxonEndpoint = event.AxonEndpoint
		neuronInfo.AxonPort = event.AxonPort
		neuronInfo.PrometheusEndpoint = event.PrometheusEndpoint
		neuronInfo.PrometheusPort = event.PrometheusPort
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// SubnetAllocationChanged (GlobalStaking) - 子网分配变更事件
func (k Keeper) handleSubnetAllocationChanged(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		User      common.Address
		Netuid    uint16
		OldAmount *big.Int
		NewAmount *big.Int
	}

	if err := k.globalStakingABI.UnpackIntoInterface(&event, "SubnetAllocationChanged", log.Data); err != nil {
		k.Logger(ctx).Error("parse SubnetAllocationChanged failed", "err", err)
		return
	}

	// 更新神经元质押信息
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.User.Hex())
	if found {
		neuronInfo.Stake = event.NewAmount.String()
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// DeallocatedFromSubnet (GlobalStaking) - 从子网撤回质押事件
func (k Keeper) handleDeallocatedFromSubnet(ctx sdk.Context, log ethTypes.Log) {
	var event struct {
		User   common.Address
		Netuid uint16
		Amount *big.Int
	}

	if err := k.globalStakingABI.UnpackIntoInterface(&event, "DeallocatedFromSubnet", log.Data); err != nil {
		k.Logger(ctx).Error("parse DeallocatedFromSubnet failed", "err", err)
		return
	}

	// 更新神经元质押信息 (减少质押)
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.User.Hex())
	if found {
		currentStake := types.SubBigIntString(neuronInfo.Stake, event.Amount.String())
		neuronInfo.Stake = currentStake
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
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
	// 优化：使用活跃的神经元信息来生成 ValidatorStake 数据
	// 这样可以确保只返回活跃的验证者，同时保持接口兼容性
	activeNeurons := k.GetActiveNeuronInfosByNetuid(ctx, netuid)

	var stakes []types.ValidatorStake
	for _, neuron := range activeNeurons {
		stake := types.ValidatorStake{
			Netuid:    neuron.Netuid,
			Validator: neuron.Account,
			Amount:    neuron.Stake,
		}
		stakes = append(stakes, stake)
	}

	// 如果没有找到活跃神经元，回退到原始的质押数据查询
	if len(stakes) == 0 {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("stake:"))
		prefixKey := uint16ToBytes(netuid)
		iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var stake types.ValidatorStake
			_ = json.Unmarshal(iterator.Value(), &stake)
			stakes = append(stakes, stake)
		}
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

// ---------------- 子网详细信息 ----------------
func (k Keeper) SetSubnetInfo(ctx sdk.Context, subnetInfo types.SubnetInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnetinfo:"))
	bz, _ := json.Marshal(subnetInfo)
	store.Set(uint16ToBytes(subnetInfo.Netuid), bz)
}

func (k Keeper) GetSubnetInfo(ctx sdk.Context, netuid uint16) (types.SubnetInfo, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnetinfo:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil {
		return types.SubnetInfo{}, false
	}
	var subnetInfo types.SubnetInfo
	_ = json.Unmarshal(bz, &subnetInfo)
	return subnetInfo, true
}

func (k Keeper) GetAllSubnetInfos(ctx sdk.Context) []types.SubnetInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("subnetinfo:"))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	var subnetInfos []types.SubnetInfo
	for ; iterator.Valid(); iterator.Next() {
		var subnetInfo types.SubnetInfo
		_ = json.Unmarshal(iterator.Value(), &subnetInfo)
		subnetInfos = append(subnetInfos, subnetInfo)
	}
	return subnetInfos
}

// ---------------- 神经元信息 ----------------
func (k Keeper) SetNeuronInfo(ctx sdk.Context, neuronInfo types.NeuronInfo) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("neuroninfo:"))
	key := append(uint16ToBytes(neuronInfo.Netuid), []byte(":"+neuronInfo.Account)...)
	bz, _ := json.Marshal(neuronInfo)
	store.Set(key, bz)
}

func (k Keeper) GetNeuronInfo(ctx sdk.Context, netuid uint16, account string) (types.NeuronInfo, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("neuroninfo:"))
	key := append(uint16ToBytes(netuid), []byte(":"+account)...)
	bz := store.Get(key)
	if bz == nil {
		return types.NeuronInfo{}, false
	}
	var neuronInfo types.NeuronInfo
	_ = json.Unmarshal(bz, &neuronInfo)
	return neuronInfo, true
}

func (k Keeper) GetAllNeuronInfosByNetuid(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("neuroninfo:"))
	prefixKey := uint16ToBytes(netuid)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()
	var neuronInfos []types.NeuronInfo
	for ; iterator.Valid(); iterator.Next() {
		var neuronInfo types.NeuronInfo
		_ = json.Unmarshal(iterator.Value(), &neuronInfo)
		neuronInfos = append(neuronInfos, neuronInfo)
	}
	return neuronInfos
}

func (k Keeper) GetAllNeuronInfos(ctx sdk.Context) []types.NeuronInfo {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("neuroninfo:"))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	var neuronInfos []types.NeuronInfo
	for ; iterator.Valid(); iterator.Next() {
		var neuronInfo types.NeuronInfo
		_ = json.Unmarshal(iterator.Value(), &neuronInfo)
		neuronInfos = append(neuronInfos, neuronInfo)
	}
	return neuronInfos
}

func (k Keeper) GetActiveNeuronInfosByNetuid(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	allNeurons := k.GetAllNeuronInfosByNetuid(ctx, netuid)
	var activeNeurons []types.NeuronInfo
	for _, neuron := range allNeurons {
		if neuron.IsActive {
			activeNeurons = append(activeNeurons, neuron)
		}
	}
	return activeNeurons
}

func (k Keeper) GetValidatorInfosByNetuid(ctx sdk.Context, netuid uint16) []types.NeuronInfo {
	allNeurons := k.GetAllNeuronInfosByNetuid(ctx, netuid)
	var validators []types.NeuronInfo
	for _, neuron := range allNeurons {
		if neuron.IsActive && neuron.IsValidator {
			validators = append(validators, neuron)
		}
	}
	return validators
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
