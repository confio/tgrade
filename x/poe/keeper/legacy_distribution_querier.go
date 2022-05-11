package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/confio/tgrade/x/poe/types"
)

var _ distributiontypes.QueryServer = &legacyDistributionGRPCQuerier{}

type legacyDistributionGRPCQuerier struct {
	keeper      ViewKeeper
	queryServer types.QueryServer
}

func NewLegacyDistributionGRPCQuerier(keeper ViewKeeper) *legacyDistributionGRPCQuerier { //nolint:golint
	return &legacyDistributionGRPCQuerier{keeper: keeper, queryServer: NewGrpcQuerier(keeper)}
}

func (q legacyDistributionGRPCQuerier) ValidatorOutstandingRewards(c context.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	resp, err := q.queryServer.ValidatorOutstandingReward(c, &types.QueryValidatorOutstandingRewardRequest{ValidatorAddress: req.ValidatorAddress})
	if err != nil {
		return nil, err
	}
	return &distributiontypes.QueryValidatorOutstandingRewardsResponse{
		Rewards: distributiontypes.ValidatorOutstandingRewards{
			Rewards: sdk.NewDecCoins(resp.Reward),
		},
	}, nil
}

func (q legacyDistributionGRPCQuerier) ValidatorCommission(c context.Context, req *distributiontypes.QueryValidatorCommissionRequest) (*distributiontypes.QueryValidatorCommissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, "operator address invalid")
	}

	return &distributiontypes.QueryValidatorCommissionResponse{
		Commission: distributiontypes.ValidatorAccumulatedCommission{},
	}, nil
}

func (q legacyDistributionGRPCQuerier) ValidatorSlashes(c context.Context, req *distributiontypes.QueryValidatorSlashesRequest) (*distributiontypes.QueryValidatorSlashesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	opAddr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "operator address invalid")
	}
	ctx := sdk.UnwrapSDKContext(c)
	got, err := q.keeper.ValsetContract(ctx).ListValidatorSlashing(ctx, opAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	slashes := make([]distributiontypes.ValidatorSlashEvent, len(got))
	for i, s := range got {
		slashes[i] = distributiontypes.ValidatorSlashEvent{
			ValidatorPeriod: s.Height,
			Fraction:        s.Portion,
		}
	}
	return &distributiontypes.QueryValidatorSlashesResponse{
		Slashes: slashes,
	}, nil
}

func (q legacyDistributionGRPCQuerier) DelegationRewards(c context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
	}

	return &distributiontypes.QueryDelegationRewardsResponse{}, nil
}

func (q legacyDistributionGRPCQuerier) DelegationTotalRewards(c context.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(req.DelegatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
	}

	return &distributiontypes.QueryDelegationTotalRewardsResponse{}, nil
}

func (q legacyDistributionGRPCQuerier) DelegatorValidators(c context.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}

	var validators []string
	res, err := q.queryServer.Validator(c, &stakingtypes.QueryValidatorRequest{ValidatorAddr: req.DelegatorAddress})
	switch {
	case err == nil:
		validators = []string{res.Validator.OperatorAddress}
	case status.Code(err) == codes.NotFound:
		validators = []string{}
	default:
		return nil, err
	}

	return &distributiontypes.QueryDelegatorValidatorsResponse{
		Validators: validators,
	}, nil
}

func (q legacyDistributionGRPCQuerier) DelegatorWithdrawAddress(c context.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DelegatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	ownerAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
	}

	// Query the `tg4-engagement` contract for the delegated withdraw address
	ctx := sdk.UnwrapSDKContext(c)
	gotVal, err := q.keeper.EngagementContract(ctx).QueryDelegated(ctx, ownerAddr)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &distributiontypes.QueryDelegatorWithdrawAddressResponse{
		WithdrawAddress: gotVal.Delegated,
	}, nil
}

func (q legacyDistributionGRPCQuerier) CommunityPool(c context.Context, req *distributiontypes.QueryCommunityPoolRequest) (*distributiontypes.QueryCommunityPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &distributiontypes.QueryCommunityPoolResponse{}, nil
}

// Params is not supported. Method returns default distribution module params.
func (q legacyDistributionGRPCQuerier) Params(c context.Context, req *distributiontypes.QueryParamsRequest) (*distributiontypes.QueryParamsResponse, error) {
	return &distributiontypes.QueryParamsResponse{
		Params: distributiontypes.Params{
			CommunityTax:        sdk.ZeroDec(),
			BaseProposerReward:  sdk.ZeroDec(),
			BonusProposerReward: sdk.ZeroDec(),
			WithdrawAddrEnabled: false,
		},
	}, nil
}
