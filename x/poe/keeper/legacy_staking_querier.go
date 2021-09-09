package keeper

import (
	"context"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ stakingtypes.QueryServer = &legacyStakingGRPCQuerier{}

type legacyStakingGRPCQuerier struct {
	keeper          Keeper
	contractQuerier types.SmartQuerier
	queryServer     types.QueryServer
}

func NewLegacyStakingGRPCQuerier(poeKeeper Keeper, q types.SmartQuerier) *legacyStakingGRPCQuerier {
	return &legacyStakingGRPCQuerier{keeper: poeKeeper, contractQuerier: q, queryServer: NewGrpcQuerier(poeKeeper, q)}
}

func (q legacyStakingGRPCQuerier) Validators(c context.Context, req *stakingtypes.QueryValidatorsRequest) (*stakingtypes.QueryValidatorsResponse, error) {
	return q.queryServer.Validators(c, req)
}

func (q legacyStakingGRPCQuerier) Validator(c context.Context, req *stakingtypes.QueryValidatorRequest) (*stakingtypes.QueryValidatorResponse, error) {
	return q.queryServer.Validator(c, req)
}

func (q legacyStakingGRPCQuerier) ValidatorDelegations(c context.Context, req *stakingtypes.QueryValidatorDelegationsRequest) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "ValidatorDelegations")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) ValidatorUnbondingDelegations(c context.Context, req *stakingtypes.QueryValidatorUnbondingDelegationsRequest) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "ValidatorUnbondingDelegations")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) Delegation(c context.Context, req *stakingtypes.QueryDelegationRequest) (*stakingtypes.QueryDelegationResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "Delegation")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) UnbondingDelegation(c context.Context, req *stakingtypes.QueryUnbondingDelegationRequest) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "UnbondingDelegation")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) DelegatorDelegations(c context.Context, req *stakingtypes.QueryDelegatorDelegationsRequest) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "DelegatorDelegations")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) DelegatorUnbondingDelegations(c context.Context, req *stakingtypes.QueryDelegatorUnbondingDelegationsRequest) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "DelegatorUnbondingDelegations")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) Redelegations(c context.Context, req *stakingtypes.QueryRedelegationsRequest) (*stakingtypes.QueryRedelegationsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "Redelegations")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) DelegatorValidators(c context.Context, req *stakingtypes.QueryDelegatorValidatorsRequest) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "DelegatorValidators")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) DelegatorValidator(c context.Context, req *stakingtypes.QueryDelegatorValidatorRequest) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "DelegatorValidator")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) HistoricalInfo(c context.Context, req *stakingtypes.QueryHistoricalInfoRequest) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "HistoricalInfo")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) Pool(c context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "Pool")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

func (q legacyStakingGRPCQuerier) Params(c context.Context, req *stakingtypes.QueryParamsRequest) (*stakingtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &stakingtypes.QueryParamsResponse{
		Params: stakingtypes.Params{
			UnbondingTime:     q.keeper.UnbondingTime(ctx),
			MaxValidators:     100,
			MaxEntries:        0,
			HistoricalEntries: q.keeper.HistoricalEntries(ctx),
			BondDenom:         "utgd",
		},
	}, nil
}

func logNotImplemented(ctx sdk.Context, msg string) {
	ctx.Logger().Error("NOT IMPLEMENTED: ", "fn", msg)
}
