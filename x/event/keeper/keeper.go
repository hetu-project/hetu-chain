// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

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

// ----------- Event topic constants -----------
// Deprecated: Use Keeper instance fields (k.topicXXX) instead of these global constants
var (
	// Legacy events remain unchanged (backward compatibility)
	SubnetRegisteredTopic  = crypto.Keccak256Hash([]byte("SubnetRegistered(address,uint16,uint256,uint256,address,string)")).Hex()
	StakedSelfTopic        = crypto.Keccak256Hash([]byte("Staked(uint16,address,uint256)")).Hex()
	UnstakedSelfTopic      = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,uint256)")).Hex()
	StakedDelegatedTopic   = crypto.Keccak256Hash([]byte("Staked(uint16,address,address,uint256)")).Hex()
	UnstakedDelegatedTopic = crypto.Keccak256Hash([]byte("Unstaked(uint16,address,address,uint256)")).Hex()
	WeightsSetTopic        = crypto.Keccak256Hash([]byte("WeightsSet(uint16,address,(address,uint256)[])")).Hex()

	// New event topics
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

// ----------- Keeper struct -----------
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey

	// Legacy contract ABIs (maintain compatibility)
	subnetRegistryABI   abi.ABI
	stakingSelfABI      abi.ABI
	stakingDelegatedABI abi.ABI
	weightsABI          abi.ABI

	// New contract ABIs
	subnetManagerABI abi.ABI
	neuronManagerABI abi.ABI
	globalStakingABI abi.ABI

	// Event topic IDs
	topicSubnetRegistered        string
	topicStakedSelf              string
	topicUnstakedSelf            string
	topicStakedDelegated         string
	topicUnstakedDelegated       string
	topicWeightsSet              string
	topicNetworkRegistered       string
	topicSubnetActivated         string
	topicNetworkConfigUpdated    string
	topicNeuronRegistered        string
	topicNeuronDeregistered      string
	topicNeuronStakeChanged      string
	topicServiceUpdated          string
	topicSubnetAllocationChanged string
	topicDeallocatedFromSubnet   string
}

// ----------- Keeper initialization -----------
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
	k := &Keeper{
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
	// Legacy events
	k.topicSubnetRegistered = k.subnetRegistryABI.Events["SubnetRegistered"].ID.Hex()
	k.topicStakedSelf = k.stakingSelfABI.Events["Staked"].ID.Hex()
	k.topicUnstakedSelf = k.stakingSelfABI.Events["Unstaked"].ID.Hex()
	k.topicStakedDelegated = k.stakingDelegatedABI.Events["Staked"].ID.Hex()
	k.topicUnstakedDelegated = k.stakingDelegatedABI.Events["Unstaked"].ID.Hex()
	k.topicWeightsSet = k.weightsABI.Events["WeightsSet"].ID.Hex()
	// New events
	k.topicNetworkRegistered = k.subnetManagerABI.Events["NetworkRegistered"].ID.Hex()
	k.topicSubnetActivated = k.subnetManagerABI.Events["SubnetActivated"].ID.Hex()
	k.topicNetworkConfigUpdated = k.subnetManagerABI.Events["NetworkConfigUpdated"].ID.Hex()
	k.topicNeuronRegistered = k.neuronManagerABI.Events["NeuronRegistered"].ID.Hex()
	k.topicNeuronDeregistered = k.neuronManagerABI.Events["NeuronDeregistered"].ID.Hex()
	k.topicNeuronStakeChanged = k.neuronManagerABI.Events["StakeAllocationChanged"].ID.Hex()
	k.topicServiceUpdated = k.neuronManagerABI.Events["ServiceUpdated"].ID.Hex()
	k.topicSubnetAllocationChanged = k.globalStakingABI.Events["SubnetAllocationChanged"].ID.Hex()
	k.topicDeallocatedFromSubnet = k.globalStakingABI.Events["DeallocatedFromSubnet"].ID.Hex()
	return k
}

// ----------- Logger -----------
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/event")
}

