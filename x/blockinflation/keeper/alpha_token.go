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
)

var (
	alphaABIOnce   sync.Once
	alphaABIVal    abi.ABI
	alphaABIValErr error
)

// getAlphaTokenABI returns the cached AlphaToken ABI or parses it once if not cached
func getAlphaTokenABI() (abi.ABI, error) {
	alphaABIOnce.Do(func() {
		alphaABIVal, alphaABIValErr = abi.JSON(bytes.NewReader(eventabi.AlphaTokenABI))
	})
	return alphaABIVal, alphaABIValErr
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

	// 7. Call the mint function on the AlphaToken contract
	moduleAddress := authtypes.NewModuleAddress(types.ModuleName)
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
