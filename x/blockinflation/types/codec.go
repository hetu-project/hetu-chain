package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc references the global blockinflation module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino JSON compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

// RegisterLegacyAminoCodec registers the necessary x/blockinflation interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register concrete types here if needed
}

// RegisterInterfaces registers the x/blockinflation interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register interfaces here if needed
}
