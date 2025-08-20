package types

import (
	"cosmossdk.io/math"
)

// SubnetRewards represents the calculated rewards for a subnet
type SubnetRewards struct {
	Netuid   uint16
	TaoIn    math.Int
	AlphaIn  math.Int
	AlphaOut math.Int
	OwnerCut math.Int // Owner cut amount
}
