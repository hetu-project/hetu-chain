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

// SyncAMMPoolState Sync the chain state and the AMM pool state of the contract
func (k Keeper) SyncAMMPoolState(ctx sdk.Context, netuid uint16) error {
	k.Logger(ctx).Debug("Starting to sync AMM pool state", "netuid", netuid)

	// 1. Get the subnet information, including the AMM pool address
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

	// Print all parameters, help debugging
	for key, value := range subnet.Params {
		k.Logger(ctx).Debug("Subnet parameter",
			"netuid", netuid,
			"key", key,
			"value", value)
	}

	// 2. Verify the AMM pool address
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

	// 3. Get the ABI of the SubnetAMM contract
	ammABI, err := getSubnetAMMABI()
	if err != nil {
		k.Logger(ctx).Error("Failed to load SubnetAMM ABI", "error", err)
		return fmt.Errorf("failed to load SubnetAMM ABI: %w", err)
	}
	k.Logger(ctx).Debug("Loaded SubnetAMM ABI")

	// 4. Get the current pool state from the contract
	moduleAddress := authtypes.NewModuleAddress(blockinflationtypes.ModuleName)
	k.Logger(ctx).Debug("Calling EVM to get pool info",
		"module_address", moduleAddress.String(),
		"amm_pool_address", ammPoolAddr.Hex())

	result, err := k.erc20Keeper.CallEVM(
		ctx,
		ammABI,
		common.BytesToAddress(moduleAddress.Bytes()),
		ammPoolAddr,
		false, // Read only call
		"getPoolInfo",
	)
	if err != nil {
		k.Logger(ctx).Error("Failed to call AMM pool contract",
			"netuid", netuid,
			"amm_pool_address", ammPoolAddr.Hex(),
			"error", err)
		return fmt.Errorf("failed to get AMM pool info: %w", err)
	}

	// 5. Parse the return result
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

	// Extract data from the returned byte array
	// Each uint256 value takes up 32 bytes
	// Note: The parsing method here depends on the specific encoding of the return value
	// Here, it is assumed that the return values are 8 uint256 values arranged in order
	// Note: The parsing method here depends on the specific encoding of the return value
	// Here it is assumed that the return value is an 8 uint256 values in order
	// Note: The parsing method here depends on the specific encoding of the return value
	// Here it is assumed that the return value is an 8 uint256 values in order
	// Note: The parsing method here depends on the specific encoding of the return value
	// Here it is assumed that the return value is an 8 uint256 values in order
	mechanismType := new(big.Int).SetBytes(result.Ret[0:32])    // The first uint256- Mechanism type
	subnetHetu := new(big.Int).SetBytes(result.Ret[32:64])      // The second uint256 - TAO number
	subnetAlphaIn := new(big.Int).SetBytes(result.Ret[64:96])   // The third uint256 - AlphaIn number
	subnetAlphaOut := new(big.Int).SetBytes(result.Ret[96:128]) // The fourth uint256 - AlphaOut number
	currentPrice := new(big.Int).SetBytes(result.Ret[128:160])  // The fifth uint256 - Current price
	movingPrice := new(big.Int).SetBytes(result.Ret[160:192])   // The sixth uint256 - Moving average price
	totalVolume := new(big.Int).SetBytes(result.Ret[192:224])   // The seventh uint256 - Total volume

	k.Logger(ctx).Debug("Parsed values from result",
		"netuid", netuid,
		"mechanism_type", mechanismType.String(),
		"subnet_hetu", subnetHetu.String(),
		"subnet_alpha_in", subnetAlphaIn.String(),
		"subnet_alpha_out", subnetAlphaOut.String(),
		"current_price", currentPrice.String(),
		"moving_price", movingPrice.String(),
		"total_volume", totalVolume.String())

	// 6. Get the chain state
	currentTaoIn := k.eventKeeper.GetSubnetTAO(ctx, netuid)
	currentAlphaIn := k.eventKeeper.GetSubnetAlphaIn(ctx, netuid)
	currentAlphaOut := k.eventKeeper.GetSubnetAlphaOut(ctx, netuid)

	k.Logger(ctx).Debug("Current chain state",
		"netuid", netuid,
		"current_tao_in", currentTaoIn.String(),
		"current_alpha_in", currentAlphaIn.String(),
		"current_alpha_out", currentAlphaOut.String())

	// 7. Check if the chain state is consistent with the contract state
	contractTaoIn := math.NewIntFromBigInt(subnetHetu)
	contractAlphaIn := math.NewIntFromBigInt(subnetAlphaIn)
	// No longer use the AlphaOut value of the contract
	// contractAlphaOut := math.NewIntFromBigInt(subnetAlphaOut)

	// 8. Only update TaoIn and AlphaIn, do not update AlphaOut

	// 8.1 Handle TaoIn
	if !currentTaoIn.Equal(contractTaoIn) {
		k.Logger(ctx).Info("Updating chain TaoIn to match contract",
			"netuid", netuid,
			"old_tao_in", currentTaoIn.String(),
			"new_tao_in", contractTaoIn.String())
		k.eventKeeper.SetSubnetTaoIn(ctx, netuid, contractTaoIn)
	}

	// 8.2 Handle AlphaIn
	if !currentAlphaIn.Equal(contractAlphaIn) {
		k.Logger(ctx).Info("Updating chain AlphaIn to match contract",
			"netuid", netuid,
			"old_alpha_in", currentAlphaIn.String(),
			"new_alpha_in", contractAlphaIn.String())
		k.eventKeeper.SetSubnetAlphaIn(ctx, netuid, contractAlphaIn)
	}

	// 8.3 No longer handle AlphaOut, keep the chain value
	k.Logger(ctx).Debug("Keeping chain AlphaOut value (not syncing from contract)",
		"netuid", netuid,
		"chain_alpha_out", currentAlphaOut.String(),
		"contract_alpha_out", subnetAlphaOut.String())

	// 9. Update the moving average price
	// Get the current moving average price
	currentMovingPrice := k.eventKeeper.GetMovingAlphaPrice(ctx, netuid)
	contractMovingPriceDec := math.LegacyNewDecFromBigInt(movingPrice)

	// If the moving average price is inconsistent, update the chain state
	if !currentMovingPrice.Equal(contractMovingPriceDec) {
		k.Logger(ctx).Info("Updating moving price",
			"netuid", netuid,
			"old_moving_price", currentMovingPrice.String(),
			"new_moving_price", contractMovingPriceDec.String())

		// Use the UpdateMovingPrice method to update the moving average price
		// Get the EMAPriceHalvingBlocks of the subnet
		halvingBlocks := subnet.EMAPriceHalvingBlocks
		// Use the moving average price of the contract as the basis, call UpdateMovingPrice to update
		params := k.GetParams(ctx)
		k.eventKeeper.UpdateMovingPrice(ctx, netuid, params.SubnetMovingAlpha, halvingBlocks)
	}

	return nil
}

