package yuma

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker 在每个区块结束时运行
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// 处理Yuma共识相关逻辑
	am.keeper.EndBlock(ctx)

	// 返回验证者更新
	return am.keeper.GetValidatorUpdates(ctx)
}