// ----------- HandleEvmLogs integrate all events -----------
func (k *Keeper) HandleEvmLogs(ctx sdk.Context, logs []ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process EVM events", "event_count", len(logs))

	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		topic := log.Topics[0].Hex()
		k.Logger(ctx).Debug("Processing EVM event", "topic", topic, "contract_address", log.Address.Hex(), "block_height", log.BlockNumber)

		switch topic {
		// Legacy events (maintain compatibility)
		case k.topicSubnetRegistered:
			k.Logger(ctx).Debug("Identified SubnetRegistered event")
			k.handleSubnetRegistered(ctx, log)
		case k.topicStakedSelf:
			k.Logger(ctx).Debug("Identified StakedSelf event")
			k.handleStaked(ctx, log)
		case k.topicUnstakedSelf:
			k.Logger(ctx).Debug("Identified UnstakedSelf event")
			k.handleUnstaked(ctx, log)
		case k.topicStakedDelegated:
			k.Logger(ctx).Debug("Identified StakedDelegated event")
			k.handleDelegatedStaked(ctx, log)
		case k.topicUnstakedDelegated:
			k.Logger(ctx).Debug("Identified UnstakedDelegated event")
			k.handleDelegatedUnstaked(ctx, log)
		case k.topicWeightsSet:
			k.Logger(ctx).Debug("Identified WeightsSet event")
			k.handleWeightsSet(ctx, log)

		// New events
		case k.topicNetworkRegistered:
			k.Logger(ctx).Debug("Identified NetworkRegistered event")
			k.handleNetworkRegistered(ctx, log)
		case k.topicSubnetActivated:
			k.Logger(ctx).Debug("Identified SubnetActivated event")
			k.handleSubnetActivated(ctx, log)
		case k.topicNetworkConfigUpdated:
			k.Logger(ctx).Debug("Identified NetworkConfigUpdated event")
			k.handleNetworkConfigUpdated(ctx, log)
		case k.topicNeuronRegistered:
			k.Logger(ctx).Debug("Identified NeuronRegistered event")
			k.handleNeuronRegistered(ctx, log)
		case k.topicNeuronDeregistered:
			k.Logger(ctx).Debug("Identified NeuronDeregistered event")
			k.handleNeuronDeregistered(ctx, log)
		case k.topicNeuronStakeChanged:
			k.Logger(ctx).Debug("Identified NeuronStakeChanged event")
			k.handleNeuronStakeChanged(ctx, log)
		case k.topicServiceUpdated:
			k.Logger(ctx).Debug("Identified ServiceUpdated event")
			k.handleServiceUpdated(ctx, log)
		case k.topicSubnetAllocationChanged:
			k.Logger(ctx).Debug("Identified SubnetAllocationChanged event")
			k.handleSubnetAllocationChanged(ctx, log)
		case k.topicDeallocatedFromSubnet:
			k.Logger(ctx).Debug("Identified DeallocatedFromSubnet event")
			k.handleDeallocatedFromSubnet(ctx, log)
		default:
			k.Logger(ctx).Debug("Unrecognized EVM event topic", "topic", topic, "contract_address", log.Address.Hex())
		}
	}

	k.Logger(ctx).Debug("Finished processing EVM events", "event_count", len(logs))
}

// ----------- Event handler methods -----------

// Subnet registration
func (k Keeper) handleSubnetRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process SubnetRegistered event", "contract_address", log.Address.Hex(), "block_height", log.BlockNumber)

	var event struct {
		Owner        common.Address
		Netuid       uint16
		LockedAmount *big.Int
		BurnedAmount *big.Int
		AmmPool      common.Address
		Param        string
	}
	if err := k.subnetRegistryABI.UnpackIntoInterface(&event, "SubnetRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse SubnetRegistered event", "error", err, "contract_address", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("Successfully parsed SubnetRegistered event",
		"netuid", event.Netuid,
		"owner", event.Owner.Hex(),
		"lockAmount", event.LockedAmount.String(),
		"burnedTao", event.BurnedAmount.String(),
		"pool", event.AmmPool.Hex(),
		"param", event.Param)

	params := types.DefaultParamsMap() // Default parameters map implemented in types
	userParams := map[string]string{}
	_ = json.Unmarshal([]byte(event.Param), &userParams)
	for k, v := range userParams {
		params[k] = v
	}
	subnet := types.Subnet{
		Netuid:                event.Netuid,
		Owner:                 event.Owner.Hex(),
		LockedAmount:          event.LockedAmount.String(),
		BurnedAmount:          event.BurnedAmount.String(),
		AmmPool:               event.AmmPool.Hex(),
		Params:                params,
		Mechanism:             1,      // Default dynamic mechanism
		EMAPriceHalvingBlocks: 201600, // Default 4 weeks (201600 blocks)
	}
	k.SetSubnet(ctx, subnet)
	k.Logger(ctx).Debug("Successfully stored subnet information", "netuid", event.Netuid, "owner", event.Owner.Hex())

	// 初始化子网的AlphaIn和TaoIn，使用LockedAmount作为初始值
	lockedAmount, ok := math.NewIntFromString(event.LockedAmount.String())
	if !ok {
		k.Logger(ctx).Error("Failed to parse LockedAmount", "netuid", event.Netuid, "amount", event.LockedAmount.String())
		lockedAmount = math.ZeroInt()
	}

	// 设置初始的SubnetAlphaIn和SubnetTaoIn
	k.SetSubnetAlphaIn(ctx, event.Netuid, lockedAmount)
	k.SetSubnetTaoIn(ctx, event.Netuid, lockedAmount)

	k.Logger(ctx).Info("Initialized subnet AlphaIn and TaoIn",
		"netuid", event.Netuid,
		"amount", lockedAmount.String())

	// 通知blockinflation模块立即同步AMM池状态
	// 注意：这需要blockinflation模块提供一个公开的接口来触发同步
	// 这里我们通过事件来通知，blockinflation模块需要监听这个事件
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"subnet_registered_sync_amm",
			sdk.NewAttribute("netuid", fmt.Sprintf("%d", event.Netuid)),
			sdk.NewAttribute("amm_pool", event.AmmPool.Hex()),
		),
	)
}

