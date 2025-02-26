package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hetu-project/hetu/v1/x/checkpointing/types"
)

var _ types.QueryServer = Keeper{}

// RawCheckpointList returns a list of checkpoint by status in the ascending order of epoch
func (k Keeper) RawCheckpointList(c context.Context, req *types.QueryRawCheckpointListRequest) (*types.QueryRawCheckpointListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	var checkpointList []*types.RawCheckpointWithMetaResponse

	ctx := sdk.UnwrapSDKContext(c)

	store := k.CheckpointsState(ctx).checkpoints
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		ckptWithMeta, err := types.BytesToCkptWithMeta(k.cdc, value)
		if err != nil {
			return false, err
		}
		if ckptWithMeta.Status == req.Status {
			if accumulate {
				checkpointList = append(checkpointList, ckptWithMeta.ToResponse())
			}
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryRawCheckpointListResponse{RawCheckpoints: checkpointList, Pagination: pageRes}, nil
}

// RawCheckpoint returns a checkpoint by epoch number
func (k Keeper) RawCheckpoint(c context.Context, req *types.QueryRawCheckpointRequest) (*types.QueryRawCheckpointResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	ckptWithMeta, err := k.CheckpointsState(ctx).GetRawCkptWithMeta(req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryRawCheckpointResponse{RawCheckpoint: ckptWithMeta.ToResponse()}, nil
}

// EpochStatus returns the status of the checkpoint at a given epoch
func (k Keeper) EpochStatus(ctx context.Context, req *types.QueryEpochStatusRequest) (*types.QueryEpochStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ckptWithMeta, err := k.CheckpointsState(sdkCtx).GetRawCkptWithMeta(req.EpochNum)
	if err != nil {
		return nil, err
	}

	return &types.QueryEpochStatusResponse{Status: ckptWithMeta.Status}, nil
}
