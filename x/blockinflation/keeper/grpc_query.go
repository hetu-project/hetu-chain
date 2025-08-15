package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/hetu-project/hetu/v1/x/blockinflation/types/generated"
)

// Params implements the generated QueryServer.Params method
func (k Keeper) Params(c context.Context, req *pb.QueryParamsRequest) (*pb.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	protoParams := &pb.Params{
		EnableBlockInflation: params.EnableBlockInflation,
		MintDenom:            params.MintDenom,
		TotalSupply:          params.TotalSupply.String(),
		DefaultBlockEmission: params.DefaultBlockEmission.String(),
		SubnetRewardBase:     params.SubnetRewardBase.String(),
		SubnetRewardK:        params.SubnetRewardK.String(),
		SubnetRewardMaxRatio: params.SubnetRewardMaxRatio.String(),
		SubnetMovingAlpha:    params.SubnetMovingAlpha.String(),
		SubnetOwnerCut:       params.SubnetOwnerCut.String(),
	}

	return &pb.QueryParamsResponse{Params: protoParams}, nil
}

// PendingSubnetRewards implements the generated QueryServer.PendingSubnetRewards method
func (k Keeper) PendingSubnetRewards(c context.Context, req *pb.QueryPendingSubnetRewardsRequest) (*pb.QueryPendingSubnetRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	pendingRewards := k.GetPendingSubnetRewards(ctx)

	protoCoin := &sdk.Coin{
		Denom:  pendingRewards.Denom,
		Amount: pendingRewards.Amount,
	}

	return &pb.QueryPendingSubnetRewardsResponse{
		PendingSubnetRewards: protoCoin,
	}, nil
}
