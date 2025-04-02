package checkpointing

import (
	"context"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/hetu-project/hetu/v1/x/checkpointing/keeper"
	ckpttypes "github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

// InitGenesis initializes the checkpointing module's state from a provided genesis
// state.
func InitGenesis(ctx context.Context, k keeper.Keeper, accountKeeper authkeeper.AccountKeeper, genState ckpttypes.GenesisState) {
	// ensure checkpointing module account is set on genesis
	if acc := accountKeeper.GetModuleAccount(ctx, ckpttypes.ModuleName); acc == nil {
		// NOTE: shouldn't occur
		panic("the checkpointing module account has not been set")
	}

	// Set the parameters
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) *ckpttypes.GenesisState {
	return &ckpttypes.GenesisState{
		Params: k.GetParams(ctx),
		// Export other state as needed
		// ...
	}
}
