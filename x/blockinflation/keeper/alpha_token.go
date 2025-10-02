package keeper

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventabi "github.com/hetu-project/hetu/v1/x/event/abi"
)

var (
	alphaABIOnce   sync.Once
	alphaABIVal    abi.ABI
	alphaABIValErr error

	subnetManagerABIOnce   sync.Once
	subnetManagerABIVal    abi.ABI
	subnetManagerABIValErr error

	subnetAMMABIOnce   sync.Once
	subnetAMMABIVal    abi.ABI
	subnetAMMABIValErr error
)

// getAlphaTokenABI returns the cached AlphaToken ABI or parses it once if not cached
func getAlphaTokenABI() (abi.ABI, error) {
	alphaABIOnce.Do(func() {
		alphaABIVal, alphaABIValErr = abi.JSON(bytes.NewReader(eventabi.AlphaTokenABI))
	})
	return alphaABIVal, alphaABIValErr
}

// getSubnetManagerABI returns the cached SubnetManager ABI or parses it once if not cached
func getSubnetManagerABI() (abi.ABI, error) {
	subnetManagerABIOnce.Do(func() {
		subnetManagerABIVal, subnetManagerABIValErr = abi.JSON(bytes.NewReader(eventabi.SubnetManagerABI))
	})
	return subnetManagerABIVal, subnetManagerABIValErr
}

// getSubnetAMMABI returns the cached SubnetAMM ABI or parses it once if not cached
func getSubnetAMMABI() (abi.ABI, error) {
	subnetAMMABIOnce.Do(func() {
		subnetAMMABIVal, subnetAMMABIValErr = abi.JSON(bytes.NewReader(eventabi.SubnetAMMABI))
	})
	return subnetAMMABIVal, subnetAMMABIValErr
}

// checkIfAuthorizedMinter Check if the address is an authorized foundry
func (k Keeper) checkIfAuthorizedMinter(ctx sdk.Context, alphaTokenABI abi.ABI, alphaTokenAddr common.Address, minterAddr common.Address) (bool, error) {
	// Using the contract's own address as the caller is a valid address
	result, err := k.erc20Keeper.CallEVM(
		ctx,
		alphaTokenABI,
		alphaTokenAddr, // Use the contract's own address as the caller to ensure that the address exists
		alphaTokenAddr, // contract address
		false,          // Read only call
		"authorized_minters",
		minterAddr, // The address to be checked as a parameter
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if address is authorized minter: %w", err)
	}

	// Record detailed logs to assist in debugging
	k.Logger(ctx).Debug("authorized_minters check result",
		"token_address", alphaTokenAddr.Hex(),
		"minter_address", minterAddr.Hex(),
		"result_exists", result != nil,
		"ret_length", len(result.Ret),
	)

	// Parse return result - Solidity Boolean value in EVM is a 32 byte value
	var isAuthorized bool
	if result != nil && len(result.Ret) > 0 {
		// Check for any non-zero bytes, indicating true
		for _, b := range result.Ret {
			if b != 0 {
				isAuthorized = true
				break
			}
		}
	}

	k.Logger(ctx).Debug("Authorization check completed",
		"minter", minterAddr.Hex(),
		"is_authorized", isAuthorized,
	)

	return isAuthorized, nil
}

// GetSubnet Manager Address Get Subnet Manager contract address
// Prioritize obtaining from environment variables. If the environment variable does not exist, attempt to read from the configuration file
func getSubnetManagerAddress() string {
	// 1. Get the address from the environment variable first
	envVarName := "SUBNET_MANAGER_CONTRACT_ADDRESS"
	if addr := os.Getenv(envVarName); addr != "" {
		return addr
	}

	// 2. If the environment variable does not exist, attempt to read from the configuration file
	if viper.GetString("subnet_manager_contract_address") != "" {
		return viper.GetString("subnet_manager_contract_address")
	}

	// 3. Try to initialize the configuration and read again
	// Set the configuration file name
	viper.SetConfigName("app")
	// Add the configuration file path
	viper.AddConfigPath(".")

	// Read the configuration file
	if err := viper.ReadInConfig(); err == nil {
		if addr := viper.GetString("subnet_manager_contract_address"); addr != "" {
			return addr
		}
	}

	// 4. Try to read config.toml
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err == nil {
		if addr := viper.GetString("subnet_manager_contract_address"); addr != "" {
			return addr
		}
	}

	// 5. If none are found, return an empty string
	return ""
}

