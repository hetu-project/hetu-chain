package yuma

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock 在每个区块结束时运行
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// 简化的模块，暂时不需要复杂的区块结束处理
	return []abci.ValidatorUpdate{}
}
