package bls12381

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	blst "github.com/supranational/blst/bindings/go"
)

// Tests single BLS sig verification
func TestVerifyBlsSig(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	sk, pk := GenKeyPair()
	sig := Sign(sk, msga)
	// a byte size of a sig (compressed) is 48
	require.Equal(t, 48, len(sig))
	// a byte size of a public key (compressed) is 96
	require.Equal(t, 96, len(pk))
	res, err := Verify(sig, pk, msga)
	require.True(t, res)
	require.Nil(t, err)
	res, err = Verify(sig, pk, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

// Tests BLS multi sig verification
func TestVerifyBlsMultiSig(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 105
	sks, pks := generateBatchTestKeyPairs(n)
	sigs := make([]Signature, n)
	for i := 0; i < n; i++ {
		sigs[i] = Sign(sks[i], msga)
	}
	multiSig, err := AggrSigList(sigs)
	require.Nil(t, err)
	res, err := VerifyMultiSig(multiSig, pks, msga)
	require.True(t, res)
	require.Nil(t, err)
	res, err = VerifyMultiSig(multiSig, pks, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

// Tests BLS multi sig verification
// insert an invalid BLS sig in aggregation
func TestVerifyBlsMultiSig2(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	sigs := make([]Signature, n)
	for i := 0; i < n-1; i++ {
		sigs[i] = Sign(sks[i], msga)
	}
	sigs[n-1] = Sign(sks[n-1], msgb)
	multiSig, err := AggrSigList(sigs)
	require.Nil(t, err)
	res, err := VerifyMultiSig(multiSig, pks, msga)
	require.False(t, res)
	require.Nil(t, err)
	res, err = VerifyMultiSig(multiSig, pks, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

func TestAccumulativeAggregation(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	var aggPK PublicKey
	var aggSig Signature
	var err error
	var res bool
	for i := 0; i < n-1; i++ {
		sig := Sign(sks[i], msga)
		aggSig, err = AggrSig(aggSig, sig)
		require.Nil(t, err)
		aggPK, err = AggrPK(aggPK, pks[i])
		require.Nil(t, err)
		res, err = Verify(aggSig, aggPK, msga)
		require.True(t, res)
		require.Nil(t, err)
	}
	sig := Sign(sks[n-1], msgb)
	aggSig, err = AggrSig(aggSig, sig)
	require.Nil(t, err)
	aggPK, err = AggrPK(aggPK, pks[n-1])
	require.Nil(t, err)
	res, err = Verify(aggSig, aggPK, msga)
	require.False(t, res)
	require.Nil(t, err)
}

func TestSKToPK(t *testing.T) {
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	for i := 0; i < n; i++ {
		ok := sks[i].PubKey().Equal(pks[i])
		require.True(t, ok)
	}
}

func generateBatchTestKeyPairs(n int) ([]PrivateKey, []PublicKey) {
	sks := make([]PrivateKey, n)
	pubks := make([]PublicKey, n)
	for i := 0; i < n; i++ {
		sk, pk := GenKeyPair()
		sks[i] = sk
		pubks[i] = pk
	}
	return sks, pubks
}


func TestNewBlsPubKeyFromHex(t *testing.T) {
	// Generate a valid key pair
	skSerialized := GenPrivKey()
	sk := new(blst.SecretKey)
	sk.Deserialize(skSerialized)
	pk := new(BlsPubKey).From(sk)
	pk_ser := pk.Serialize()
	pk_compressed := pk.Compress()
	hexStr := hex.EncodeToString(pk_ser)

	// Test with a valid hex string
	newPk, err := NewBlsPubKeyFromHex(hexStr)
	require.Nil(t, err)
	require.Equal(t, pk_compressed, newPk.Bytes())

	// Test with an invalid hex string
	invalidHexStr := "invalidhexstring"
	_, err = NewBlsPubKeyFromHex(invalidHexStr)
	require.NotNil(t, err)

	// Test with a hex string of invalid length
	invalidLengthHexStr := hex.EncodeToString(pk_ser[:48])
	_, err = NewBlsPubKeyFromHex(invalidLengthHexStr)
	require.NotNil(t, err)
}