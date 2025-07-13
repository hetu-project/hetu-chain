package types

import (
	"cosmossdk.io/errors"
)

// distribution module errors
var (
	ErrEmptySubnetID          = errors.Register(ModuleName, 1, "subnet ID cannot be empty")
	ErrEmptyValidatorAddress  = errors.Register(ModuleName, 2, "validator address cannot be empty")
	ErrInvalidEpoch           = errors.Register(ModuleName, 3, "invalid epoch")
	ErrInvalidRewardAmount    = errors.Register(ModuleName, 4, "invalid reward amount")
	ErrSubnetNotFound         = errors.Register(ModuleName, 5, "subnet not found")
	ErrValidatorNotFound      = errors.Register(ModuleName, 6, "validator not found")
	ErrInsufficientRewardPool = errors.Register(ModuleName, 7, "insufficient reward pool balance")
)
