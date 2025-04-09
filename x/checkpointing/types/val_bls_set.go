package types

import (
	"math/big"

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

func (ks *ValidatorWithBlsKeySet) GetTotalPower() *big.Int {
	total := big.NewInt(0)
	for _, val := range ks.ValSet {
		valPower := new(big.Int)
		valPower.SetString(val.VotingPower, 10)
		total.Add(total, valPower)
	}
	return total
}
