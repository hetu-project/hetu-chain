package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyEpochWindows = []byte("EpochWindows")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(epochWindows uint64) Params {
	return Params{
		EpochWindows: epochWindows,
	}
}

// DefaultParams returns default erc20 module parameters
func DefaultParams() Params {
	return Params{
		EpochWindows: 5, // Default value from abci.go
	}
}

func (p Params) Validate() error {
	if p.EpochWindows == 0 {
		return fmt.Errorf("epoch windows cannot be 0")
	}

	return validateEpochWindows(p.EpochWindows)
}

func validateEpochWindows(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("epoch windows cannot be 0")
	}

	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEpochWindows, &p.EpochWindows, validateEpochWindows),
	}
}
