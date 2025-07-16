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
}

// NewParams creates a new Params instance
func NewParams(enableBlockInflation bool, mintDenom string, totalSupply, defaultBlockEmission math.Int) Params {
	return Params{
		EnableBlockInflation: enableBlockInflation,
		MintDenom:            mintDenom,
		TotalSupply:          totalSupply,
		DefaultBlockEmission: defaultBlockEmission,
	}
}

// DefaultParams returns default blockinflation parameters
func DefaultParams() Params {
	return NewParams(
		true,                                // Enable block inflation by default
		"ahetu",                             // Default denom (changed from arao to ahetu)
		math.NewInt(21_000_000_000_000_000), // 21,000,000,000,000,000 aHETU (10^18 precision)
		math.NewInt(1_000_000_000_000_000),  // 1,000,000,000,000,000 aHETU per block (1 HETU per block)
	)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyEnableBlockInflation, &p.EnableBlockInflation, validateEnableBlockInflation),
		paramstypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramstypes.NewParamSetPair(KeyTotalSupply, &p.TotalSupply, validateTotalSupply),
		paramstypes.NewParamSetPair(KeyDefaultBlockEmission, &p.DefaultBlockEmission, validateDefaultBlockEmission),
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