// Validator self-staking
func (k Keeper) handleStaked(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process Staked event", "contract_address", log.Address.Hex())

	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Staked", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse Staked event", "error", err, "contract_address", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("Successfully parsed Staked event", "netuid", event.Netuid, "validator", event.Validator.Hex(), "amount", event.Amount.String())

	stake, _ := k.GetValidatorStake(ctx, event.Netuid, event.Validator.Hex())
	if stake.Netuid == 0 {
		stake = types.ValidatorStake{Netuid: event.Netuid, Validator: event.Validator.Hex(), Amount: "0"}
	}
	stake.Amount = types.AddBigIntString(stake.Amount, event.Amount.String())
	k.SetValidatorStake(ctx, stake)

	k.Logger(ctx).Debug("Successfully updated validator stake", "netuid", event.Netuid, "validator", event.Validator.Hex(), "newAmount", stake.Amount)
}

// Validator self-unstaking
func (k Keeper) handleUnstaked(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process Unstaked event", "contract_address", log.Address.Hex())

	var event struct {
		Netuid    uint16
		Validator common.Address
		Amount    *big.Int
	}
	if err := k.stakingSelfABI.UnpackIntoInterface(&event, "Unstaked", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse Unstaked event", "error", err, "contract_address", log.Address.Hex())
		return
	}

	k.Logger(ctx).Debug("Successfully parsed Unstaked event", "netuid", event.Netuid, "validator", event.Validator.Hex(), "amount", event.Amount.String())
	stake, _ := k.GetValidatorStake(ctx, event.Netuid, event.Validator.Hex())
	if stake.Netuid == 0 {
		return
	}
	stake.Amount = types.SubBigIntString(stake.Amount, event.Amount.String())
	k.SetValidatorStake(ctx, stake)
}

// Delegated staking
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

// Delegated unstaking
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

// Weights set
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

// ----------- New event handler methods -----------

