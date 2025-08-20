package keeper

import (
	"bytes"
	"fmt"
	"math/big"
	"os"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/viper"

	blockinflationtypes "github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventabi "github.com/hetu-project/hetu/v1/x/event/abi"
)

// SyncAMMPoolState 同步链上状态和合约 AMM 池状态
func (k Keeper) SyncAMMPoolState(ctx sdk.Context, netuid uint16) error {
	k.Logger(ctx).Debug("Starting to sync AMM pool state", "netuid", netuid)

	// 1. 获取子网信息，包括 AMM 池地址
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		k.Logger(ctx).Error("Failed to find subnet", "netuid", netuid)
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	k.Logger(ctx).Debug("Retrieved subnet information",
		"netuid", netuid,
		"owner", subnet.Owner,
		"locked_amount", subnet.LockedAmount,
		"burned_amount", subnet.BurnedAmount,
		"amm_pool", subnet.AmmPool,
		"mechanism", subnet.Mechanism,
		"ema_price_halving_blocks", subnet.EMAPriceHalvingBlocks,
		"first_emission_block", subnet.FirstEmissionBlock)

	// 打印所有参数，帮助调试
	for key, value := range subnet.Params {
		k.Logger(ctx).Debug("Subnet parameter",
			"netuid", netuid,
			"key", key,
			"value", value)
	}

	// 2. 验证 AMM 池地址
	ammPoolAddress := subnet.AmmPool

	if ammPoolAddress == "" {
		k.Logger(ctx).Error("AMM pool address is empty", "netuid", netuid)
		return fmt.Errorf("invalid or missing AMM pool address in subnet params: %d", netuid)
	}

	if !common.IsHexAddress(ammPoolAddress) {
		k.Logger(ctx).Error("Invalid AMM pool address format",
			"netuid", netuid,
			"amm_pool_address", ammPoolAddress)
		return fmt.Errorf("invalid or missing AMM pool address in subnet params: %d", netuid)
	}

	ammPoolAddr := common.HexToAddress(ammPoolAddress)
	k.Logger(ctx).Debug("Validated AMM pool address",
		"netuid", netuid,
		"amm_pool_address", ammPoolAddr.Hex())

	// 3. 获取 SubnetAMM 合约的 ABI
	ammABI, err := getSubnetAMMABI()
	if err != nil {
		k.Logger(ctx).Error("Failed to load SubnetAMM ABI", "error", err)
		return fmt.Errorf("failed to load SubnetAMM ABI: %w", err)
	}
	k.Logger(ctx).Debug("Loaded SubnetAMM ABI")

	// 4. 从合约中获取当前池状态
	moduleAddress := authtypes.NewModuleAddress(blockinflationtypes.ModuleName)
	k.Logger(ctx).Debug("Calling EVM to get pool info",
		"module_address", moduleAddress.String(),
		"amm_pool_address", ammPoolAddr.Hex())

	result, err := k.erc20Keeper.CallEVM(
		ctx,
		ammABI,
		common.BytesToAddress(moduleAddress.Bytes()),
		ammPoolAddr,
		false, // 只读调用
		"getPoolInfo",
	)
	if err != nil {
		k.Logger(ctx).Error("Failed to call AMM pool contract",
			"netuid", netuid,
			"amm_pool_address", ammPoolAddr.Hex(),
			"error", err)
		return fmt.Errorf("failed to get AMM pool info: %w", err)
	}

	// 5. 解析返回结果
	if result == nil {
		k.Logger(ctx).Error("Null result from getPoolInfo call", "netuid", netuid)
		return fmt.Errorf("invalid result from getPoolInfo call")
	}

	if len(result.Ret) < 32*8 {
		k.Logger(ctx).Error("Insufficient data in result",
			"netuid", netuid,
			"result_length", len(result.Ret),
			"expected_min_length", 32*8)
		return fmt.Errorf("invalid result from getPoolInfo call")
	}

	k.Logger(ctx).Debug("Got valid result from getPoolInfo call",
		"netuid", netuid,
		"result_length", len(result.Ret))

	// 从返回的字节数组中提取数据
	// 每个uint256值占32字节
	// 注意：这里的解析方式取决于合约返回值的具体编码方式
	// 这里假设返回值是按照顺序排列的8个uint256值
	mechanismType := new(big.Int).SetBytes(result.Ret[0:32])    // 第1个uint256 - 机制类型
	subnetHetu := new(big.Int).SetBytes(result.Ret[32:64])      // 第2个uint256 - TAO数量
	subnetAlphaIn := new(big.Int).SetBytes(result.Ret[64:96])   // 第3个uint256 - AlphaIn数量
	subnetAlphaOut := new(big.Int).SetBytes(result.Ret[96:128]) // 第4个uint256 - AlphaOut数量
	currentPrice := new(big.Int).SetBytes(result.Ret[128:160])  // 第5个uint256 - 当前价格
	movingPrice := new(big.Int).SetBytes(result.Ret[160:192])   // 第6个uint256 - 移动平均价格
	totalVolume := new(big.Int).SetBytes(result.Ret[192:224])   // 第7个uint256 - 总交易量

	k.Logger(ctx).Debug("Parsed values from result",
		"netuid", netuid,
		"mechanism_type", mechanismType.String(),
		"subnet_hetu", subnetHetu.String(),
		"subnet_alpha_in", subnetAlphaIn.String(),
		"subnet_alpha_out", subnetAlphaOut.String(),
		"current_price", currentPrice.String(),
		"moving_price", movingPrice.String(),
		"total_volume", totalVolume.String())

	// 6. 获取链上状态
	currentTaoIn := k.eventKeeper.GetSubnetTAO(ctx, netuid)
	currentAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
	currentAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)

	k.Logger(ctx).Debug("Current chain state",
		"netuid", netuid,
		"current_tao_in", currentTaoIn.String(),
		"current_alpha_in", currentAlphaIn.String(),
		"current_alpha_out", currentAlphaOut.String())

	// 7. 检查链上状态与合约状态是否一致
	contractTaoIn := math.NewIntFromBigInt(subnetHetu)
	contractAlphaIn := math.NewIntFromBigInt(subnetAlphaIn)
	contractAlphaOut := math.NewIntFromBigInt(subnetAlphaOut)

	// 检查是否需要更新链上状态
	needUpdateChain := !currentTaoIn.Equal(contractTaoIn) ||
		!currentAlphaIn.Equal(contractAlphaIn) ||
		!currentAlphaOut.Equal(contractAlphaOut)

	// 8. 如果不一致，更新链上状态以匹配合约状态
	if needUpdateChain {
		k.Logger(ctx).Info("Updating chain state to match contract",
			"netuid", netuid,
			"old_tao_in", currentTaoIn.String(),
			"new_tao_in", contractTaoIn.String(),
			"old_alpha_in", currentAlphaIn.String(),
			"new_alpha_in", contractAlphaIn.String(),
			"old_alpha_out", currentAlphaOut.String(),
			"new_alpha_out", contractAlphaOut.String())

		// 更新链上状态
		k.eventKeeper.SetSubnetTaoIn(ctx, netuid, contractTaoIn)
		k.eventKeeper.SetSubnetAlphaIn(ctx, netuid, contractAlphaIn)
		k.eventKeeper.SetSubnetAlphaOut(ctx, netuid, contractAlphaOut)

		// 记录日志
		k.Logger(ctx).Info("Synced chain state with contract",
			"netuid", netuid,
			"tao_in", contractTaoIn.String(),
			"alpha_in", contractAlphaIn.String(),
			"alpha_out", contractAlphaOut.String())
	} else {
		k.Logger(ctx).Debug("Chain state already matches contract state", "netuid", netuid)
	}

	// 9. 更新移动平均价格
	// 获取当前移动平均价格
	currentMovingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)
	contractMovingPriceDec := math.LegacyNewDecFromBigInt(movingPrice)

	// 如果移动平均价格不一致，更新链上状态
	if !currentMovingPrice.Equal(contractMovingPriceDec) {
		k.Logger(ctx).Info("Updating moving price",
			"netuid", netuid,
			"old_moving_price", currentMovingPrice.String(),
			"new_moving_price", contractMovingPriceDec.String())

		// 使用UpdateMovingPrice方法更新移动平均价格
		// 获取子网的EMAPriceHalvingBlocks
		halvingBlocks := subnet.EMAPriceHalvingBlocks
		// 使用合约的移动平均价格作为基础，调用UpdateMovingPrice进行更新
		params := k.GetParams(ctx)
		k.eventKeeper.UpdateMovingPrice(ctx, netuid, params.SubnetMovingAlpha, halvingBlocks)
	}

	return nil
}

