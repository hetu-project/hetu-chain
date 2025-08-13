package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
)

// EventKeeper event module interface - used to get data from event module
type EventKeeper interface {
	// Subnet related
	GetSubnet(ctx sdk.Context, netuid uint16) (eventtypes.Subnet, bool)
	GetAllSubnets(ctx sdk.Context) []eventtypes.Subnet

	// Stake related
	GetValidatorStake(ctx sdk.Context, netuid uint16, validator string) (eventtypes.ValidatorStake, bool)
	GetAllValidatorStakesByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.ValidatorStake

	// Weight related
	GetValidatorWeight(ctx sdk.Context, netuid uint16, validator string) (eventtypes.ValidatorWeight, bool)

	// New: Neuron info interfaces (optional, for future optimization)
	GetActiveNeuronInfosByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.NeuronInfo
	GetValidatorInfosByNetuid(ctx sdk.Context, netuid uint16) []eventtypes.NeuronInfo
}