// MintAlphaTokens mints alpha tokens to the specified address
// amount is the ERC-20 smallest-unit amount (usually 18 decimals).
func (k Keeper) MintAlphaTokens(ctx sdk.Context, netuid uint16, recipient string, amount *big.Int) error {
	// 1. Get the subnet to find information about it
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// 2. Get the subnet detailed information to get the AlphaToken address
	subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet info not found: %d", netuid)
	}

	// 3. Verify the AlphaToken address
	alphaTokenAddress := subnetInfo.AlphaToken
	if alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
		// Try to get from subnet.Params (backward compatibility)
		alphaTokenAddress, ok := subnet.Params["alpha_token"]
		alphaTokenAddress = strings.TrimSpace(alphaTokenAddress)
		if !ok || alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
			return fmt.Errorf("invalid or missing alpha token address: %d", netuid)
		}
	}

	// 4. Parse and validate the AlphaToken address
	alphaTokenAddr := common.HexToAddress(alphaTokenAddress)
	if alphaTokenAddr == (common.Address{}) {
		return fmt.Errorf("alpha token address cannot be the zero address: %s", alphaTokenAddress)
	}

	// 5. Get the AlphaToken ABI (cached)
	alphaTokenABI, err := getAlphaTokenABI()
	if err != nil {
		return fmt.Errorf("failed to load AlphaToken ABI: %w", err)
	}

	// 6. Validate and convert recipient address
	recipient = strings.TrimSpace(recipient)
	if !common.IsHexAddress(recipient) {
		return fmt.Errorf("invalid recipient address: %s", recipient)
	}
	recipientAddr := common.HexToAddress(recipient)
	if recipientAddr == (common.Address{}) {
		return fmt.Errorf("recipient address cannot be the zero address: %s", recipient)
	}

	// 7. Validate amount
	if amount == nil {
		return fmt.Errorf("amount cannot be nil")
	}
	if amount.Sign() <= 0 {
		return fmt.Errorf("amount must be positive, got: %s", amount.String())
	}

	// 8. Check if the blockinflation module is an authorized foundry
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
	isAuthorized, err := k.checkIfAuthorizedMinter(ctx, alphaTokenABI, alphaTokenAddr, common.BytesToAddress(moduleAddress.Bytes()))
	if err != nil {
		k.Logger(ctx).Error("Failed to check if module is authorized minter",
			"netuid", netuid,
			"error", err,
		)
		// Continue to execute, try to mint, which may fail
	}

	// 9. If it is not an authorized foundry, return an error
	if !isAuthorized {
		k.Logger(ctx).Error("Module is not an authorized minter",
			"netuid", netuid,
			"module_address", moduleAddress.String(),
		)

		// Get the hexadecimal form of the module address, making it easier for the subnet owner to add
		moduleHexAddress := common.BytesToAddress(moduleAddress.Bytes()).Hex()

		return fmt.Errorf("blockinflation module is not an authorized minter for subnet %d. "+
			"Please ask the subnet owner to call addSubnetMinter(%d, %s) on the SubnetManager contract at %s",
			netuid, netuid, moduleHexAddress, getSubnetManagerAddress())
	}

	// 10. Call the mint function on the AlphaToken contract
	_, err = k.erc20Keeper.CallEVM(
		ctx,
		alphaTokenABI,
		common.BytesToAddress(moduleAddress.Bytes()),
		alphaTokenAddr,
		true, // commit
		"mint",
		recipientAddr,
		amount,
	)
	if err != nil {
		return fmt.Errorf("failed to mint alpha tokens: %w", err)
	}

	k.Logger(ctx).Info("Minted alpha tokens",
		"netuid", netuid,
		"recipient", recipient,
		"amount", amount.String(),
		"token_address", alphaTokenAddr.Hex(),
	)

	return nil
}
