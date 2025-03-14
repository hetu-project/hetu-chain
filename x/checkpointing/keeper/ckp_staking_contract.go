package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/store/prefix"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hetu-project/hetu/v1/contracts"
	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// GetValidatorStake queries the staking contract to get a validator's stake
func (k Keeper) GetValidatorStake(ctx sdk.Context, validatorAddr common.Address) (*big.Int, error) {
	ckptContractABI := contracts.CKPTValStakingContract.ABI

	StakingContractAddress, err := k.GetValidatorContractAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get staking contract address: %w", err)
	}
	res, err := k.evm.CallEVM(ctx, ckptContractABI, types.ModuleAddress, StakingContractAddress, false, "getStake", validatorAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to call staking contract: %w", err)
	}

	unpacked, err := ckptContractABI.Unpack("getStake", res.Ret)
	if err != nil || len(unpacked) == 0 {
		return nil, fmt.Errorf("failed to unpack staking contract response: %w", err)
	}

	stake, ok := unpacked[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid response type from staking contract")
	}

	return stake, nil
}

// GetTopValidators queries the staking contract to get the top N validators by stake
func (k Keeper) GetTopValidators(ctx sdk.Context, count uint64) ([]types.Validator, []string, []string, error) {
	ckptContractABI := contracts.CKPTValStakingContract.ABI

	StakingContractAddress, err := k.GetValidatorContractAddresses(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get staking contract address: %w", err)
	}
	res, err := k.evm.CallEVM(ctx, ckptContractABI, types.ModuleAddress, StakingContractAddress, false, "getTopValidators", big.NewInt(int64(count)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to call staking contract: %w", err)
	}

	// Define a struct to hold the unpacked result
	type TopValidatorsResult struct {
		Addresses      []common.Address
		Stakes         []*big.Int
		DispatcherURLs []string
		BlsPublicKeys  []string
	}

	var result TopValidatorsResult
	if err := ckptContractABI.UnpackIntoInterface(&result, "getTopValidators", res.Ret); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unpack staking contract response: %w", err)
	}

	// Convert to Validator structs
	validators := make([]types.Validator, len(result.Addresses))
	dispatcherURLs := make([]string, len(result.Addresses))
	blsPublicKeys := make([]string, len(result.Addresses))

	for i, addr := range result.Addresses {
		validators[i] = types.Validator{
			Addr:  addr.Bytes(),
			Power: result.Stakes[i].Int64(),
		}
		dispatcherURLs[i] = result.DispatcherURLs[i]
		blsPublicKeys[i] = result.BlsPublicKeys[i]
	}

	return validators, dispatcherURLs, blsPublicKeys, nil
}

// GetValidatorDispatcherURL queries the staking contract to get a validator's dispatcher URL
func (k Keeper) GetValidatorDispatcherURL(ctx sdk.Context, validatorAddr common.Address) (string, error) {
	ckptContractABI := contracts.CKPTValStakingContract.ABI

	StakingContractAddress, err := k.GetValidatorContractAddresses(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get staking contract address: %w", err)
	}

	res, err := k.evm.CallEVM(ctx, ckptContractABI, types.ModuleAddress, StakingContractAddress, false, "getValidator", validatorAddr)
	if err != nil {
		return "", fmt.Errorf("failed to call staking contract: %w", err)
	}

	// Define a struct to hold the unpacked result
	type ValidatorInfo struct {
		StakedAmount   *big.Int
		PendingRewards *big.Int
		UnstakeTime    *big.Int
		DispatcherURL  string
		IsActive       bool
	}

	var result ValidatorInfo
	if err := ckptContractABI.UnpackIntoInterface(&result, "getValidator", res.Ret); err != nil {
		return "", fmt.Errorf("failed to unpack staking contract response: %w", err)
	}

	return result.DispatcherURL, nil
}

// StoreValidatorContractAddresses stores the contract addresses for validators
func (k Keeper) StoreValidatorContractAddresses(ctx sdk.Context, contractAddr common.Address) error {
	store := k.contractAddressStore(ctx)

	key := tmhash.Sum([]byte(types.CkPTContractAddrKey))
	store.Set(key, contractAddr.Bytes())
	return nil
}

// GetValidatorContractAddresses retrieves the contract addresses
func (k Keeper) GetValidatorContractAddresses(ctx sdk.Context) (common.Address, error) {
	store := k.contractAddressStore(ctx)
	key := tmhash.Sum([]byte(types.CkPTContractAddrKey))
	contractAddressesBytes := store.Get(key)
	var contractAddr common.Address
	if contractAddressesBytes == nil {
		return contractAddr, fmt.Errorf("no contract addresses found")
	}
	contractAddr.SetBytes(contractAddressesBytes)

	return contractAddr, nil
}

// contractAddressStore returns the KVStore for validator contract addresses
func (k Keeper) contractAddressStore(ctx sdk.Context) prefix.Store {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(storeAdapter, nil)
}
