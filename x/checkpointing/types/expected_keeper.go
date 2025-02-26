package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event Hooks
// These can be utilized to communicate between a checkpointing keeper and another
// keeper which must take particular actions when raw checkpoints change
// state. The second keeper must implement this interface, which then the
// checkpointing keeper can call.

// CheckpointingHooks event hooks for raw checkpoint object (noalias)
type CheckpointingHooks interface {
	AfterBlsKeyRegistered(ctx context.Context, valAddr sdk.ValAddress) error         // Must be called when a BLS key is registered
	AfterRawCheckpointSealed(ctx context.Context, epoch uint64) error                // Must be called when a raw checkpoint is SEALED
	AfterRawCheckpointConfirmed(ctx context.Context, epoch uint64) error             // Must be called when a raw checkpoint is CONFIRMED
	AfterRawCheckpointForgotten(ctx context.Context, ckpt *RawCheckpoint) error      // Must be called when a raw checkpoint is FORGOTTEN
	AfterRawCheckpointFinalized(ctx context.Context, epoch uint64) error             // Must be called when a raw checkpoint is FINALIZED
	AfterRawCheckpointBlsSigVerified(ctx context.Context, ckpt *RawCheckpoint) error // Must be called when a raw checkpoint's multi-sig is verified
}
