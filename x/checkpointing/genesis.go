package checkpointing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/hetu-project/hetu/v1/types"
	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	ckpttypes "github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// InitGenesis initializes the checkpointing module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, accountKeeper authkeeper.AccountKeeper, genState types.GenesisState) {
	// ensure checkpointing module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, ckpttypes.ModuleName); acc == nil {
		// NOTE: shouldn't occur
		panic("the checkpointing module account has not been set")
	}
}
