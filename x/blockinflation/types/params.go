package types

import (
	"fmt"

	"cosmossdk.io/math"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter keys
var (
	KeyEnableBlockInflation = []byte("EnableBlockInflation")
	KeyMintDenom            = []byte("MintDenom")
	KeyTotalSupply          = []byte("TotalSupply")
	KeyDefaultBlockEmission = []byte("DefaultBlockEmission")
	KeySubnetRewardBase     = []byte("SubnetRewardBase")
	KeySubnetRewardK        = []byte("SubnetRewardK")
	KeySubnetRewardMaxRatio = []byte("SubnetRewardMaxRatio")
	KeySubnetMovingAlpha    = []byte("SubnetMovingAlpha")
	KeySubnetOwnerCut       = []byte("SubnetOwnerCut")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the blockinflation module.
type Params struct {
	// EnableBlockInflation enables or disables block inflation
	EnableBlockInflation bool `json:"enable_block_inflation" yaml:"enable_block_inflation"`
	// MintDenom defines the denomination of the minted token
	MintDenom string `json:"mint_denom" yaml:"mint_denom"`
	// TotalSupply defines the maximum total supply (21,000,000,000,000,000 RAO)
	TotalSupply math.Int `json:"total_supply" yaml:"total_supply"`
	// DefaultBlockEmission defines the default block emission (1,000,000,000 RAO)
	DefaultBlockEmission math.Int `json:"default_block_emission" yaml:"default_block_emission"`
	// SubnetRewardBase defines the base subnet reward ratio (e.g., 0.10)
	SubnetRewardBase math.LegacyDec `json:"subnet_reward_base" yaml:"subnet_reward_base"`
	// SubnetRewardK defines the growth rate coefficient (e.g., 0.1)
	SubnetRewardK math.LegacyDec `json:"subnet_reward_k" yaml:"subnet_reward_k"`
	// SubnetRewardMaxRatio defines the maximum subnet reward ratio (e.g., 0.5)
	SubnetRewardMaxRatio math.LegacyDec `json:"subnet_reward_max_ratio" yaml:"subnet_reward_max_ratio"`
	// SubnetMovingAlpha defines the moving average coefficient for subnet reward
	SubnetMovingAlpha math.LegacyDec `json:"subnet_moving_alpha" yaml:"subnet_moving_alpha"`
	// SubnetOwnerCut defines the percentage of alpha_out that goes to subnet owners (e.g., 0.18 = 18%)
	SubnetOwnerCut math.LegacyDec `json:"subnet_owner_cut" yaml:"subnet_owner_cut"`
}

// NewParams creates a new Params instance
func NewParams(enableBlockInflation bool, mintDenom string, totalSupply, defaultBlockEmission math.Int, subnetRewardBase, subnetRewardK, subnetRewardMaxRatio, subnetMovingAlpha, subnetOwnerCut math.LegacyDec) Params {
	return Params{
		EnableBlockInflation: enableBlockInflation,
		MintDenom:            mintDenom,
		TotalSupply:          totalSupply,
		DefaultBlockEmission: defaultBlockEmission,
		SubnetRewardBase:     subnetRewardBase,
		SubnetRewardK:        subnetRewardK,
		SubnetRewardMaxRatio: subnetRewardMaxRatio,
		SubnetMovingAlpha:    subnetMovingAlpha,
		SubnetOwnerCut:       subnetOwnerCut,
	}
}

// DefaultParams returns default blockinflation parameters
func DefaultParams() Params {
	return NewParams(
		true,                                // Enable block inflation by default
		"ahetu",                             // Default denom (changed from arao to ahetu)
		math.NewInt(21_000_000_000_000_000), // 21,000,000,000,000,000 aHETU (10^18 precision)
		math.NewInt(1_000_000_000_000_000),  // 1,000,000,000,000,000 aHETU per block (1 HETU per block)
		math.LegacyNewDecWithPrec(10, 2),    // Default SubnetRewardBase (0.10)
		math.LegacyNewDecWithPrec(10, 2),    // Default SubnetRewardK (0.10)
		math.LegacyNewDecWithPrec(50, 2),    // Default SubnetRewardMaxRatio (0.50)
		math.LegacyNewDecWithPrec(3, 6),     // Default SubnetMovingAlpha (0.000003)
		math.LegacyNewDecWithPrec(18, 2),    // Default SubnetOwnerCut (0.18)
	)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyEnableBlockInflation, &p.EnableBlockInflation, validateEnableBlockInflation),
		paramstypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramstypes.NewParamSetPair(KeyTotalSupply, &p.TotalSupply, validateTotalSupply),
		paramstypes.NewParamSetPair(KeyDefaultBlockEmission, &p.DefaultBlockEmission, validateDefaultBlockEmission),
		paramstypes.NewParamSetPair(KeySubnetRewardBase, &p.SubnetRewardBase, validateSubnetRewardBase),
		paramstypes.NewParamSetPair(KeySubnetRewardK, &p.SubnetRewardK, validateSubnetRewardK),
		paramstypes.NewParamSetPair(KeySubnetRewardMaxRatio, &p.SubnetRewardMaxRatio, validateSubnetRewardMaxRatio),
		paramstypes.NewParamSetPair(KeySubnetMovingAlpha, &p.SubnetMovingAlpha, validateSubnetMovingAlpha),
		paramstypes.NewParamSetPair(KeySubnetOwnerCut, &p.SubnetOwnerCut, validateSubnetOwnerCut),
	}
}

// Validate performs basic validation on blockinflation parameters.
func (p Params) Validate() error {
	if err := validateEnableBlockInflation(p.EnableBlockInflation); err != nil {
		return err
	}
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateTotalSupply(p.TotalSupply); err != nil {
		return err
	}
	if err := validateDefaultBlockEmission(p.DefaultBlockEmission); err != nil {
		return err
	}
	if err := validateSubnetRewardBase(p.SubnetRewardBase); err != nil {
		return err
	}
	if err := validateSubnetRewardK(p.SubnetRewardK); err != nil {
		return err
	}
	if err := validateSubnetRewardMaxRatio(p.SubnetRewardMaxRatio); err != nil {
		return err
	}
	if err := validateSubnetMovingAlpha(p.SubnetMovingAlpha); err != nil {
		return err
	}
	if err := validateSubnetOwnerCut(p.SubnetOwnerCut); err != nil {
		return err
	}
	return nil
}

func validateEnableBlockInflation(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("mint denom cannot be empty")
	}
	return nil
}

func validateTotalSupply(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsPositive() {
		return fmt.Errorf("total supply must be positive")
	}
	return nil
}

func validateDefaultBlockEmission(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsPositive() {
		return fmt.Errorf("default block emission must be positive")
	}
	return nil
}

func validateSubnetRewardBase(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("subnet reward base cannot be negative")
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("subnet reward base cannot be greater than 1")
	}
	return nil
}

func validateSubnetRewardK(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("subnet reward k cannot be negative")
	}
	return nil
}

func validateSubnetRewardMaxRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("subnet reward max ratio cannot be negative")
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("subnet reward max ratio cannot be greater than 1")
	}
	return nil
}

func validateSubnetMovingAlpha(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("subnet moving alpha cannot be negative")
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("subnet moving alpha cannot be greater than 1")
	}
	return nil
}

func validateSubnetOwnerCut(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("subnet owner cut cannot be negative")
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("subnet owner cut cannot be greater than 1")
	}
	return nil
}
