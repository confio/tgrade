package keeper

import (
	"context"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = &grpcQuerier{}

// queryKeeper is a subset of the keeper's methods
type queryKeeper interface {
	IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool)
	IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}
type grpcQuerier struct {
	keeper queryKeeper
}

func NewGrpcQuerier(keeper queryKeeper) *grpcQuerier {
	return &grpcQuerier{keeper: keeper}
}

func (q grpcQuerier) PrivilegedContracts(c context.Context, _ *types.QueryPrivilegedContractsRequest) (*types.QueryPrivilegedContractsResponse, error) {
	var result types.QueryPrivilegedContractsResponse
	q.keeper.IteratePrivileged(sdk.UnwrapSDKContext(c), func(address sdk.AccAddress) bool {
		result.Contracts = append(result.Contracts, address.String())
		return false
	})
	return &result, nil
}

func (q grpcQuerier) ContractsByPrivilegeType(c context.Context, req *types.QueryContractsByPrivilegeTypeRequest) (*types.QueryContractsByPrivilegeTypeResponse, error) {
	var result types.QueryContractsByPrivilegeTypeResponse
	cType := types.PrivilegeTypeFrom(req.PrivilegeType)
	if cType == nil {
		return nil, status.Error(codes.NotFound, "privilege type")
	}
	q.keeper.IteratePrivilegedContractsByType(sdk.UnwrapSDKContext(c), *cType, func(_ uint8, contractAddr sdk.AccAddress) bool {
		result.Contracts = append(result.Contracts, contractAddr.String())
		return false
	})
	return &result, nil
}
