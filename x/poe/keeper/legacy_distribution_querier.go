package keeper

import (
	"context"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ distributiontypes.QueryServer = &legacyDistributionGRPCQuerier{}

type legacyDistributionGRPCQuerier struct {
	keeper      ContractSource
	queryServer types.QueryServer
}

func NewLegacyDistributionGRPCQuerier(keeper ViewKeeper, contractQuerier types.SmartQuerier) *legacyDistributionGRPCQuerier {
	return &legacyDistributionGRPCQuerier{keeper: keeper, queryServer: NewGrpcQuerier(keeper, contractQuerier)}
}

func (q legacyDistributionGRPCQuerier) ValidatorOutstandingRewards(c context.Context, req *distributiontypes.QueryValidatorOutstandingRewardsRequest) (*distributiontypes.QueryValidatorOutstandingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
	}

	return &distributiontypes.QueryValidatorOutstandingRewardsResponse{
		Rewards: distributiontypes.ValidatorOutstandingRewards{},
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
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
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
	if _, err := sdk.AccAddressFromBech32(req.ValidatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")
	}

	return &distributiontypes.QueryValidatorSlashesResponse{}, nil
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
	add, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "delegator address invalid")

	}
	return &distributiontypes.QueryDelegatorWithdrawAddressResponse{
		WithdrawAddress: add.String(),
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
		}}, nil
}
