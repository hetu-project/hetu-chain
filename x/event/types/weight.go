package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorWeight represents a validator's weight assignments to other validators in a subnet
// The weights determine how much influence each validator has in the consensus mechanism
type ValidatorWeight struct {
	Netuid    uint16            `json:"netuid"`    // Subnet ID
	Validator string            `json:"validator"` // Validator address (bech32 encoded cosmos validator address)
	Weights   map[string]uint64 `json:"weights"`   // Map of validator weights (bech32 encoded cosmos account addresses to weight values)
}

// GetValidatorAddress returns the validator address as sdk.ValAddress
func (vw ValidatorWeight) GetValidatorAddress() (sdk.ValAddress, error) {
	if vw.Validator == "" {
		return nil, fmt.Errorf("empty validator address")
	}
	return sdk.ValAddressFromBech32(vw.Validator)
}

// SetValidatorAddress sets the validator address from sdk.ValAddress
func (vw *ValidatorWeight) SetValidatorAddress(addr sdk.ValAddress) {
	vw.Validator = addr.String()
}

// GetWeight returns the weight for a specific validator
func (vw ValidatorWeight) GetWeight(addr sdk.AccAddress) uint64 {
	return vw.Weights[addr.String()]
}

// SetWeight sets the weight for a specific validator
func (vw *ValidatorWeight) SetWeight(addr sdk.AccAddress, weight uint64) {
	if vw.Weights == nil {
		vw.Weights = make(map[string]uint64)
	}
	vw.Weights[addr.String()] = weight
}