// SyncAllAMMPools 同步所有活跃子网的 AMM 池状态
func (k Keeper) SyncAllAMMPools(ctx sdk.Context) {
	k.Logger(ctx).Debug("Starting to sync all AMM pools")

	allSubnets := k.eventKeeper.GetAllSubnetNetuids(ctx)
	k.Logger(ctx).Debug("Retrieved all subnet netuids",
		"count", len(allSubnets),
		"netuids", fmt.Sprintf("%v", allSubnets))

	for _, netuid := range allSubnets {
		k.Logger(ctx).Debug("Syncing AMM pool for subnet", "netuid", netuid)

		if err := k.SyncAMMPoolState(ctx, netuid); err != nil {
			k.Logger(ctx).Error("Failed to sync AMM pool state",
				"netuid", netuid,
				"error", err,
			)
		} else {
			k.Logger(ctx).Debug("Successfully synced AMM pool for subnet", "netuid", netuid)
		}
	}

	k.Logger(ctx).Debug("Finished syncing all AMM pools", "count", len(allSubnets))
}

// HandleSubnetRegisteredEvent 处理子网注册事件，立即同步AMM池状态
// 这个函数应该在监听到子网注册事件时被调用
func (k Keeper) HandleSubnetRegisteredEvent(ctx sdk.Context, netuid uint16, ammPoolAddress string) {
	k.Logger(ctx).Info("Handling subnet registered event",
		"netuid", netuid,
		"amm_pool_address", ammPoolAddress)

	// 立即同步AMM池状态
	if err := k.SyncAMMPoolState(ctx, netuid); err != nil {
		k.Logger(ctx).Error("Failed to sync AMM pool state after subnet registration",
			"netuid", netuid,
			"error", err,
		)
	} else {
		k.Logger(ctx).Info("Successfully synced AMM pool after subnet registration",
			"netuid", netuid)
	}
}

