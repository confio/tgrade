package poetesting

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/contract"
)

// var _ keeper.DistributionContract = DistributionContractMock{}

type DistributionContractMock struct {
	ValidatorOutstandingRewardFn func(ctx types.Context, addr types.AccAddress) (types.Coin, error)
}

func (m DistributionContractMock) ValidatorOutstandingReward(ctx types.Context, addr types.AccAddress) (types.Coin, error) {
	if m.ValidatorOutstandingRewardFn == nil {
		panic("not expected to be called")
	}
	return m.ValidatorOutstandingRewardFn(ctx, addr)
}

// var _ keeper.ValsetContract = ValsetContractMock{}

type ValsetContractMock struct {
	QueryValidatorFn func(ctx types.Context, opAddr types.AccAddress) (*stakingtypes.Validator, error)
	ListValidatorsFn func(ctx types.Context) ([]stakingtypes.Validator, error)
	QueryConfigFn    func(ctx types.Context) (*contract.ValsetConfigResponse, error)
}

func (m ValsetContractMock) QueryValidator(ctx types.Context, opAddr types.AccAddress) (*stakingtypes.Validator, error) {
	if m.QueryValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.QueryValidatorFn(ctx, opAddr)
}

func (m ValsetContractMock) ListValidators(ctx types.Context) ([]stakingtypes.Validator, error) {
	if m.ListValidatorsFn == nil {
		panic("not expected to be called")
	}
	return m.ListValidatorsFn(ctx)
}

func (m ValsetContractMock) QueryConfig(ctx types.Context) (*contract.ValsetConfigResponse, error) {
	if m.QueryConfigFn == nil {
		panic("not expected to be called")
	}
	return m.QueryConfigFn(ctx)
}

// var _ keeper.StakeContract = StakeContractMock{}

type StakeContractMock struct {
	QueryStakingUnbondingPeriodFn func(ctx types.Context) (time.Duration, error)
	QueryStakingUnbondingFn       func(ctx types.Context, opAddr types.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error)
	QueryStakedAmountFn           func(ctx types.Context, opAddr types.AccAddress) (*types.Int, error)
}

func (m StakeContractMock) QueryStakedAmount(ctx types.Context, opAddr types.AccAddress) (*types.Int, error) {
	if m.QueryStakedAmountFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakedAmountFn(ctx, opAddr)
}

func (m StakeContractMock) QueryStakingUnbondingPeriod(ctx types.Context) (time.Duration, error) {
	if m.QueryStakingUnbondingPeriodFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingPeriodFn(ctx)
}
func (m StakeContractMock) QueryStakingUnbonding(ctx types.Context, opAddr types.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
	if m.QueryStakingUnbondingFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingFn(ctx, opAddr)
}

// var _ keeper.EngagementContract = EngagementContractMock{}

type EngagementContractMock struct {
	UpdateAdminFn func(ctx types.Context, newAdmin, sender types.AccAddress) error
}

func (m EngagementContractMock) UpdateAdmin(ctx types.Context, newAdmin, sender types.AccAddress) error {
	if m.UpdateAdminFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateAdminFn(ctx, newAdmin, sender)
}
