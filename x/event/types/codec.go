package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterLegacyAminoCodec registers the necessary x/event interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Subnet{}, "hetu/event/Subnet", nil)
	cdc.RegisterConcrete(SubnetInfo{}, "hetu/event/SubnetInfo", nil)
	cdc.RegisterConcrete(NeuronInfo{}, "hetu/event/NeuronInfo", nil)
	cdc.RegisterConcrete(ValidatorStake{}, "hetu/event/ValidatorStake", nil)
	cdc.RegisterConcrete(Delegation{}, "hetu/event/Delegation", nil)
	cdc.RegisterConcrete(ValidatorWeight{}, "hetu/event/ValidatorWeight", nil)
}
