package keeper

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventabi "github.com/hetu-project/hetu/v1/x/event/abi"
	eventtypes "github.com/hetu-project/hetu/v1/x/event/types"
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
	result, err := k.erc20Keeper.CallEVM(
		ctx,
		alphaTokenABI,
		minterAddr,
		alphaTokenAddr,
		false, // 只读调用
		"authorized_minters",
		minterAddr,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if address is authorized minter: %w", err)
	}

	// 解析返回结果
	var isAuthorized bool
	if result != nil && len(result.Ret) > 0 {
		isAuthorized = result.Ret[0] == 1 // 假设返回的是布尔值，1表示true
	}

	return isAuthorized, nil
}

// addModuleAsAuthorizedMinter 将 blockinflation 模块添加为授权铸造者
func (k Keeper) addModuleAsAuthorizedMinter(ctx sdk.Context, netuid uint16, subnet eventtypes.Subnet, alphaTokenAddr common.Address) error {
	// 获取 SubnetManager 合约地址
	subnetManagerAddress, ok := subnet.Params["subnet_manager"]
	if !ok || !common.IsHexAddress(subnetManagerAddress) {
		return fmt.Errorf("invalid or missing subnet manager address in subnet params: %d", netuid)
	}

	// 获取 SubnetManager ABI
	subnetManagerABI, err := getSubnetManagerABI()
	if err != nil {
		return fmt.Errorf("failed to load SubnetManager ABI: %w", err)
	}

	// 调用 SubnetManager 的 addSubnetMinter 函数
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
	_, err = k.erc20Keeper.CallEVM(
		ctx,
		subnetManagerABI,
		common.BytesToAddress(moduleAddress.Bytes()),
		common.HexToAddress(subnetManagerAddress),
		true, // 提交
		"addSubnetMinter",
		uint16(netuid),
		common.BytesToAddress(moduleAddress.Bytes()),
	)
	if err != nil {
		return fmt.Errorf("failed to add module as authorized minter: %w", err)
	}

	k.Logger(ctx).Info("Added blockinflation module as authorized minter",
		"netuid", netuid,
		"module_address", moduleAddress.String(),
		"subnet_manager", subnetManagerAddress,
	)

	return nil
}

// MintAlphaTokens mints alpha tokens to the specified address
// amount is the ERC-20 smallest-unit amount (usually 18 decimals).
func (k Keeper) MintAlphaTokens(ctx sdk.Context, netuid uint16, recipient string, amount *big.Int) error {
	// 1. Get the subnet to find information about it
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// 2. Resolve and validate the subnet's AlphaToken address
	alphaTokenAddress, ok := subnet.Params["alpha_token"]
	alphaTokenAddress = strings.TrimSpace(alphaTokenAddress)
	if !ok || alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
		return fmt.Errorf("invalid or missing alpha token address in subnet params: %d", netuid)
	}

	// 3. Parse and validate the AlphaToken address
	alphaTokenAddr := common.HexToAddress(alphaTokenAddress)
	if alphaTokenAddr == (common.Address{}) {
		return fmt.Errorf("alpha token address cannot be the zero address: %s", alphaTokenAddress)
	}

	// 4. Get the AlphaToken ABI (cached)
	alphaTokenABI, err := getAlphaTokenABI()
	if err != nil {
		return fmt.Errorf("failed to load AlphaToken ABI: %w", err)
	}

	// 5. Validate and convert recipient address
	recipient = strings.TrimSpace(recipient)
	if !common.IsHexAddress(recipient) {
		return fmt.Errorf("invalid recipient address: %s", recipient)
	}
	recipientAddr := common.HexToAddress(recipient)
	if recipientAddr == (common.Address{}) {
		return fmt.Errorf("recipient address cannot be the zero address: %s", recipient)
	}

	// 6. Validate amount
	if amount == nil {
		return fmt.Errorf("amount cannot be nil")
	}
	if amount.Sign() <= 0 {
		return fmt.Errorf("amount must be positive, got: %s", amount.String())
	}

	// 7. 检查 blockinflation 模块是否为授权铸造者
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
	isAuthorized, err := k.checkIfAuthorizedMinter(ctx, alphaTokenABI, alphaTokenAddr, common.BytesToAddress(moduleAddress.Bytes()))
	if err != nil {
		k.Logger(ctx).Error("Failed to check if module is authorized minter",
			"netuid", netuid,
			"error", err,
		)
		// 继续执行，尝试铸造，可能会失败
	}

	// 8. 如果不是授权铸造者，尝试添加
	if !isAuthorized {
		k.Logger(ctx).Info("Module is not an authorized minter, attempting to add",
			"netuid", netuid,
			"module_address", moduleAddress.String(),
		)

		if err := k.addModuleAsAuthorizedMinter(ctx, netuid, subnet, alphaTokenAddr); err != nil {
			k.Logger(ctx).Error("Failed to add module as authorized minter",
				"netuid", netuid,
				"error", err,
			)
			// 继续执行，尝试铸造，可能会失败
		}
	}

	// 9. Call the mint function on the AlphaToken contract
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
