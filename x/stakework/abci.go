package stakework

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock runs at the end of each block
/*
This file implements the Cosmos SDK's ABCI (Application Blockchain Interface) interface:
Purpose:
EndBlock: Called at the end of each block, used to execute logic at block end
Currently a simplified implementation, only returns an empty validator update list
Can add cleanup work, state updates, etc. at block end here
*/
func EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
