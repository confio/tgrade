package poetesting

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/contract"
)

// var _ keeper.DistributionContract = DistributionContractMock{}

type DistributionContractMock struct {
	ValidatorOutstandingRewardFn func(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error)
}

func (m DistributionContractMock) ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
	if m.ValidatorOutstandingRewardFn == nil {
		panic("not expected to be called")
	}
	return m.ValidatorOutstandingRewardFn(ctx, addr)
}

// var _ keeper.ValsetContract = ValsetContractMock{}

type ValsetContractMock struct {
	QueryValidatorFn        func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error)
	ListValidatorsFn        func(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error)
	QueryConfigFn           func(ctx sdk.Context) (*contract.ValsetConfigResponse, error)
	ListValidatorSlashingFn func(ctx sdk.Context, opAddr sdk.AccAddress) ([]contract.ValidatorSlashing, error)
	UpdateAdminFn           func(ctx sdk.Context, new sdk.AccAddress, sender sdk.AccAddress) error
}

func (m ValsetContractMock) UpdateAdmin(ctx sdk.Context, new sdk.AccAddress, sender sdk.AccAddress) error {
	if m.UpdateAdminFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateAdminFn(ctx, new, sender)
}

func (m ValsetContractMock) QueryValidator(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
	if m.QueryValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.QueryValidatorFn(ctx, opAddr)
}

func (m ValsetContractMock) ListValidators(ctx sdk.Context, pagination *contract.Paginator) ([]stakingtypes.Validator, contract.PaginationCursor, error) {
	if m.ListValidatorsFn == nil {
		panic("not expected to be called")
	}
	return m.ListValidatorsFn(ctx, pagination)
}

func (m ValsetContractMock) QueryConfig(ctx sdk.Context) (*contract.ValsetConfigResponse, error) {
	if m.QueryConfigFn == nil {
		panic("not expected to be called")
	}
	return m.QueryConfigFn(ctx)
}

func (m ValsetContractMock) ListValidatorSlashing(ctx sdk.Context, opAddr sdk.AccAddress) ([]contract.ValidatorSlashing, error) {
	if m.ListValidatorSlashingFn == nil {
		panic("not expected to be called")
	}
	return m.ListValidatorSlashingFn(ctx, opAddr)
}

// var _ keeper.StakeContract = StakeContractMock{}

type StakeContractMock struct {
	QueryStakingUnbondingPeriodFn func(ctx sdk.Context) (time.Duration, error)
	QueryStakingUnbondingFn       func(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error)
	QueryStakedAmountFn           func(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error)
}

func (m StakeContractMock) QueryStakedAmount(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
	if m.QueryStakedAmountFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakedAmountFn(ctx, opAddr)
}

func (m StakeContractMock) QueryStakingUnbondingPeriod(ctx sdk.Context) (time.Duration, error) {
	if m.QueryStakingUnbondingPeriodFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingPeriodFn(ctx)
}
func (m StakeContractMock) QueryStakingUnbonding(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
	if m.QueryStakingUnbondingFn == nil {
		panic("not expected to be called")
	}
	return m.QueryStakingUnbondingFn(ctx, opAddr)
}

// var _ keeper.EngagementContract = EngagementContractMock{}

type EngagementContractMock struct {
	UpdateAdminFn func(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error
}

func (m EngagementContractMock) UpdateAdmin(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error {
	if m.UpdateAdminFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateAdminFn(ctx, newAdmin, sender)
}