// NetworkRegistered - Network registration event
func (k Keeper) handleNetworkRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process NetworkRegistered event", "contract_address", log.Address.Hex(), "block_height", log.BlockNumber)

	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("NetworkRegistered event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	owner := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
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

	if err := k.subnetManagerABI.UnpackIntoInterface(&nonIndexed, "NetworkRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse NetworkRegistered event", "error", err, "contract_address", log.Address.Hex())
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
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
	}{
		Netuid:       netuid,
		Owner:        owner,
		AlphaToken:   nonIndexed.AlphaToken,
		AmmPool:      nonIndexed.AmmPool,
		LockedAmount: nonIndexed.LockedAmount,
		PoolAmount:   nonIndexed.PoolAmount,
		BurnedAmount: nonIndexed.BurnedAmount,
		Name:         nonIndexed.Name,
		Hyperparams:  nonIndexed.Hyperparams,
	}

	k.Logger(ctx).Debug("Successfully parsed NetworkRegistered event",
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

	// Handle parameters - Merge Core network parameters with default values
	params := types.DefaultParamsMap()

	// Override Core network parameters (if event has values)
	if event.Hyperparams.Rho != 0 {
		params[types.KeyRho] = strconv.FormatFloat(float64(event.Hyperparams.Rho)/10000.0, 'f', -1, 64)
	}
	if event.Hyperparams.Kappa != 0 {
		params[types.KeyKappa] = fmt.Sprintf("%d", event.Hyperparams.Kappa)
	}
	if event.Hyperparams.ImmunityPeriod != 0 {
		params[types.KeyImmunityPeriod] = fmt.Sprintf("%d", event.Hyperparams.ImmunityPeriod)
	}
	if event.Hyperparams.Tempo != 0 {
		params[types.KeyTempo] = fmt.Sprintf("%d", event.Hyperparams.Tempo)
	}
	if event.Hyperparams.MaxValidators != 0 {
		params[types.KeyMaxValidators] = fmt.Sprintf("%d", event.Hyperparams.MaxValidators)
	}
	if event.Hyperparams.ActivityCutoff != 0 {
		params[types.KeyActivityCutoff] = fmt.Sprintf("%d", event.Hyperparams.ActivityCutoff)
	}
	if event.Hyperparams.MaxAllowedUids != 0 {
		params[types.KeyMaxAllowedUids] = fmt.Sprintf("%d", event.Hyperparams.MaxAllowedUids)
	}
	if event.Hyperparams.MaxAllowedValidators != 0 {
		params[types.KeyMaxAllowedValidators] = fmt.Sprintf("%d", event.Hyperparams.MaxAllowedValidators)
	}
	if event.Hyperparams.MinAllowedWeights != 0 {
		params[types.KeyMinAllowedWeights] = fmt.Sprintf("%d", event.Hyperparams.MinAllowedWeights)
	}
	if event.Hyperparams.MaxWeightsLimit != 0 {
		params[types.KeyMaxWeightsLimit] = fmt.Sprintf("%d", event.Hyperparams.MaxWeightsLimit)
	}

	// Directly add Economic & Governance parameters
	params[types.KeyBaseNeuronCost] = event.Hyperparams.BaseNeuronCost.String()
	params[types.KeyCurrentDifficulty] = fmt.Sprintf("%d", event.Hyperparams.CurrentDifficulty)
	params[types.KeyTargetRegsPerInterval] = fmt.Sprintf("%d", event.Hyperparams.TargetRegsPerInterval)
	params[types.KeyMaxRegsPerBlock] = fmt.Sprintf("%d", event.Hyperparams.MaxRegsPerBlock)
	params[types.KeyWeightsRateLimit] = fmt.Sprintf("%d", event.Hyperparams.WeightsRateLimit)
	params[types.KeyWeightsSetRateLimit] = fmt.Sprintf("%d", event.Hyperparams.WeightsRateLimit) // Compatible with stakework
	params[types.KeyRegistrationAllowed] = fmt.Sprintf("%t", event.Hyperparams.RegistrationAllowed)
	params[types.KeyCommitRevealEnabled] = fmt.Sprintf("%t", event.Hyperparams.CommitRevealEnabled)
	params[types.KeyCommitRevealPeriod] = fmt.Sprintf("%d", event.Hyperparams.CommitRevealPeriod)
	params[types.KeyServingRateLimit] = fmt.Sprintf("%d", event.Hyperparams.ServingRateLimit)
	params[types.KeyValidatorThreshold] = event.Hyperparams.ValidatorThreshold.String()
	params[types.KeyNeuronThreshold] = event.Hyperparams.NeuronThreshold.String()

	// Create subnet structure
	subnet := types.Subnet{
		Netuid:                netuid,
		Owner:                 owner.Hex(),
		LockedAmount:          event.LockedAmount.String(),
		BurnedAmount:          event.BurnedAmount.String(),
		AmmPool:               event.AmmPool.Hex(),
		Params:                params,
		Mechanism:             1,      // Default dynamic mechanism
		EMAPriceHalvingBlocks: 201600, // Default 4 weeks (201600 blocks)
	}

	k.Logger(ctx).Debug("Creating new Subnet",
		"netuid", subnet.Netuid,
		"owner", subnet.Owner,
		"locked_amount", subnet.LockedAmount,
		"burned_amount", subnet.BurnedAmount,
		"amm_pool", subnet.AmmPool,
		"mechanism", subnet.Mechanism,
		"ema_price_halving_blocks", subnet.EMAPriceHalvingBlocks,
		"first_emission_block", subnet.FirstEmissionBlock)

	// Store subnet information
	k.SetSubnet(ctx, subnet)

	// Store detailed subnet information
	subnetInfo := types.SubnetInfo{
		Netuid:         netuid,
		Owner:          owner.Hex(),
		AlphaToken:     event.AlphaToken.Hex(),
		AmmPool:        event.AmmPool.Hex(),
		LockedAmount:   event.LockedAmount.String(),
		PoolInitialTao: event.PoolAmount.String(),
		BurnedAmount:   event.BurnedAmount.String(),
		CreatedAt:      uint64(ctx.BlockTime().Unix()),
		IsActive:       false, // Initial state is inactive
		Name:           event.Name,
		Description:    "", // NetworkRegistered event does not have a description
	}

	k.Logger(ctx).Debug("Creating new SubnetInfo",
		"netuid", subnetInfo.Netuid,
		"owner", subnetInfo.Owner,
		"alpha_token", subnetInfo.AlphaToken,
		"amm_pool", subnetInfo.AmmPool,
		"locked_amount", subnetInfo.LockedAmount,
		"pool_initial_tao", subnetInfo.PoolInitialTao,
		"burned_amount", subnetInfo.BurnedAmount,
		"created_at", subnetInfo.CreatedAt,
		"is_active", subnetInfo.IsActive,
		"name", subnetInfo.Name,
		"activated_at", subnetInfo.ActivatedAt,
		"activated_block", subnetInfo.ActivatedBlock)

	k.SetSubnetInfo(ctx, subnetInfo)

	// 初始化子网的AlphaIn和TaoIn，使用LockedAmount作为初始值
	lockedAmount, ok := math.NewIntFromString(event.LockedAmount.String())
	if !ok {
		k.Logger(ctx).Error("Failed to parse LockedAmount", "netuid", event.Netuid, "amount", event.LockedAmount.String())
		lockedAmount = math.ZeroInt()
	}

	// 设置初始的SubnetAlphaIn和SubnetTaoIn
	k.SetSubnetAlphaIn(ctx, event.Netuid, lockedAmount)
	k.SetSubnetTaoIn(ctx, event.Netuid, lockedAmount)

	k.Logger(ctx).Info("Initialized subnet AlphaIn and TaoIn",
		"netuid", event.Netuid,
		"amount", lockedAmount.String())

	// 通知blockinflation模块立即同步AMM池状态
	// 注意：这需要blockinflation模块提供一个公开的接口来触发同步
	// 这里我们通过事件来通知，blockinflation模块需要监听这个事件
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"subnet_registered_sync_amm",
			sdk.NewAttribute("netuid", fmt.Sprintf("%d", event.Netuid)),
			sdk.NewAttribute("amm_pool", event.AmmPool.Hex()),
		),
	)

	k.Logger(ctx).Info("Network registered successfully",
		"netuid", netuid,
		"name", event.Name,
		"owner", owner.Hex())
}

