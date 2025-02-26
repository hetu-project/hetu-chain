package types

import errorsmod "cosmossdk.io/errors"

// x/checkpointing module errors
var (
	// NOTE: code 1 is reserved for internal errors
	ErrCkptAlreadyExist        = errorsmod.Register(ModuleName, 2, "raw checkpoint already exists")
	ErrCkptHashNotEqual        = errorsmod.Register(ModuleName, 3, "hash does not equal to raw checkpoint")
	ErrCkptDoesNotExist        = errorsmod.Register(ModuleName, 4, "raw checkpoint does not exist")
	ErrBlsKeyAlreadyExist      = errorsmod.Register(ModuleName, 5, "BLS public key already exists")
	ErrInvalidRawCheckpoint    = errorsmod.Register(ModuleName, 6, "raw checkpoint is invalid")
	ErrInvalidCkptStatus       = errorsmod.Register(ModuleName, 7, "raw checkpoint's status is invalid")
	ErrInvalidBlsSignature     = errorsmod.Register(ModuleName, 8, "BLS signature is invalid")
	ErrBlsKeyDoesNotExist      = errorsmod.Register(ModuleName, 9, "BLS public key does not exist")
	ErrConflictingCheckpoint   = errorsmod.Register(ModuleName, 10, "Conflicting checkpoint is found")
	ErrInvalidAppHash          = errorsmod.Register(ModuleName, 11, "Provided app hash is Invalid")
	ErrInsufficientVotingPower = errorsmod.Register(ModuleName, 12, "Accumulated voting power is not greater than 2/3 of total power")
	ErrValAddrDoesNotExist     = errorsmod.Register(ModuleName, 13, "Validator address does not exist")
)
