package keeper

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
)

// SyncAMMPoolState 同步链上状态和合约 AMM 池状态
func (k Keeper) SyncAMMPoolState(ctx sdk.Context, netuid uint16) error {
	// 1. 获取子网信息，包括 AMM 池地址
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// 2. 验证 AMM 池地址
	ammPoolAddress, ok := subnet.Params["amm_pool"]
	if !ok || ammPoolAddress == "" || !common.IsHexAddress(ammPoolAddress) {
		return fmt.Errorf("invalid or missing AMM pool address in subnet params: %d", netuid)
	}
	ammPoolAddr := common.HexToAddress(ammPoolAddress)

	// 3. 获取 SubnetAMM 合约的 ABI
	ammABI, err := getSubnetAMMABI()
	if err != nil {
		return fmt.Errorf("failed to load SubnetAMM ABI: %w", err)
	}

	// 4. 从合约中获取当前池状态
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
	result, err := k.erc20Keeper.CallEVM(
		ctx,
		ammABI,
		common.BytesToAddress(moduleAddress.Bytes()),
		ammPoolAddr,
		false, // 只读调用
		"getPoolInfo",
	)
	if err != nil {
		return fmt.Errorf("failed to get AMM pool info: %w", err)
	}

	// 5. 解析返回结果
	if result == nil || len(result.Ret) < 32*8 {
		return fmt.Errorf("invalid result from getPoolInfo call")
	}

	// 从返回的字节数组中提取数据
	// 每个uint256值占32字节
	// 注意：这里的解析方式取决于合约返回值的具体编码方式
	// 这里假设返回值是按照顺序排列的8个uint256值
	subnetHetu := new(big.Int).SetBytes(result.Ret[32:64])      // 第2个uint256
	subnetAlphaIn := new(big.Int).SetBytes(result.Ret[64:96])   // 第3个uint256
	subnetAlphaOut := new(big.Int).SetBytes(result.Ret[96:128]) // 第4个uint256
	currentPrice := new(big.Int).SetBytes(result.Ret[128:160])  // 第5个uint256
	movingPrice := new(big.Int).SetBytes(result.Ret[160:192])   // 第6个uint256

	// 6. 更新链上状态以匹配合约状态
	taoIn := math.NewIntFromBigInt(subnetHetu)
	alphaIn := math.NewIntFromBigInt(subnetAlphaIn)

	// 更新链上状态
	k.eventKeeper.SetSubnetTaoIn(ctx, netuid, taoIn)
	k.eventKeeper.SetSubnetAlphaIn(ctx, netuid, alphaIn)

	// 记录日志
	k.Logger(ctx).Info("Synced AMM pool state with contract",
		"netuid", netuid,
		"tao_in", taoIn.String(),
		"alpha_in", alphaIn.String(),
		"alpha_out", subnetAlphaOut.String(),
		"current_price", currentPrice.String(),
		"moving_price", movingPrice.String(),
	)

	return nil
}

// SyncAllAMMPools 同步所有活跃子网的 AMM 池状态
func (k Keeper) SyncAllAMMPools(ctx sdk.Context) {
	allSubnets := k.eventKeeper.GetAllSubnetNetuids(ctx)
	for _, netuid := range allSubnets {
		if err := k.SyncAMMPoolState(ctx, netuid); err != nil {
			k.Logger(ctx).Error("Failed to sync AMM pool state",
				"netuid", netuid,
				"error", err,
			)
		}
	}
}
