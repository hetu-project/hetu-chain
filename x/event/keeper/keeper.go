// Copyright 2024 Hetu Project
// This file is part of the Hetu Network packages.

package keeper

import (
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	eventabi "github.com/hetu-project/hetu/v1/x/event/abi"
	"github.com/hetu-project/hetu/v1/x/event/types"
	evmmodulekeeper "github.com/hetu-project/hetu/v1/x/evm/keeper"
)

// 事件 topic hash 常量（需与合约事件签名一致）
var (
	SubnetRegisteredTopic        = crypto.Keccak256Hash([]byte("SubnetRegistered(uint16,address,string,string,string,string,string,uint256,uint256,uint256,uint256,uint256,uint256)")).Hex()
	SubnetRegisteredTopic2       = crypto.Keccak256Hash([]byte("SubnetRegistered(uint16,address,string,string,string,string,string,string,string,string,string,string,string)")).Hex()
	SubnetMultiParamUpdatedTopic = crypto.Keccak256Hash([]byte("SubnetMultiParamUpdated(uint16,string[],uint256[])")).Hex()
	TaoStakedTopic               = crypto.Keccak256Hash([]byte("TaoStaked(uint16,address,uint256)")).Hex()
	TaoUnstakedTopic             = crypto.Keccak256Hash([]byte("TaoUnstaked(uint16,address,uint256)")).Hex()
)

var (
	subnetRegistryABI     abi.ABI
	subnetParamManagerABI abi.ABI
	taoStakingABI         abi.ABI
)

func init() {
	var err error
	subnetRegistryABI, err = abi.JSON(strings.NewReader(string(eventabi.SubnetRegistryABI)))
	if err != nil {
		panic(err)
	}
	subnetParamManagerABI, err = abi.JSON(strings.NewReader(string(eventabi.SubnetParamManagerABI)))
	if err != nil {
		panic(err)
	}
	taoStakingABI, err = abi.JSON(strings.NewReader(string(eventabi.TaoStakingABI)))
	if err != nil {
		panic(err)
	}
}

// 解析 SubnetRegistered 事件
func parseSubnetRegistered(log ethTypes.Log) (types.SubnetInfo, error) {
	var event struct {
		Netuid                uint16
		Owner                 common.Address
		Name                  string
		Github                string
		Discord               string
		Website               string
		Description           string
		Kappa                 *big.Int
		BondsPenalty          *big.Int
		BondsMovingAverage    *big.Int
		AlphaLow              *big.Int
		AlphaHigh             *big.Int
		AlphaSigmoidSteepness *big.Int
	}
	err := subnetRegistryABI.UnpackIntoInterface(&event, "SubnetRegistered", log.Data)
	if err != nil {
		return types.SubnetInfo{}, err
	}
	return types.SubnetInfo{
		Netuid:                event.Netuid,
		Owner:                 event.Owner.Hex(),
		Name:                  event.Name,
		Github:                event.Github,
		Discord:               event.Discord,
		Website:               event.Website,
		Description:           event.Description,
		Kappa:                 event.Kappa.String(),
		BondsPenalty:          event.BondsPenalty.String(),
		BondsMovingAverage:    event.BondsMovingAverage.String(),
		AlphaLow:              event.AlphaLow.String(),
		AlphaHigh:             event.AlphaHigh.String(),
		AlphaSigmoidSteepness: event.AlphaSigmoidSteepness.String(),
	}, nil
}

// 解析 SubnetMultiParamUpdated 事件
func parseSubnetMultiParamUpdated(log ethTypes.Log) (uint16, []string, []string, error) {
	var event struct {
		Netuid uint16
		Params []string
		Values []*big.Int
	}
	err := subnetParamManagerABI.UnpackIntoInterface(&event, "SubnetMultiParamUpdated", log.Data)
	if err != nil {
		return 0, nil, nil, err
	}
	values := make([]string, len(event.Values))
	for i, v := range event.Values {
		values[i] = v.String()
	}
	return event.Netuid, event.Params, values, nil
}

// 解析 TaoStaked 事件
func parseTaoStaked(log ethTypes.Log) (uint16, string, string, error) {
	var event struct {
		Netuid uint16
		Staker common.Address
		Amount *big.Int
	}
	err := taoStakingABI.UnpackIntoInterface(&event, "TaoStaked", log.Data)
	if err != nil {
		return 0, "", "", err
	}
	return event.Netuid, event.Staker.Hex(), event.Amount.String(), nil
}

// 解析 TaoUnstaked 事件
func parseTaoUnstaked(log ethTypes.Log) (uint16, string, string, error) {
	var event struct {
		Netuid uint16
		Staker common.Address
		Amount *big.Int
	}
	err := taoStakingABI.UnpackIntoInterface(&event, "TaoUnstaked", log.Data)
	if err != nil {
		return 0, "", "", err
	}
	return event.Netuid, event.Staker.Hex(), event.Amount.String(), nil
}

// Keeper of this module maintains collections of events.
type Keeper struct {
	cdc       codec.Codec
	storeKey  storetypes.StoreKey
	evmKeeper *evmmodulekeeper.Keeper
}

// NewKeeper returns a new instance of event Keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, evmKeeper *evmmodulekeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		evmKeeper: evmKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/event")
}

// HandleEvmLogs 处理 EVM 日志并更新业务表
func (k *Keeper) HandleEvmLogs(ctx sdk.Context, logs []ethTypes.Log) {
	for _, log := range logs {
		if len(log.Topics) == 0 {
			continue
		}
		topic := log.Topics[0].Hex()
		fmt.Printf("解析事件topic: %s\n", topic)
		fmt.Printf("解析事件SubnetRegisteredTopic: %s\n", SubnetRegisteredTopic)
		fmt.Printf("解析事件SubnetRegisteredTopic2: %s\n", SubnetRegisteredTopic2)
		sig := "SubnetRegistered(uint16,address,string,string,string,string,string,string,string,string,string,string,string)"
		fmt.Println(crypto.Keccak256Hash([]byte(sig)).Hex())
		switch topic {
		case SubnetRegisteredTopic2:
			fmt.Printf("解析子网创建事件\n")
			info, err := parseSubnetRegistered(log)
			if err != nil {
				ctx.Logger().Error("parseSubnetRegistered failed", "err", err)
				continue
			}
			k.SetSubnetInfo(ctx, info)
		case SubnetMultiParamUpdatedTopic:
			netuid, params, values, err := parseSubnetMultiParamUpdated(log)
			if err != nil {
				ctx.Logger().Error("parseSubnetMultiParamUpdated failed", "err", err)
				continue
			}
			for i, param := range params {
				k.UpdateSubnetParam(ctx, netuid, param, values[i])
			}
		case TaoStakedTopic:
			netuid, staker, amount, err := parseTaoStaked(log)
			if err != nil {
				ctx.Logger().Error("parseTaoStaked failed", "err", err)
				continue
			}
			k.SetStake(ctx, netuid, staker, amount)
		case TaoUnstakedTopic:
			netuid, staker, amount, err := parseTaoUnstaked(log)
			if err != nil {
				ctx.Logger().Error("parseTaoUnstaked failed", "err", err)
				continue
			}
			k.SetStake(ctx, netuid, staker, amount)
		}
	}
}
