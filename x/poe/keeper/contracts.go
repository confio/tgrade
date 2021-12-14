package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

type DistributionContract interface {
	// ValidatorOutstandingReward returns amount or 0 for an unknown address
	ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error)
}

func (k Keeper) DistributionContract(ctx sdk.Context) DistributionContract {
	distContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	return contract.NewDistributionContractAdapter(distContractAddr, k.twasmKeeper, err)
}

type ValsetContract interface {
	ListValidators(ctx sdk.Context, pagination *types.Paginator) ([]stakingtypes.Validator, error)
	QueryValidator(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error)
	ListValidatorSlashing(ctx sdk.Context, opAddr sdk.AccAddress) ([]contract.ValidatorSlashing, error)
	QueryConfig(ctx sdk.Context) (*contract.ValsetConfigResponse, error)
	UpdateAdmin(ctx sdk.Context, new sdk.AccAddress, sender sdk.AccAddress) error
}

func (k Keeper) ValsetContract(ctx sdk.Context) ValsetContract {
	distContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	return contract.NewValsetContractAdapter(distContractAddr, k.twasmKeeper, err)
}

type StakeContract interface {
	// QueryStakedAmount returns amount in default denom or nil value for an unknown address
	QueryStakedAmount(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error)
	QueryStakingUnbondingPeriod(ctx sdk.Context) (time.Duration, error)
	// QueryStakingUnbonding returns the unbondings or empty list for an unknown address
	QueryStakingUnbonding(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error)
}

func (k Keeper) StakeContract(ctx sdk.Context) StakeContract {
	distContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	return contract.NewStakeContractAdapter(distContractAddr, k.twasmKeeper, err)
}

type EngagementContract interface {
	UpdateAdmin(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error
}

func (k Keeper) EngagementContract(ctx sdk.Context) EngagementContract {
	engContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	return contract.NewEngagementContractAdapter(engContractAddr, k.twasmKeeper, err)
}
