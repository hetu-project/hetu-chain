package generated

import (
	"context"

	"google.golang.org/grpc"

	eventtypes "github.com/hetu-project/hetu/v1/hetu/event/v1"
)

// QueryClient defines the client API for Query service.
type QueryClient interface {
	// Subnets returns all subnets with pagination
	Subnets(ctx context.Context, in *eventtypes.QuerySubnetsRequest, opts ...grpc.CallOption) (*eventtypes.QuerySubnetsResponse, error)
	// Subnet returns subnet info by ID
	Subnet(ctx context.Context, in *eventtypes.QuerySubnetRequest, opts ...grpc.CallOption) (*eventtypes.QuerySubnetResponse, error)
	// SubnetNeurons returns all neurons in a subnet
	SubnetNeurons(ctx context.Context, in *eventtypes.QuerySubnetNeuronsRequest, opts ...grpc.CallOption) (*eventtypes.QuerySubnetNeuronsResponse, error)
	// SubnetPool returns pool info for a subnet
	SubnetPool(ctx context.Context, in *eventtypes.QuerySubnetPoolRequest, opts ...grpc.CallOption) (*eventtypes.QuerySubnetPoolResponse, error)
	// ValidatorWeights returns weights for a validator in a subnet
	ValidatorWeights(ctx context.Context, in *eventtypes.QueryValidatorWeightsRequest, opts ...grpc.CallOption) (*eventtypes.QueryValidatorWeightsResponse, error)
}

// QueryServer defines the server API for Query service.
type QueryServer interface {
	// Subnets returns all subnets with pagination
	Subnets(ctx context.Context, in *eventtypes.QuerySubnetsRequest) (*eventtypes.QuerySubnetsResponse, error)
	// Subnet returns subnet info by ID
	Subnet(ctx context.Context, in *eventtypes.QuerySubnetRequest) (*eventtypes.QuerySubnetResponse, error)
	// SubnetNeurons returns all neurons in a subnet
	SubnetNeurons(ctx context.Context, in *eventtypes.QuerySubnetNeuronsRequest) (*eventtypes.QuerySubnetNeuronsResponse, error)
	// SubnetPool returns pool info for a subnet
	SubnetPool(ctx context.Context, in *eventtypes.QuerySubnetPoolRequest) (*eventtypes.QuerySubnetPoolResponse, error)
	// ValidatorWeights returns weights for a validator in a subnet
	ValidatorWeights(ctx context.Context, in *eventtypes.QueryValidatorWeightsRequest) (*eventtypes.QueryValidatorWeightsResponse, error)
}

// NewQueryClient creates a new QueryClient instance.
func NewQueryClient(conn grpc.ClientConnInterface) QueryClient {
	return eventtypes.NewQueryClient(conn)
}

// RegisterQueryServer registers the QueryServer implementation with the gRPC server.
func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	eventtypes.RegisterQueryServer(s, &queryServerWrapper{
		srv: srv,
	})
}

// queryServerWrapper wraps the QueryServer implementation to make it compatible with eventtypes.QueryServer.
type queryServerWrapper struct {
	srv QueryServer
	eventtypes.UnimplementedQueryServer
}

func (q *queryServerWrapper) Subnets(ctx context.Context, req *eventtypes.QuerySubnetsRequest) (*eventtypes.QuerySubnetsResponse, error) {
	return q.srv.Subnets(ctx, req)
}

func (q *queryServerWrapper) Subnet(ctx context.Context, req *eventtypes.QuerySubnetRequest) (*eventtypes.QuerySubnetResponse, error) {
	return q.srv.Subnet(ctx, req)
}

func (q *queryServerWrapper) SubnetNeurons(ctx context.Context, req *eventtypes.QuerySubnetNeuronsRequest) (*eventtypes.QuerySubnetNeuronsResponse, error) {
	return q.srv.SubnetNeurons(ctx, req)
}

func (q *queryServerWrapper) SubnetPool(ctx context.Context, req *eventtypes.QuerySubnetPoolRequest) (*eventtypes.QuerySubnetPoolResponse, error) {
	return q.srv.SubnetPool(ctx, req)
}

func (q *queryServerWrapper) ValidatorWeights(ctx context.Context, req *eventtypes.QueryValidatorWeightsRequest) (*eventtypes.QueryValidatorWeightsResponse, error) {
	return q.srv.ValidatorWeights(ctx, req)
}
