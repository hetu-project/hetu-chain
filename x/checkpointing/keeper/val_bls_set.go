package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// GetValidatorBlsKeySet returns the set of validators of a given epoch with BLS public key
// the validators are ordered by their address in ascending order
func (k Keeper) GetValidatorBlsKeySet(ctx context.Context, epochNumber uint64) *types.ValidatorWithBlsKeySet {
	store := k.valBlsSetStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	valBlsKeySetBytes := store.Get(epochNumberBytes)
	valBlsKeySet, err := types.BytesToValidatorBlsKeySet(k.cdc, valBlsKeySetBytes)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal validator BLS key set: %w", err))
	}
	return valBlsKeySet
}

func (k Keeper) GetCurrentValidatorBlsKeySet(ctx context.Context, epochNumber uint64) *types.ValidatorWithBlsKeySet {
	return k.GetValidatorBlsKeySet(ctx, epochNumber)
}

// InitValidatorBLSSet stores the validator set with BLS keys in the beginning of the current epoch
// This is called upon BeginBlock
func (k Keeper) InitValidatorBLSSet(ctx sdk.Context, epochNumber uint64) error {
	// Get the top validators from the staking contract
	valset, dispatcherURLs, blsPublicKeys, err := k.GetTopValidators(ctx, types.DefaultValidatorSize) // Get top 512 validators
	if err != nil {
		return fmt.Errorf("failed to get validators from staking contract: %w", err)
	}
	if len(valset) != len(dispatcherURLs) {
		return fmt.Errorf("validator set and dispatcher URLs set have different lengths")
	}
	if len(valset) != len(blsPublicKeys) {
		return fmt.Errorf("validator set and dispatcher URLs set have different lengths")
	}

	valBlsSet := &types.ValidatorWithBlsKeySet{
		ValSet: make([]*types.ValidatorWithBlsKey, len(valset)),
	}

	for i, val := range valset {
		// Move blspubkey to validator staking contract
		blsPub, err := hex.DecodeString(blsPublicKeys[i])
		if err != nil {
			return fmt.Errorf("failed to decode BLS public key: %w", err)
		}
		valBls := &types.ValidatorWithBlsKey{
			ValidatorAddress: common.BytesToAddress(val.Addr).Hex(),
			BlsPubKey:        blsPub,
			VotingPower:      uint64(val.Power),
			DispatcherUrl:    dispatcherURLs[i],
		}
		valBlsSet.ValSet[i] = valBls
	}

	valBlsSetBytes := types.ValidatorBlsKeySetToBytes(k.cdc, valBlsSet)
	store := k.valBlsSetStore(ctx)
	store.Set(types.ValidatorBlsKeySetKey(epochNumber), valBlsSetBytes)

	return nil
}

// ClearValidatorSet removes the validator BLS set of a given epoch
// TODO: This is called upon the epoch is checkpointed
func (k Keeper) ClearValidatorSet(ctx context.Context, epochNumber uint64) {
	store := k.valBlsSetStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	store.Delete(epochNumberBytes)
}

// valBlsSetStore returns the KVStore of the validator BLS set of a given epoch
// prefix: ValidatorBLSSetKey
// key: epoch number
// value: ValidatorBLSKeySet
func (k Keeper) valBlsSetStore(ctx context.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, types.ValidatorBlsKeySetPrefix)
}
