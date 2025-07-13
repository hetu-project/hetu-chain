package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
}

// StakingKeeper defines the expected staking keeper
type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator
	GetLastValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool))
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (delegations []stakingtypes.Delegation)
	GetDelegatorUnbondingDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (unbondingDelegations []stakingtypes.UnbondingDelegation)
	GetDelegatorRedelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (redelegations []stakingtypes.Redelegation)
	GetDelegatorValidators(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (validators []stakingtypes.Validator)
	GetDelegatorValidator(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetUnbondingDelegations(ctx sdk.Context, valAddr sdk.ValAddress, maxRetrieve uint16) (unbondingDelegations []stakingtypes.UnbondingDelegation)
	GetRedelegations(ctx sdk.Context, delAddr sdk.AccAddress, maxRetrieve uint16) (redelegations []stakingtypes.Redelegation)
	GetRedelegationsFromSrcValidator(ctx sdk.Context, valAddr sdk.ValAddress, maxRetrieve uint16) (redelegations []stakingtypes.Redelegation)
	GetUnbondingDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (unbondingDelegation stakingtypes.UnbondingDelegation, found bool)
	GetRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddrSrc, valAddrDst sdk.ValAddress) (redelegation stakingtypes.Redelegation, found bool)
	GetRedelegationsFromSrcValidator(ctx sdk.Context, valAddr sdk.ValAddress, maxRetrieve uint16) (redelegations []stakingtypes.Redelegation)
	GetUnbondingDelegationsFromValidator(ctx sdk.Context, valAddr sdk.ValAddress, maxRetrieve uint16) (unbondingDelegations []stakingtypes.UnbondingDelegation)
	GetAllSDKDelegations(ctx sdk.Context) []stakingtypes.Delegation
}

// YumaKeeper defines the expected yuma keeper
type YumaKeeper interface {
	GetSubnetCount(ctx sdk.Context) int64
}
