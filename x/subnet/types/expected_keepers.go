package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

// EventKeeper defines the expected interface for the event module keeper
type EventKeeper interface {
	GetSubnet(ctx sdk.Context, netuid uint16) (eventtypes.SubnetInfo, bool)
	GetSubnetInfo(ctx sdk.Context, netuid uint16) (eventtypes.SubnetInfo, bool)
	GetAllSubnetNetuids(ctx sdk.Context) []uint16
	GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaIn(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOut(ctx sdk.Context, netuid uint16) math.Int
	GetAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	GetMovingAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	GetNeuronInfo(ctx sdk.Context, netuid uint16, uid string) (eventtypes.NeuronInfo, bool)
	GetAllNeuronsBySubnet(ctx sdk.Context, netuid uint16) []eventtypes.NeuronInfo
}

// BlockInflationKeeper defines the expected interface for the blockinflation module keeper
type BlockInflationKeeper interface {
	CalculateSubnetRewards(ctx sdk.Context, blockEmission math.Int, subnetsToEmitTo []uint16) (map[uint16]blockinflationtypes.SubnetRewards, error)
	GetParams(ctx sdk.Context) blockinflationtypes.Params
}

// ERC20Keeper defines the expected interface for the ERC20 module keeper
type ERC20Keeper interface {
	CallEVM(ctx sdk.Context, abi interface{}, from, contract common.Address, commit bool, method string, args ...interface{}) (interface{}, error)
}
