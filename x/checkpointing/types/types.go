package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hetu-project/hetu/v1/crypto/bls12381"
)

const (
	// HashSize is the size in bytes of a hash
	HashSize   = sha256.Size
)

// BlsSigner is an interface for signing BLS messages
type BlsSigner interface {
	SignMsgWithBls(msg []byte) (bls12381.Signature, error)
	BlsPubKey() (bls12381.PublicKey, error)
}

type BlockHash []byte

type BlsSigHash []byte

type RawCkptHash []byte

func NewCheckpoint(epochNum uint64, blockHash BlockHash) *RawCheckpoint {
	return &RawCheckpoint{
		EpochNum:    epochNum,
		BlockHash:   &blockHash,
		Bitmap:      nil,
		BlsMultiSig: nil,
	}
}

func NewCheckpointWithMeta(ckpt *RawCheckpoint, status CheckpointStatus) *RawCheckpointWithMeta {
	return &RawCheckpointWithMeta{
		Ckpt:      ckpt,
		Status:    status,
		Lifecycle: []*CheckpointStateUpdate{},
	}
}

// RecordStateUpdate appends a new state update to the raw ckpt with meta
// where the time/height are captured by the current ctx
func (cm *RawCheckpointWithMeta) RecordStateUpdate(ctx context.Context, status CheckpointStatus) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height, time := sdkCtx.HeaderInfo().Height, sdkCtx.HeaderInfo().Time
	stateUpdate := &CheckpointStateUpdate{
		State:       status,
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	cm.Lifecycle = append(cm.Lifecycle, stateUpdate)
}

func (bh *BlockHash) Unmarshal(bz []byte) error {
	if len(bz) != HashSize {
		return fmt.Errorf(
			"invalid block hash length, expected: %d, got: %d",
			HashSize, len(bz))
	}
	*bh = bz
	return nil
}

func (bh *BlockHash) Size() (n int) {
	if bh == nil {
		return 0
	}
	return len(*bh)
}

func (bh *BlockHash) Equal(l BlockHash) bool {
	return bh.String() == l.String()
}

func (bh *BlockHash) String() string {
	return hex.EncodeToString(*bh)
}

func (bh *BlockHash) MustMarshal() []byte {
	bz, err := bh.Marshal()
	if err != nil {
		panic(err)
	}
	return bz
}

func (bh *BlockHash) Marshal() ([]byte, error) {
	return *bh, nil
}

func (bh BlockHash) MarshalTo(data []byte) (int, error) {
	copy(data, bh)
	return len(data), nil
}

func (bh *BlockHash) ValidateBasic() error {
	if bh == nil {
		return errors.New("invalid block hash")
	}
	if len(*bh) != HashSize {
		return errors.New("invalid block hash")
	}
	return nil
}

// ValidateBasic does sanity checks on a raw checkpoint
func (ckpt RawCheckpoint) ValidateBasic() error {
	err := ckpt.BlockHash.ValidateBasic()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidRawCheckpoint, "error validating block hash: %s", err.Error())
	}
	err = ckpt.BlsMultiSig.ValidateBasic()
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidRawCheckpoint, "error validating BLS multi-signature: %s", err.Error())
	}

	return nil
}

func CkptWithMetaToBytes(cdc codec.BinaryCodec, ckptWithMeta *RawCheckpointWithMeta) []byte {
	return cdc.MustMarshal(ckptWithMeta)
}

func BytesToCkptWithMeta(cdc codec.BinaryCodec, bz []byte) (*RawCheckpointWithMeta, error) {
	ckptWithMeta := new(RawCheckpointWithMeta)
	err := cdc.Unmarshal(bz, ckptWithMeta)
	return ckptWithMeta, err
}
