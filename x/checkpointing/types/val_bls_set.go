package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/hetu-project/hetu/v1/crypto/bls12381"
)

func ValidatorBlsKeySetToBytes(cdc codec.BinaryCodec, valBlsSet *ValidatorWithBlsKeySet) []byte {
	return cdc.MustMarshal(valBlsSet)
}

func BytesToValidatorBlsKeySet(cdc codec.BinaryCodec, bz []byte) (*ValidatorWithBlsKeySet, error) {
	valBlsSet := new(ValidatorWithBlsKeySet)
	err := cdc.Unmarshal(bz, valBlsSet)
	return valBlsSet, err
}

func (ks *ValidatorWithBlsKeySet) GetBLSKeySet() []bls12381.PublicKey {
	var blsKeySet []bls12381.PublicKey
	for _, val := range ks.ValSet {
		blsKeySet = append(blsKeySet, val.BlsPubKey)
	}
	return blsKeySet
}

func (ks *ValidatorWithBlsKeySet) GetTotalPower() uint64 {
	var total uint64
	for _, val := range ks.ValSet {
		total += val.VotingPower
	}

	return total
}
