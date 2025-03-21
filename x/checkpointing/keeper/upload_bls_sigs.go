package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

type UploadBlsSigState struct {
	cdc         codec.BinaryCodec
	uploadBlsSigs storetypes.KVStore
}

// BLSCallback handles the BLS signature upload
func (k Keeper) BLSCallback(ctx context.Context, msg *types.MsgBLSCallback) (*types.MsgBLSCallbackResponse, error) {
	store := k.UploadBlsSigState(ctx)
	err := store.StoreBLSSignature(msg.EpochNum, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgBLSCallbackResponse{}, nil
}

func (k Keeper) UploadBlsSigState(ctx context.Context) UploadBlsSigState {
	// Build the CheckpointsState storage
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return UploadBlsSigState{
		cdc:         k.cdc,
		uploadBlsSigs: prefix.NewStore(storeAdapter, types.EpochKeyToValBlsSigKey),
	}
}

// StoreBLSSignature stores the BLS signatures for a given epoch
func (us UploadBlsSigState) StoreBLSSignature(epoch uint64, msg *types.MsgBLSCallback) error {
	if len(msg.AddrSigs) == 0 {
		return types.ErrInvalidBlsSignature.Wrap("empty BLS signatures")
	}

	// Check if there are existing signatures for the given epoch
	existingBytes := us.uploadBlsSigs.Get(sdk.Uint64ToBigEndian(epoch))
	if existingBytes != nil {
		var existingMsg types.MsgBLSCallback
		if err := us.cdc.Unmarshal(existingBytes, &existingMsg); err != nil {
			return err
		}

		// Merge the new signatures with the existing ones
		for _, sig := range msg.AddrSigs {
			existingMsg.AddrSigs = append(existingMsg.AddrSigs, sig)
		}
		msg = &existingMsg
	}

	addrSigsBytes, err := us.cdc.Marshal(msg)
	if err != nil {
		return err
	}

	us.uploadBlsSigs.Set(sdk.Uint64ToBigEndian(epoch), addrSigsBytes)
	return nil
}

// GetBLSSignatures retrieves the BLS signatures for a given epoch
func (us UploadBlsSigState) GetBLSSignatures(epoch uint64) (*types.MsgBLSCallback, error) {
	ckptsKey := sdk.Uint64ToBigEndian(epoch)
	rawBytes := us.uploadBlsSigs.Get(ckptsKey)
	if rawBytes == nil {
		return nil, types.ErrReportBlsSigDoesNotExist.Wrapf("no BLS signatures found for epoch %v", epoch)
	}

	var msg types.MsgBLSCallback
	if err := us.cdc.Unmarshal(rawBytes, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