// SubnetActivated - Subnet activation event
func (k Keeper) handleSubnetActivated(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("SubnetActivated event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	owner := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		Timestamp   *big.Int
		BlockNumber *big.Int
	}

	if err := k.subnetManagerABI.UnpackIntoInterface(&nonIndexed, "SubnetActivated", log.Data); err != nil {
		k.Logger(ctx).Error("parse SubnetActivated failed", "err", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
		Netuid      uint16
		Owner       common.Address
		Timestamp   *big.Int
		BlockNumber *big.Int
	}{
		Netuid:      netuid,
		Owner:       owner,
		Timestamp:   nonIndexed.Timestamp,
		BlockNumber: nonIndexed.BlockNumber,
	}

	k.Logger(ctx).Debug("Processing SubnetActivated event",
		"netuid", event.Netuid,
		"owner", event.Owner.Hex(),
		"timestamp", event.Timestamp.String(),
		"event_block_number", event.BlockNumber.String(),
		"current_block_height", ctx.BlockHeight())

	// Update subnet activation status
	subnet, found := k.GetSubnet(ctx, event.Netuid)
	if found {
		k.Logger(ctx).Debug("Before update: Subnet found",
			"netuid", subnet.Netuid,
			"owner", subnet.Owner,
			"first_emission_block", subnet.FirstEmissionBlock)

		subnet.FirstEmissionBlock = event.BlockNumber.Uint64()
		k.SetSubnet(ctx, subnet)

		k.Logger(ctx).Debug("After update: ",
			"netuid", subnet.Netuid,
			"first_emission_block", subnet.FirstEmissionBlock)
	} else {
		k.Logger(ctx).Error("Subnet not found for activation", "netuid", event.Netuid)
	}

	// Update detailed subnet information
	subnetInfo, found := k.GetSubnetInfo(ctx, event.Netuid)
	if found {
		k.Logger(ctx).Debug("Before update: SubnetInfo found",
			"netuid", subnetInfo.Netuid,
			"owner", subnetInfo.Owner,
			"is_active", subnetInfo.IsActive,
			"activated_at", subnetInfo.ActivatedAt,
			"activated_block", subnetInfo.ActivatedBlock)

		subnetInfo.IsActive = true
		subnetInfo.ActivatedAt = event.Timestamp.Uint64()
		subnetInfo.ActivatedBlock = event.BlockNumber.Uint64()
		k.SetSubnetInfo(ctx, subnetInfo)

		k.Logger(ctx).Debug("After update: SubnetInfo activation status updated",
			"netuid", subnetInfo.Netuid,
			"is_active", subnetInfo.IsActive,
			"activated_at", subnetInfo.ActivatedAt,
			"activated_block", subnetInfo.ActivatedBlock)
	} else {
		k.Logger(ctx).Error("SubnetInfo not found for activation", "netuid", event.Netuid)
	}

	// 通知blockinflation模块立即同步AMM池状态
	// 子网激活后也需要同步AMM池状态
	subnet, found = k.GetSubnet(ctx, event.Netuid)
	if found && subnet.AmmPool != "" {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"subnet_registered_sync_amm",
				sdk.NewAttribute("netuid", fmt.Sprintf("%d", event.Netuid)),
				sdk.NewAttribute("amm_pool", subnet.AmmPool),
			),
		)
		k.Logger(ctx).Info("Triggered AMM pool sync after subnet activation",
			"netuid", event.Netuid,
			"amm_pool", subnet.AmmPool)
	}
}

