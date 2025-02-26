package types

import (
	"bytes"
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m BlsSig) Hash() BlsSigHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.BlockHash.MustMarshal(),
		m.BlsSig.MustMarshal(),
		[]byte(m.SignerAddress),
	}
	return hash(fields)
}

func (m RawCheckpoint) Hash() RawCkptHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.BlockHash.MustMarshal(),
		m.BlsMultiSig.MustMarshal(),
		m.Bitmap,
	}
	return hash(fields)
}

func (m RawCheckpoint) HashStr() string {
	return m.Hash().String()
}

// SignedMsg is the message corresponding to the BLS sig in this raw checkpoint
// Its value should be (epoch_number || app_hash)
func (m RawCheckpoint) SignedMsg() []byte {
	return append(sdk.Uint64ToBigEndian(m.EpochNum), *m.BlockHash...)
}

func hash(fields [][]byte) []byte {
	var bz []byte
	for _, b := range fields {
		bz = append(bz, b...)
	}
	return tmhash.Sum(bz)
}

func (m BlsSigHash) Bytes() []byte {
	return m
}

func (m RawCkptHash) Bytes() []byte {
	return m
}

func (m RawCkptHash) Equals(h RawCkptHash) bool {
	return bytes.Equal(m.Bytes(), h.Bytes())
}

func (m RawCkptHash) String() string {
	return hex.EncodeToString(m)
}

func FromStringToCkptHash(s string) (RawCkptHash, error) {
	return hex.DecodeString(s)
}

func GetSignBytes(epoch uint64, hash []byte) []byte {
	return append(sdk.Uint64ToBigEndian(epoch), hash...)
}
