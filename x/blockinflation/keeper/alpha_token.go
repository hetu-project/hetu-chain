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

// checkIfAuthorizedMinter 检查地址是否为授权铸造者
func (k Keeper) checkIfAuthorizedMinter(ctx sdk.Context, alphaTokenABI abi.ABI, alphaTokenAddr common.Address, minterAddr common.Address) (bool, error) {
	// 使用合约自身地址作为调用者，这是一个有效的地址
	result, err := k.erc20Keeper.CallEVM(
		ctx,
		alphaTokenABI,
		alphaTokenAddr, // 使用合约自身地址作为调用者，确保地址存在
		alphaTokenAddr, // 合约地址
		false,          // 只读调用
		"authorized_minters",
		minterAddr, // 要检查的地址作为参数
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if address is authorized minter: %w", err)
	}

	// 记录详细日志，帮助调试
	k.Logger(ctx).Debug("authorized_minters check result",
		"token_address", alphaTokenAddr.Hex(),
		"minter_address", minterAddr.Hex(),
		"result_exists", result != nil,
		"ret_length", len(result.Ret),
	)

	// 解析返回结果 - Solidity布尔值在EVM中是32字节的值
	var isAuthorized bool
	if result != nil && len(result.Ret) > 0 {
		// 检查是否有任何非零字节，表示true
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

// getSubnetManagerAddress 获取SubnetManager合约地址
// 优先从环境变量获取，如果环境变量不存在则尝试从配置文件读取
func getSubnetManagerAddress() string {
	// 1. 优先从环境变量获取
	envVarName := "SUBNET_MANAGER_CONTRACT_ADDRESS"
	if addr := os.Getenv(envVarName); addr != "" {
		return addr
	}

	// 2. 如果环境变量不存在，尝试从配置文件读取
	if viper.GetString("subnet_manager_contract_address") != "" {
		return viper.GetString("subnet_manager_contract_address")
	}

	// 3. 尝试初始化配置并再次读取
	// 设置配置文件名称
	viper.SetConfigName("app")
	// 添加配置文件路径
	viper.AddConfigPath(".")

	// 读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		if addr := viper.GetString("subnet_manager_contract_address"); addr != "" {
			return addr
		}
	}

	// 4. 尝试读取config.toml
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err == nil {
		if addr := viper.GetString("subnet_manager_contract_address"); addr != "" {
			return addr
		}
	}

	// 5. 如果都找不到，返回空字符串
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

	// 2. 获取子网详细信息以获取AlphaToken地址
	subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet info not found: %d", netuid)
	}

	// 3. 验证Alpha代币地址
	alphaTokenAddress := subnetInfo.AlphaToken
	if alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
		// 尝试从subnet.Params中获取（向后兼容）
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

	// 8. 检查 blockinflation 模块是否为授权铸造者
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
	isAuthorized, err := k.checkIfAuthorizedMinter(ctx, alphaTokenABI, alphaTokenAddr, common.BytesToAddress(moduleAddress.Bytes()))
	if err != nil {
		k.Logger(ctx).Error("Failed to check if module is authorized minter",
			"netuid", netuid,
			"error", err,
		)
		// 继续执行，尝试铸造，可能会失败
	}

	// 9. 如果不是授权铸造者，返回错误
	if !isAuthorized {
		k.Logger(ctx).Error("Module is not an authorized minter",
			"netuid", netuid,
			"module_address", moduleAddress.String(),
		)

		// 获取模块地址的十六进制形式，方便子网所有者添加
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