// NetworkConfigUpdated - Network configuration update event
func (k Keeper) handleNetworkConfigUpdated(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("NetworkConfigUpdated event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	// 注意：indexed string只存储哈希，无法直接获取原始字符串
	paramNameHash := log.Topics[1]
	updater := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		OldValue *big.Int
		NewValue *big.Int
	}

	if err := k.subnetManagerABI.UnpackIntoInterface(&nonIndexed, "NetworkConfigUpdated", log.Data); err != nil {
		k.Logger(ctx).Error("parse NetworkConfigUpdated failed", "err", err)
		return
	}

	// 4. 对于indexed string参数，我们可能需要通过已知值映射来恢复
	// 这里简单记录哈希值
	k.Logger(ctx).Info("Network config updated",
		"param_hash", paramNameHash.Hex(),
		"old_value", nonIndexed.OldValue.String(),
		"new_value", nonIndexed.NewValue.String(),
		"updater", updater.Hex())
}

// NeuronRegistered - Neuron registration event
func (k Keeper) handleNeuronRegistered(ctx sdk.Context, log ethTypes.Log) {
	k.Logger(ctx).Debug("Starting to process NeuronRegistered event", "contract_address", log.Address.Hex(), "block_height", log.BlockNumber)

	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("NeuronRegistered event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	account := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		Stake                  *big.Int
		IsValidator            bool
		RequestedValidatorRole bool
		AxonEndpoint           string
		AxonPort               uint32
		PrometheusEndpoint     string
		PrometheusPort         uint32
		BlockNumber            *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&nonIndexed, "NeuronRegistered", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse NeuronRegistered event", "error", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
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
	}{
		Netuid:                 netuid,
		Account:                account,
		Stake:                  nonIndexed.Stake,
		IsValidator:            nonIndexed.IsValidator,
		RequestedValidatorRole: nonIndexed.RequestedValidatorRole,
		AxonEndpoint:           nonIndexed.AxonEndpoint,
		AxonPort:               nonIndexed.AxonPort,
		PrometheusEndpoint:     nonIndexed.PrometheusEndpoint,
		PrometheusPort:         nonIndexed.PrometheusPort,
		BlockNumber:            nonIndexed.BlockNumber,
	}

	k.Logger(ctx).Debug("Processing NeuronRegistered event",
		"netuid", event.Netuid,
		"account", event.Account.Hex(),
		"stake", event.Stake.String(),
		"is_validator", event.IsValidator,
		"requested_validator_role", event.RequestedValidatorRole,
		"axon_endpoint", event.AxonEndpoint,
		"axon_port", event.AxonPort,
		"prometheus_endpoint", event.PrometheusEndpoint,
		"prometheus_port", event.PrometheusPort,
		"block_number", event.BlockNumber.String())

	// Create neuron information
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
	k.Logger(ctx).Debug("Successfully stored neuron information", "netuid", event.Netuid, "account", event.Account.Hex(), "isValidator", event.IsValidator)
}

