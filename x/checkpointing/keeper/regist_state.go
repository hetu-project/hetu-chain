package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/crypto/bls12381"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

type RegistrationState struct {
	cdc codec.BinaryCodec
	// addrToBlsKeys maps validator addresses to BLS public keys
	addrToBlsKeys storetypes.KVStore
	// blsKeysToAddr maps BLS public keys to validator addresses
	blsKeysToAddr storetypes.KVStore
}

func (k Keeper) RegistrationState(ctx context.Context) RegistrationState {
	// Build the RegistrationState storage
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return RegistrationState{
		cdc:           k.cdc,
		addrToBlsKeys: prefix.NewStore(storeAdapter, types.AddrToBlsKeyPrefix),
		blsKeysToAddr: prefix.NewStore(storeAdapter, types.BlsKeyToAddrPrefix),
	}
}

// CreateRegistration inserts the BLS key into the addr -> key and key -> addr storage
func (rs RegistrationState) CreateRegistration(key bls12381.PublicKey, valAddr common.Address) error {
	blsPubKey, err := rs.GetBlsPubKey(valAddr)

	// we should disallow a validator to register with different BLS public keys
	if err == nil && !blsPubKey.Equal(key) {
		return types.ErrBlsKeyAlreadyExist.Wrapf("the validator has registered a BLS public key")
	}

	// we should disallow the same BLS public key is registered by different validators
	bkToAddrKey := types.BlsKeyToAddrKey(key)
	rawAddr := rs.blsKeysToAddr.Get(bkToAddrKey)
	addr := common.BytesToAddress(rawAddr)
	if rawAddr != nil && !(addr == valAddr) {
		return types.ErrBlsKeyAlreadyExist.Wrapf("same BLS public key is registered by another validator")
	}

	// save concrete BLS public key object
	blsPkKey := valAddr.Bytes()
	rs.addrToBlsKeys.Set(blsPkKey, key)
	rs.blsKeysToAddr.Set(bkToAddrKey, valAddr.Bytes())

	return nil
}

// GetBlsPubKey retrieves BLS public key by validator's address
func (rs RegistrationState) GetBlsPubKey(addr common.Address) (bls12381.PublicKey, error) {
	pkKey := addr.Bytes()
	rawBytes := rs.addrToBlsKeys.Get(pkKey)
	if rawBytes == nil {
		return nil, types.ErrBlsKeyDoesNotExist.Wrapf("BLS public key does not exist with address %s", addr)
	}
	pk := new(bls12381.PublicKey)
	err := pk.Unmarshal(rawBytes)

	return *pk, err
}

// GetValAddr returns the validator address of the BLS public key
func (rs RegistrationState) GetValAddr(key bls12381.PublicKey) (common.Address, error) {
	pkKey := types.BlsKeyToAddrKey(key)
	rawBytes := rs.blsKeysToAddr.Get(pkKey)
	if rawBytes == nil {
		return common.BytesToAddress(nil), types.ErrValAddrDoesNotExist.Wrapf("validator address does not exist with BLS public key %s", key)
	}
	return common.BytesToAddress(rawBytes), nil
}

// Exists checks whether a BLS key exists
func (rs RegistrationState) Exists(addr common.Address) bool {
	pkKey := addr.Bytes()
	return rs.addrToBlsKeys.Has(pkKey)
}
