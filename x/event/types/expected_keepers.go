// SPDX-License-Identifier: MIT
package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EventKeeper defines the standard interface for the event module keeper
// This is the single source of truth for all modules that need to interact with the event module
type EventKeeper interface {
	// Subnet related
	GetSubnet(ctx sdk.Context, netuid uint16) (Subnet, bool)
	GetAllSubnets(ctx sdk.Context) []Subnet
	GetAllSubnetNetuids(ctx sdk.Context) []uint16
	GetSubnetsToEmitTo(ctx sdk.Context) []uint16
	GetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16) (uint64, bool)
	SetSubnetFirstEmissionBlock(ctx sdk.Context, netuid uint16, blockNumber uint64)
	GetSubnetInfo(ctx sdk.Context, netuid uint16) (SubnetInfo, bool)
	GetAllSubnetInfos(ctx sdk.Context) []SubnetInfo

	// Stake related
	GetValidatorStake(ctx sdk.Context, netuid uint16, validator string) (ValidatorStake, bool)
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []ValidatorStake

	// Weight related
	GetValidatorWeight(ctx sdk.Context, netuid uint16, validator string) (ValidatorWeight, bool)

	// Neuron info related
	GetNeuronInfo(ctx sdk.Context, netuid uint16, account string) (NeuronInfo, bool)
	GetActiveNeuronInfosByNetuid(ctx sdk.Context, netuid uint16) []NeuronInfo
	GetValidatorInfosByNetuid(ctx sdk.Context, netuid uint16) []NeuronInfo

	// Price related
	GetAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	GetMovingAlphaPrice(ctx sdk.Context, netuid uint16) math.LegacyDec
	UpdateMovingPrice(ctx sdk.Context, netuid uint16, movingAlpha math.LegacyDec, halvingBlocks uint64)

	// Alpha/TAO tracking
	GetSubnetAlphaIn(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOut(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTAO(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaInEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16) math.Int
	GetSubnetTaoInEmission(ctx sdk.Context, netuid uint16) math.Int
	AddSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTAO(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetAlphaOutEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddSubnetTaoInEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	SetSubnetAlphaOut(ctx sdk.Context, netuid uint16, amount math.Int)
	SetSubnetAlphaIn(ctx sdk.Context, netuid uint16, amount math.Int)
	SetSubnetTaoIn(ctx sdk.Context, netuid uint16, amount math.Int)

	// Emission related
	GetPendingEmission(ctx sdk.Context, netuid uint16) math.Int
	SetPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingEmission(ctx sdk.Context, netuid uint16, amount math.Int)
	GetPendingOwnerCut(ctx sdk.Context, netuid uint16) math.Int
	SetPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)
	AddPendingOwnerCut(ctx sdk.Context, netuid uint16, amount math.Int)

	// Epoch tracking
	GetBlocksSinceLastStep(ctx sdk.Context, netuid uint16) uint64
	SetBlocksSinceLastStep(ctx sdk.Context, netuid uint16, blocks uint64)
	GetLastMechanismStepBlock(ctx sdk.Context, netuid uint16) int64
	SetLastMechanismStepBlock(ctx sdk.Context, netuid uint16, block int64)
}