// NeuronDeregistered - Neuron deregistration event
func (k Keeper) handleNeuronDeregistered(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("NeuronDeregistered event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	account := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		BlockNumber *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&nonIndexed, "NeuronDeregistered", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse NeuronDeregistered event", "error", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
		Netuid      uint16
		Account     common.Address
		BlockNumber *big.Int
	}{
		Netuid:      netuid,
		Account:     account,
		BlockNumber: nonIndexed.BlockNumber,
	}

	k.Logger(ctx).Debug("Processing NeuronDeregistered event",
		"netuid", event.Netuid,
		"account", event.Account.Hex(),
		"block_number", event.BlockNumber.String())

	// Update neuron status to inactive without deleting
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.Account.Hex())
	if found {
		neuronInfo.IsActive = false
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// StakeAllocationChanged (NeuronManager) - Neuron stake allocation change event
func (k Keeper) handleNeuronStakeChanged(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("StakeAllocationChanged event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	account := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		OldStake    *big.Int
		NewStake    *big.Int
		BlockNumber *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&nonIndexed, "StakeAllocationChanged", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse StakeAllocationChanged event", "error", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
		Netuid      uint16
		Account     common.Address
		OldStake    *big.Int
		NewStake    *big.Int
		BlockNumber *big.Int
	}{
		Netuid:      netuid,
		Account:     account,
		OldStake:    nonIndexed.OldStake,
		NewStake:    nonIndexed.NewStake,
		BlockNumber: nonIndexed.BlockNumber,
	}

	k.Logger(ctx).Debug("Processing StakeAllocationChanged event",
		"netuid", event.Netuid,
		"account", event.Account.Hex(),
		"old_stake", event.OldStake.String(),
		"new_stake", event.NewStake.String(),
		"block_number", event.BlockNumber.String())

	// Update neuron stake information
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.Account.Hex())
	if found {
		neuronInfo.Stake = event.NewStake.String()
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// ServiceUpdated - Service information update event
func (k Keeper) handleServiceUpdated(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("ServiceUpdated event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	netuidBI := new(big.Int).SetBytes(log.Topics[1].Bytes())
	netuid := uint16(netuidBI.Uint64())

	account := common.BytesToAddress(log.Topics[2].Bytes())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		AxonEndpoint       string
		AxonPort           uint32
		PrometheusEndpoint string
		PrometheusPort     uint32
		BlockNumber        *big.Int
	}

	if err := k.neuronManagerABI.UnpackIntoInterface(&nonIndexed, "ServiceUpdated", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse ServiceUpdated event", "error", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
		Netuid             uint16
		Account            common.Address
		AxonEndpoint       string
		AxonPort           uint32
		PrometheusEndpoint string
		PrometheusPort     uint32
		BlockNumber        *big.Int
	}{
		Netuid:             netuid,
		Account:            account,
		AxonEndpoint:       nonIndexed.AxonEndpoint,
		AxonPort:           nonIndexed.AxonPort,
		PrometheusEndpoint: nonIndexed.PrometheusEndpoint,
		PrometheusPort:     nonIndexed.PrometheusPort,
		BlockNumber:        nonIndexed.BlockNumber,
	}

	k.Logger(ctx).Debug("Processing ServiceUpdated event",
		"netuid", event.Netuid,
		"account", event.Account.Hex(),
		"axon_endpoint", event.AxonEndpoint,
		"axon_port", event.AxonPort,
		"prometheus_endpoint", event.PrometheusEndpoint,
		"prometheus_port", event.PrometheusPort,
		"block_number", event.BlockNumber.String())

	// Update neuron service information
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

// SubnetAllocationChanged (GlobalStaking) - Subnet allocation change event
func (k Keeper) handleSubnetAllocationChanged(ctx sdk.Context, log ethTypes.Log) {
	// 1. 首先检查Topics长度
	if len(log.Topics) < 3 {
		k.Logger(ctx).Error("SubnetAllocationChanged event has insufficient topics", "topics_length", len(log.Topics))
		return
	}

	// 2. 从Topics中提取indexed参数
	user := common.BytesToAddress(log.Topics[1].Bytes())

	netuidBI := new(big.Int).SetBytes(log.Topics[2].Bytes())
	netuid := uint16(netuidBI.Uint64())

	// 3. 解析非indexed参数
	var nonIndexed struct {
		OldAmount *big.Int
		NewAmount *big.Int
	}

	if err := k.globalStakingABI.UnpackIntoInterface(&nonIndexed, "SubnetAllocationChanged", log.Data); err != nil {
		k.Logger(ctx).Error("Failed to parse SubnetAllocationChanged event", "error", err)
		return
	}

	// 4. 组装完整的event结构体
	event := struct {
		User      common.Address
		Netuid    uint16
		OldAmount *big.Int
		NewAmount *big.Int
	}{
		User:      user,
		Netuid:    netuid,
		OldAmount: nonIndexed.OldAmount,
		NewAmount: nonIndexed.NewAmount,
	}

	k.Logger(ctx).Debug("Processing SubnetAllocationChanged event",
		"user", event.User.Hex(),
		"netuid", event.Netuid,
		"old_amount", event.OldAmount.String(),
		"new_amount", event.NewAmount.String())

	// Update neuron stake information
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.User.Hex())
	if found {
		neuronInfo.Stake = event.NewAmount.String()
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// DeallocatedFromSubnet (GlobalStaking) - Withdraw stake from subnet event
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

	// Update neuron stake information (reduce stake)
	neuronInfo, found := k.GetNeuronInfo(ctx, event.Netuid, event.User.Hex())
	if found {
		currentStake := types.SubBigIntString(neuronInfo.Stake, event.Amount.String())
		neuronInfo.Stake = currentStake
		neuronInfo.LastUpdate = uint64(ctx.BlockTime().Unix())
		k.SetNeuronInfo(ctx, neuronInfo)
	}
}

// ----------- Utility functions -----------

func uint16ToBytes(u uint16) []byte {
	return []byte{byte(u >> 8), byte(u)}
}

// ----------- Storage/Query methods -----------

// ---------------- Subnet ----------------
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

// ---------------- Stake ----------------
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
	// Optimization: Use active neuron information to generate ValidatorStake data
	// This ensures only active validators are returned while maintaining interface compatibility
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

	// If no active neurons found, fallback to original stake data query
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

// ---------------- Delegation ----------------
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

// ---------------- Weight ----------------
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

// ---------------- Stake Aggregation ----------------
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
		// if subnet.Netuid != 0 { // Filter out root subnet
		// 	netuids = append(netuids, subnet.Netuid)
		// }
		netuids = append(netuids, subnet.Netuid)
	}
	return netuids
}

// GetSubnetsToEmitTo returns subnets that have first emission block number set
func (k Keeper) GetSubnetsToEmitTo(ctx sdk.Context) []uint16 {
	currentBlock := ctx.BlockHeight()
	k.Logger(ctx).Debug("GetSubnetsToEmitTo called", "current_block_height", currentBlock)

	subnets := k.GetAllSubnets(ctx)
	k.Logger(ctx).Debug("GetAllSubnets returned", "subnet_count", len(subnets))

	var netuids []uint16
	for _, subnet := range subnets {
		k.Logger(ctx).Debug("Checking subnet for emission eligibility",
			"netuid", subnet.Netuid,
			"first_emission_block", subnet.FirstEmissionBlock,
			"current_block", currentBlock,
			"is_eligible", subnet.FirstEmissionBlock > 0 && uint64(currentBlock) >= subnet.FirstEmissionBlock)

		// if subnet.Netuid != 0 && subnet.FirstEmissionBlock > 0 { // Filter out root subnet and subnets without first emission block set
		// 	netuids = append(netuids, subnet.Netuid)
		// }
		if subnet.FirstEmissionBlock > 0 { // Filter out subnets without first emission block set
			// Additional check: only include subnets where current block height >= first emission block
			if uint64(currentBlock) >= subnet.FirstEmissionBlock {
				netuids = append(netuids, subnet.Netuid)
				k.Logger(ctx).Debug("Subnet added to emission list", "netuid", subnet.Netuid)
			} else {
				k.Logger(ctx).Debug("Subnet not yet eligible for emission",
					"netuid", subnet.Netuid,
					"first_emission_block", subnet.FirstEmissionBlock,
					"blocks_remaining", subnet.FirstEmissionBlock-uint64(currentBlock))
			}
		} else {
			k.Logger(ctx).Debug("Subnet has no first emission block set", "netuid", subnet.Netuid)
		}
	}

	k.Logger(ctx).Debug("GetSubnetsToEmitTo returning", "eligible_subnet_count", len(netuids), "netuids", netuids)
	return netuids
}

// ---------------- Subnet Info ----------------
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

// ---------------- Neuron Info ----------------
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

// ---------------- Price Related Storage ----------------

// SetSubnetMovingPrice sets the moving price for a subnet
func (k Keeper) SetSubnetMovingPrice(ctx sdk.Context, netuid uint16, price math.LegacyDec) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("moving_price:"))
	priceBytes := []byte(price.String())
	store.Set(uint16ToBytes(netuid), priceBytes)
}

