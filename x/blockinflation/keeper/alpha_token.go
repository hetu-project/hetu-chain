package keeper

import (
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/hetu-project/hetu/v1/x/blockinflation/types"
	eventabi "github.com/hetu-project/hetu/v1/x/event/abi"
)

// MintAlphaTokens mints alpha tokens to the specified address
func (k Keeper) MintAlphaTokens(ctx sdk.Context, netuid uint16, recipient string, amount uint64) error {
	// 1. Get the subnet info to find the AlphaToken address
	subnet, found := k.eventKeeper.GetSubnet(ctx, netuid)
	if !found {
		return fmt.Errorf("subnet not found: %d", netuid)
	}

	// 获取子网信息
	subnetInfo, found := getSubnetInfo(subnet)
	if !found || subnetInfo.AlphaToken == "" {
		return fmt.Errorf("subnet has no alpha token: %d", netuid)
	}

	// 3. Parse the AlphaToken address
	alphaTokenAddr := common.HexToAddress(subnetInfo.AlphaToken)
	if (alphaTokenAddr == common.Address{}) {
		return fmt.Errorf("invalid alpha token address: %s", subnetInfo.AlphaToken)
	}

	// 4. Get the AlphaToken ABI
	alphaTokenABI, err := abi.JSON(strings.NewReader(string(eventabi.AlphaTokenABI)))
	if err != nil {
		return fmt.Errorf("failed to parse AlphaToken ABI: %w", err)
	}

	// 5. Convert recipient address to Ethereum address
	recipientAddr := common.HexToAddress(recipient)
	if (recipientAddr == common.Address{}) {
		return fmt.Errorf("invalid recipient address: %s", recipient)
	}

	// 6. Convert amount to big.Int
	amountBig := new(big.Int).SetUint64(amount)

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
		amountBig,
	)
	if err != nil {
		return fmt.Errorf("failed to mint alpha tokens: %w", err)
	}

	k.Logger(ctx).Info("Minted alpha tokens",
		"netuid", netuid,
		"recipient", recipient,
		"amount", amount,
		"token_address", alphaTokenAddr.Hex(),
	)

	return nil
}
