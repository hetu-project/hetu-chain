package stakework

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock 在每个区块结束时运行
/*
这个文件实现了 Cosmos SDK 的 ABCI (Application Blockchain Interface) 接口：
作用：
EndBlock: 在每个区块结束时被调用，用于执行区块结束时的逻辑
目前是简化实现，只返回空的验证者更新列表
可以在这里添加区块结束时的清理工作、状态更新等
*/
func (am AppModule) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