// GetSubnetMovingPrice gets the moving price for a subnet
func (k Keeper) GetSubnetMovingPrice(ctx sdk.Context, netuid uint16) math.LegacyDec {
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

// SetSubnetAlphaIn sets the Alpha in amount for a subnet
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

// ---------------- Price Calculation Functions ----------------

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
	subnetTAO := k.GetSubnetTaoIn(ctx, netuid)
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
	currentAmount := k.GetSubnetTaoIn(ctx, netuid)
	newAmount := currentAmount.Add(amount)
	k.SetSubnetTaoIn(ctx, netuid, newAmount)

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

// SetBlocksSinceLastStep sets the block counter since last epoch
func (k Keeper) SetBlocksSinceLastStep(ctx sdk.Context, netuid uint16, value uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("blocks_since_last_step:"))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, value)
	store.Set(uint16ToBytes(netuid), bz)
}

// GetBlocksSinceLastStep gets the block counter since last epoch
func (k Keeper) GetBlocksSinceLastStep(ctx sdk.Context, netuid uint16) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("blocks_since_last_step:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil || len(bz) != 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// SetLastMechanismStepBlock sets the block number of last epoch run
func (k Keeper) SetLastMechanismStepBlock(ctx sdk.Context, netuid uint16, blockHeight int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("last_mechanism_step_block:"))
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(blockHeight))
	store.Set(uint16ToBytes(netuid), bz)
}

// GetLastMechanismStepBlock gets the block number of last epoch run
func (k Keeper) GetLastMechanismStepBlock(ctx sdk.Context, netuid uint16) int64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte("last_mechanism_step_block:"))
	bz := store.Get(uint16ToBytes(netuid))
	if bz == nil || len(bz) != 8 {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

// GetSubnetTaoIn gets the TAO amount for a subnet
func (k Keeper) GetSubnetTaoIn(ctx sdk.Context, netuid uint16) math.Int {
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

// GetSubnetTAO gets the TAO amount for a subnet (alias for GetSubnetTaoIn)
func (k Keeper) GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int {
	return k.GetSubnetTaoIn(ctx, netuid)
}
