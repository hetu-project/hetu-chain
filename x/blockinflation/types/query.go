package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"

	pb "github.com/hetu-project/hetu/v1/x/blockinflation/types/generated"
)

// 将生成的 proto 类型转换为本地类型
func ProtoParamsToParams(protoParams *pb.Params) Params {
	totalSupply, _ := math.NewIntFromString(protoParams.TotalSupply)
	defaultBlockEmission, _ := math.NewIntFromString(protoParams.DefaultBlockEmission)
	subnetRewardBase, _ := math.LegacyNewDecFromStr(protoParams.SubnetRewardBase)
	subnetRewardK, _ := math.LegacyNewDecFromStr(protoParams.SubnetRewardK)
	subnetRewardMaxRatio, _ := math.LegacyNewDecFromStr(protoParams.SubnetRewardMaxRatio)
	subnetMovingAlpha, _ := math.LegacyNewDecFromStr(protoParams.SubnetMovingAlpha)
	subnetOwnerCut, _ := math.LegacyNewDecFromStr(protoParams.SubnetOwnerCut)

	return NewParams(
		protoParams.EnableBlockInflation,
		protoParams.MintDenom,
		totalSupply,
		defaultBlockEmission,
		subnetRewardBase,
		subnetRewardK,
		subnetRewardMaxRatio,
		subnetMovingAlpha,
		subnetOwnerCut,
	)
}

// QueryParamsRequest is the request type for the Query/Params RPC method.
type QueryParamsRequest struct{}

// QueryParamsResponse is the response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	// Params defines the parameters of the module.
	Params Params `json:"params"`
}

// QueryPendingSubnetRewardsRequest is the request type for the Query/PendingSubnetRewards RPC method.
type QueryPendingSubnetRewardsRequest struct{}

// QueryPendingSubnetRewardsResponse is the response type for the Query/PendingSubnetRewards RPC method.
type QueryPendingSubnetRewardsResponse struct {
	// PendingSubnetRewards defines the pending subnet rewards.
	PendingSubnetRewards sdk.Coin `json:"pending_subnet_rewards"`
}

// QueryClient is the client API for Query service.
type QueryClient interface {
	// Params queries the parameters of the module.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// PendingSubnetRewards queries the pending subnet rewards.
	PendingSubnetRewards(ctx context.Context, in *QueryPendingSubnetRewardsRequest, opts ...grpc.CallOption) (*QueryPendingSubnetRewardsResponse, error)
}

// 实现 QueryClient 接口，使用生成的 gRPC 客户端
type queryClient struct {
	grpcClient pb.QueryClient
}

// NewQueryClient 创建一个新的查询客户端
func NewQueryClient(grpcConn grpc.ClientConnInterface) QueryClient {
	return &queryClient{grpcClient: pb.NewQueryClient(grpcConn)}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	// 使用生成的 gRPC 客户端调用 Params 方法
	res, err := c.grpcClient.Params(ctx, &pb.QueryParamsRequest{}, opts...)
	if err != nil {
		return nil, err
	}

	// 将 proto 响应转换为本地类型
	return &QueryParamsResponse{
		Params: ProtoParamsToParams(res.Params),
	}, nil
}

func (c *queryClient) PendingSubnetRewards(ctx context.Context, in *QueryPendingSubnetRewardsRequest, opts ...grpc.CallOption) (*QueryPendingSubnetRewardsResponse, error) {
	// 使用生成的 gRPC 客户端调用 PendingSubnetRewards 方法
	res, err := c.grpcClient.PendingSubnetRewards(ctx, &pb.QueryPendingSubnetRewardsRequest{}, opts...)
	if err != nil {
		return nil, err
	}

	// 将 proto 响应转换为本地类型
	return &QueryPendingSubnetRewardsResponse{
		PendingSubnetRewards: sdk.Coin{
			Denom:  res.PendingSubnetRewards.Denom,
			Amount: math.NewIntFromBigInt(res.PendingSubnetRewards.Amount.BigInt()),
		},
	}, nil
}