// SyncChainStateToContract 将链上状态同步到合约
// 在计算奖励后调用，将TaoIn和AlphaIn注入到AMM合约中
func (k Keeper) SyncChainStateToContract(ctx sdk.Context, netuid uint16, reward blockinflationtypes.SubnetRewards) error {
	// 获取子网信息
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// 验证AMM池地址
	ammPoolAddress := subnet.AmmPool
	if ammPoolAddress == "" || !common.IsHexAddress(ammPoolAddress) {
		return fmt.Errorf("invalid AMM pool address: %s", ammPoolAddress)
	}

	// 获取ABI
	ammABI, err := getSubnetAMMABI()
	if err != nil {
		return fmt.Errorf("failed to load SubnetAMM ABI: %w", err)
	}

	// 准备调用参数
	moduleAddress := authtypes.NewModuleAddress(blockinflationtypes.ModuleName)
	ammPoolAddr := common.HexToAddress(ammPoolAddress)

	// 只有当TaoIn为正时才注入流动性
	if reward.TaoIn.IsPositive() {
		// 计算需要注入的AlphaIn数量
		alphaInAmount := reward.AlphaIn

		k.Logger(ctx).Debug("Preparing to inject liquidity to contract",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", alphaInAmount.String(),
			"amm_pool_address", ammPoolAddr.Hex())

		// 从子网信息中获取WHETU代币地址
		// 这需要在配置中设置或者从环境变量中获取
		whetuAddress := getWHETUAddress()
		if whetuAddress == "" || !common.IsHexAddress(whetuAddress) {
			return fmt.Errorf("invalid or missing WHETU token address")
		}
		whetuAddr := common.HexToAddress(whetuAddress)

		// 1. 铸造Cosmos原生HETU代币
		params := k.GetParams(ctx)
		err = k.bankKeeper.MintCoins(
			ctx,
			blockinflationtypes.ModuleName,
			sdk.NewCoins(sdk.NewCoin(params.MintDenom, reward.TaoIn)),
		)
		if err != nil {
			return fmt.Errorf("failed to mint HETU tokens: %w", err)
		}

		// 2. 将原生HETU转换为WHETU
		// 使用WHETU合约的deposit函数，需要发送原生HETU
		whetuABI, err := abi.JSON(bytes.NewReader(eventabi.WHETUABI))
		if err != nil {
			return fmt.Errorf("failed to parse WHETU ABI: %w", err)
		}

		// 创建deposit函数的调用数据
		depositData, err := whetuABI.Pack("deposit")
		if err != nil {
			return fmt.Errorf("failed to pack deposit function data: %w", err)
		}

		// 使用evmKeeper直接发送交易
		evmModuleAddr := common.BytesToAddress(moduleAddress.Bytes())

		// 获取模块账户的nonce
		nonce := k.evmKeeper.GetNonce(ctx, evmModuleAddr)

		// 创建消息而不是交易
		msg := ethtypes.NewMessage(
			evmModuleAddr,
			&whetuAddr,
			nonce,
			reward.TaoIn.BigInt(), // value
			1000000,               // gasLimit
			big.NewInt(0),         // gasPrice
			big.NewInt(0),         // gasFeeCap
			big.NewInt(0),         // gasTipCap
			depositData,           // data
			nil,                   // accessList
			false,                 // checkNonce
		)

		// 直接调用ApplyMessage而不是EthereumTx
		res, err := k.evmKeeper.ApplyMessage(ctx, msg, nil, true)
		if err != nil {
			k.Logger(ctx).Error("Failed to deposit HETU to WHETU contract",
				"error", err,
				"module_address", moduleAddress.String(),
				"whetu_address", whetuAddr.Hex(),
				"amount", reward.TaoIn.String(),
			)
			return fmt.Errorf("failed to deposit HETU to WHETU contract: %w", err)
		}

		if res.Failed() {
			k.Logger(ctx).Error("WHETU deposit transaction failed",
				"vm_error", res.VmError,
				"module_address", moduleAddress.String(),
				"whetu_address", whetuAddr.Hex(),
				"amount", reward.TaoIn.String(),
			)
			return fmt.Errorf("WHETU deposit transaction failed: %s", res.VmError)
		}

		k.Logger(ctx).Info("Successfully deposited HETU to WHETU contract",
			"amount", reward.TaoIn.String(),
			"module_address", moduleAddress.String(),
			"tx_hash", res.Hash,
		)

		// 3. 铸造Alpha代币并转给模块账户
		if alphaInAmount.IsPositive() {
			// 使用MintAlphaTokens函数，它会处理授权检查和铸造
			// 铸造给模块账户，使用EVM格式的地址
			evmAddress := common.BytesToAddress(moduleAddress.Bytes()).Hex()
			err = k.MintAlphaTokens(ctx, netuid, evmAddress, alphaInAmount.BigInt())
			if err != nil {
				return fmt.Errorf("failed to mint Alpha tokens: %w", err)
			}

			k.Logger(ctx).Info("Successfully minted Alpha tokens for AMM pool",
				"netuid", netuid,
				"alpha_amount", alphaInAmount.String())
		}

		// 获取子网详细信息以获取AlphaToken地址
		subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
		if !found {
			return fmt.Errorf("subnet info not found: %d", netuid)
		}

		// 验证Alpha代币地址
		alphaTokenAddress := subnetInfo.AlphaToken
		if alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
			return fmt.Errorf("invalid or missing alpha token address: %d", netuid)
		}

		k.Logger(ctx).Debug("Retrieved token addresses",
			"netuid", netuid,
			"whetu_token", whetuAddr.Hex(),
			"alpha_token", alphaTokenAddress)

		// 检查模块账户的WHETU代币余额
		erc20ABI, err := abi.JSON(strings.NewReader(erc20AbiJSON))
		if err != nil {
			return fmt.Errorf("failed to parse ERC20 ABI: %w", err)
		}

		// 检查WHETU余额
		whetuBalanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			false, // 只读调用
			"balanceOf",
			common.BytesToAddress(moduleAddress.Bytes()),
		)
		if err != nil {
			return fmt.Errorf("failed to check WHETU balance: %w", err)
		}

		var whetuBalance *big.Int
		if whetuBalanceResult != nil && len(whetuBalanceResult.Ret) >= 32 {
			whetuBalance = new(big.Int).SetBytes(whetuBalanceResult.Ret)
		} else {
			whetuBalance = big.NewInt(0)
		}

		// 检查Alpha余额
		alphaBalanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			common.HexToAddress(alphaTokenAddress),
			false, // 只读调用
			"balanceOf",
			common.BytesToAddress(moduleAddress.Bytes()),
		)
		if err != nil {
			return fmt.Errorf("failed to check Alpha balance: %w", err)
		}

		var alphaBalance *big.Int
		if alphaBalanceResult != nil && len(alphaBalanceResult.Ret) >= 32 {
			alphaBalance = new(big.Int).SetBytes(alphaBalanceResult.Ret)
		} else {
			alphaBalance = big.NewInt(0)
		}

		k.Logger(ctx).Info("Module account balances before approval",
			"module_address", moduleAddress.String(),
			"whetu_balance", whetuBalance.String(),
			"alpha_balance", alphaBalance.String(),
			"required_whetu", reward.TaoIn.BigInt().String(),
			"required_alpha", alphaInAmount.BigInt().String())

		// 检查余额是否足够
		if whetuBalance.Cmp(reward.TaoIn.BigInt()) < 0 {
			return fmt.Errorf("insufficient WHETU balance: have %s, need %s", whetuBalance.String(), reward.TaoIn.BigInt().String())
		}

		if alphaBalance.Cmp(alphaInAmount.BigInt()) < 0 {
			return fmt.Errorf("insufficient Alpha balance: have %s, need %s", alphaBalance.String(), alphaInAmount.BigInt().String())
		}

		// 4. 批准AMM合约使用WHETU代币
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			true, // 提交交易
			"approve",
			ammPoolAddr,
			reward.TaoIn.BigInt(),
		)
		if err != nil {
			return fmt.Errorf("failed to approve WHETU tokens: %w", err)
		}

		// 5. 批准AMM合约使用Alpha代币
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			common.HexToAddress(alphaTokenAddress),
			true, // 提交交易
			"approve",
			ammPoolAddr,
			alphaInAmount.BigInt(),
		)
		if err != nil {
			return fmt.Errorf("failed to approve Alpha tokens: %w", err)
		}

		// 检查批准后的授权额度
		whetuAllowanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			false, // 只读调用
			"allowance",
			common.BytesToAddress(moduleAddress.Bytes()),
			ammPoolAddr,
		)
		if err != nil {
			return fmt.Errorf("failed to check WHETU allowance: %w", err)
		}

		var whetuAllowance *big.Int
		if whetuAllowanceResult != nil && len(whetuAllowanceResult.Ret) >= 32 {
			whetuAllowance = new(big.Int).SetBytes(whetuAllowanceResult.Ret)
		} else {
			whetuAllowance = big.NewInt(0)
		}

		alphaAllowanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			common.HexToAddress(alphaTokenAddress),
			false, // 只读调用
			"allowance",
			common.BytesToAddress(moduleAddress.Bytes()),
			ammPoolAddr,
		)
		if err != nil {
			return fmt.Errorf("failed to check Alpha allowance: %w", err)
		}

		var alphaAllowance *big.Int
		if alphaAllowanceResult != nil && len(alphaAllowanceResult.Ret) >= 32 {
			alphaAllowance = new(big.Int).SetBytes(alphaAllowanceResult.Ret)
		} else {
			alphaAllowance = big.NewInt(0)
		}

		k.Logger(ctx).Info("Token allowances after approval",
			"whetu_allowance", whetuAllowance.String(),
			"alpha_allowance", alphaAllowance.String(),
			"required_whetu", reward.TaoIn.BigInt().String(),
			"required_alpha", alphaInAmount.BigInt().String())

		// 检查授权额度是否足够
		if whetuAllowance.Cmp(reward.TaoIn.BigInt()) < 0 {
			return fmt.Errorf("insufficient WHETU allowance: have %s, need %s", whetuAllowance.String(), reward.TaoIn.BigInt().String())
		}

		if alphaAllowance.Cmp(alphaInAmount.BigInt()) < 0 {
			return fmt.Errorf("insufficient Alpha allowance: have %s, need %s", alphaAllowance.String(), alphaInAmount.BigInt().String())
		}

		k.Logger(ctx).Debug("Successfully approved tokens for AMM contract",
			"netuid", netuid,
			"tao_amount", reward.TaoIn.String(),
			"alpha_amount", alphaInAmount.String())

		// 6. 调用injectLiquidity方法，注入WHETU和Alpha代币
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			ammABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			ammPoolAddr,
			true, // 提交交易
			"injectLiquidity",
			reward.TaoIn.BigInt(),
			alphaInAmount.BigInt(),
		)

		if err != nil {
			// 尝试获取AMM合约状态，可能有助于诊断问题
			poolInfoResult, poolInfoErr := k.erc20Keeper.CallEVM(
				ctx,
				ammABI,
				common.BytesToAddress(moduleAddress.Bytes()),
				ammPoolAddr,
				false, // 只读调用
				"getPoolInfo",
			)

			if poolInfoErr == nil && poolInfoResult != nil && len(poolInfoResult.Ret) >= 32*7 {
				mechanismType := new(big.Int).SetBytes(poolInfoResult.Ret[0:32])
				subnetHetu := new(big.Int).SetBytes(poolInfoResult.Ret[32:64])
				subnetAlphaIn := new(big.Int).SetBytes(poolInfoResult.Ret[64:96])
				currentPrice := new(big.Int).SetBytes(poolInfoResult.Ret[128:160])

				k.Logger(ctx).Error("injectLiquidity failed, current AMM pool state",
					"mechanism_type", mechanismType.String(),
					"subnet_hetu", subnetHetu.String(),
					"subnet_alpha_in", subnetAlphaIn.String(),
					"current_price", currentPrice.String(),
					"error", err,
				)
			}

			return fmt.Errorf("failed to inject liquidity to contract: %w", err)
		}

		k.Logger(ctx).Info("Successfully injected liquidity to contract",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", alphaInAmount.String(),
			"amm_pool_address", ammPoolAddr.Hex())
	}

	return nil
}

// getWHETUAddress 获取WHETU代币地址
// 优先从环境变量获取，如果环境变量不存在则尝试从配置文件读取
func getWHETUAddress() string {
	// 1. 优先从环境变量获取
	envVarName := "WHETU_CONTRACT_ADDRESS"
	if addr := os.Getenv(envVarName); addr != "" {
		return addr
	}

	// 2. 如果环境变量不存在，尝试从配置文件读取
	if viper.GetString("whetu_contract_address") != "" {
		return viper.GetString("whetu_contract_address")
	}

	// 3. 如果都找不到，返回硬编码的地址（仅用于测试，生产环境应该配置）
	return "0x6AE1198a992b550aa56626f236E7CBd62a785C1F" // 替换为实际部署的WHETU合约地址
}

// 标准ERC20 ABI
const erc20AbiJSON = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "_spender",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "approve",
		"outputs": [
			{
				"name": "",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "_owner",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"name": "balance",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "_owner",
				"type": "address"
			},
			{
				"name": "_spender",
				"type": "address"
			}
		],
		"name": "allowance",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "mint",
		"outputs": [
			{
				"name": "",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