// SyncAllAMMPools Sync the AMM pool state of all active subnets
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

// HandleSubnetRegisteredEvent Handle the subnet registered event, immediately sync the AMM pool state
// This function should be called when listening to the subnet registered event
func (k Keeper) HandleSubnetRegisteredEvent(ctx sdk.Context, netuid uint16, ammPoolAddress string) {
	k.Logger(ctx).Info("Handling subnet registered event",
		"netuid", netuid,
		"amm_pool_address", ammPoolAddress)

	// Immediately sync the AMM pool state
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

// SyncChainStateToContract Sync the chain state to the contract
// Called after calculating the rewards, inject TaoIn and AlphaIn into the AMM contract
func (k Keeper) SyncChainStateToContract(ctx sdk.Context, netuid uint16, reward blockinflationtypes.SubnetRewards) error {
	// Get the subnet information
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// Verify the AMM pool address
	ammPoolAddress := subnet.AmmPool
	if ammPoolAddress == "" || !common.IsHexAddress(ammPoolAddress) {
		return fmt.Errorf("invalid AMM pool address: %s", ammPoolAddress)
	}

	// Get the ABI
	ammABI, err := getSubnetAMMABI()
	if err != nil {
		return fmt.Errorf("failed to load SubnetAMM ABI: %w", err)
	}

	// Prepare the call parameters
	moduleAddress := authtypes.NewModuleAddress(blockinflationtypes.ModuleName)
	ammPoolAddr := common.HexToAddress(ammPoolAddress)

	// Only inject liquidity when TaoIn is positive
	if reward.TaoIn.IsPositive() {
		// Calculate the amount of AlphaIn to inject
		alphaInAmount := reward.AlphaIn

		k.Logger(ctx).Debug("Preparing to inject liquidity to contract",
			"netuid", netuid,
			"tao_in", reward.TaoIn.String(),
			"alpha_in", alphaInAmount.String(),
			"amm_pool_address", ammPoolAddr.Hex())

		// Get the WHETU token address from the subnet information
		// This needs to be set in the configuration or obtained from the environment variable
		whetuAddress := getWHETUAddress()
		if whetuAddress == "" || !common.IsHexAddress(whetuAddress) {
			return fmt.Errorf("invalid or missing WHETU token address")
		}
		whetuAddr := common.HexToAddress(whetuAddress)

		// 1. Mint the Cosmos native HETU token
		params := k.GetParams(ctx)
		err = k.bankKeeper.MintCoins(
			ctx,
			blockinflationtypes.ModuleName,
			sdk.NewCoins(sdk.NewCoin(params.MintDenom, reward.TaoIn)),
		)
		if err != nil {
			return fmt.Errorf("failed to mint HETU tokens: %w", err)
		}

		// 2. Convert the native HETU to WHETU
		// Use the deposit function of the WHETU contract, need to send the native HETU
		whetuABI, err := abi.JSON(bytes.NewReader(eventabi.WHETUABI))
		if err != nil {
			return fmt.Errorf("failed to parse WHETU ABI: %w", err)
		}

		// Create the call data for the deposit function
		depositData, err := whetuABI.Pack("deposit")
		if err != nil {
			return fmt.Errorf("failed to pack deposit function data: %w", err)
		}

		// Use evmKeeper to send the transaction directly
		evmModuleAddr := common.BytesToAddress(moduleAddress.Bytes())

		// Get the nonce of the module account
		nonce := k.evmKeeper.GetNonce(ctx, evmModuleAddr)

		// Create a message instead of a transaction
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

		// Directly call ApplyMessage instead of EthereumTx
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

		// 3. Mint the Alpha token and transfer it to the module account
		if alphaInAmount.IsPositive() {
			// Use the MintAlphaTokens function, it will handle the authorization check and minting
			// Mint the Alpha token to the module account, using the EVM format address
			evmAddress := common.BytesToAddress(moduleAddress.Bytes()).Hex()
			err = k.MintAlphaTokens(ctx, netuid, evmAddress, alphaInAmount.BigInt())
			if err != nil {
				return fmt.Errorf("failed to mint Alpha tokens: %w", err)
			}

			k.Logger(ctx).Info("Successfully minted Alpha tokens for AMM pool",
				"netuid", netuid,
				"alpha_amount", alphaInAmount.String())
		}

		// Get the subnet detailed information to get the AlphaToken address
		subnetInfo, found := k.eventKeeper.GetSubnetInfo(ctx, netuid)
		if !found {
			return fmt.Errorf("subnet info not found: %d", netuid)
		}

		// Verify the AlphaToken address
		alphaTokenAddress := subnetInfo.AlphaToken
		if alphaTokenAddress == "" || !common.IsHexAddress(alphaTokenAddress) {
			return fmt.Errorf("invalid or missing alpha token address: %d", netuid)
		}

		k.Logger(ctx).Debug("Retrieved token addresses",
			"netuid", netuid,
			"whetu_token", whetuAddr.Hex(),
			"alpha_token", alphaTokenAddress)

		// Check the WHETU token balance of the module account
		erc20ABI, err := abi.JSON(strings.NewReader(erc20AbiJSON))
		if err != nil {
			return fmt.Errorf("failed to parse ERC20 ABI: %w", err)
		}

		// Check the WHETU balance
		whetuBalanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			false, // Read only call
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

		// Check the Alpha balance
		alphaBalanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			common.HexToAddress(alphaTokenAddress),
			false, // Read only call
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

		// Check if the balance is sufficient
		if whetuBalance.Cmp(reward.TaoIn.BigInt()) < 0 {
			return fmt.Errorf("insufficient WHETU balance: have %s, need %s", whetuBalance.String(), reward.TaoIn.BigInt().String())
		}

		if alphaBalance.Cmp(alphaInAmount.BigInt()) < 0 {
			return fmt.Errorf("insufficient Alpha balance: have %s, need %s", alphaBalance.String(), alphaInAmount.BigInt().String())
		}

		// 4. Approve the AMM contract to use the WHETU token
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			true, // Submit transaction
			"approve",
			ammPoolAddr,
			reward.TaoIn.BigInt(),
		)
		if err != nil {
			return fmt.Errorf("failed to approve WHETU tokens: %w", err)
		}

		// 5. Approve the AMM contract to use the Alpha token
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			common.HexToAddress(alphaTokenAddress),
			true, // Submit transaction
			"approve",
			ammPoolAddr,
			alphaInAmount.BigInt(),
		)
		if err != nil {
			return fmt.Errorf("failed to approve Alpha tokens: %w", err)
		}

		// Check the authorization amount after approval
		whetuAllowanceResult, err := k.erc20Keeper.CallEVM(
			ctx,
			erc20ABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			whetuAddr,
			false, // Read only call
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
			false, // Read only call
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

		// Check if the authorization amount is sufficient
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

		// 6. Call the injectLiquidity method, inject the WHETU and Alpha token
		_, err = k.erc20Keeper.CallEVM(
			ctx,
			ammABI,
			common.BytesToAddress(moduleAddress.Bytes()),
			ammPoolAddr,
			true, // Submit transaction
			"injectLiquidity",
			reward.TaoIn.BigInt(),
			alphaInAmount.BigInt(),
		)

		if err != nil {
			// Try to get the AMM contract state, which may help diagnose the problem
			poolInfoResult, poolInfoErr := k.erc20Keeper.CallEVM(
				ctx,
				ammABI,
				common.BytesToAddress(moduleAddress.Bytes()),
				ammPoolAddr,
				false, // Read only call
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

// getWHETUAddress Get the WHETU token address
// Get the WHETU token address from the environment variable first, if the environment variable does not exist, try to read from the configuration file
func getWHETUAddress() string {
	// 1. Get the WHETU token address from the environment variable first
	envVarName := "WHETU_CONTRACT_ADDRESS"
	if addr := os.Getenv(envVarName); addr != "" {
		return addr
	}

	// 2. If the environment variable does not exist, try to read from the configuration file
	if viper.GetString("whetu_contract_address") != "" {
		return viper.GetString("whetu_contract_address")
	}

	// 3. If none are found, return the hardcoded address (only for testing, the production environment should be configured)
	return "0x6AE1198a992b550aa56626f236E7CBd62a785C1F" // Replace with the actual deployed WHETU contract address
}

// Standard ERC20 ABI
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
