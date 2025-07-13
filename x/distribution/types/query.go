package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// QueryParamsRequest is the request type for the Query/Params RPC method.
type QueryParamsRequest struct{}

// QueryParamsResponse is the response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

// QuerySubnetRewardRatioRequest is the request type for the Query/SubnetRewardRatio RPC method.
type QuerySubnetRewardRatioRequest struct{}

// QuerySubnetRewardRatioResponse is the response type for the Query/SubnetRewardRatio RPC method.
type QuerySubnetRewardRatioResponse struct {
	SubnetRewardRatio string `protobuf:"bytes,1,opt,name=subnet_reward_ratio,json=subnetRewardRatio,proto3" json:"subnet_reward_ratio,omitempty"`
	SubnetCount       int64  `protobuf:"varint,2,opt,name=subnet_count,json=subnetCount,proto3" json:"subnet_count,omitempty"`
}

// QueryServer defines the gRPC querier service for distribution module
type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	SubnetRewardRatio(context.Context, *QuerySubnetRewardRatioRequest) (*QuerySubnetRewardRatioResponse, error)
}

// RegisterLegacyAminoCodec registers the legacy amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register types for amino codec
}

// RegisterInterfaces registers the interfaces
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Register interfaces
}
