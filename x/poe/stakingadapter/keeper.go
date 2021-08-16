package stakingadapter

import (
	"context"
	"errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// StakingAdapter connect to POE contract

var (
	_ wasmtypes.StakingKeeper       = &StakingAdapter{}
	_ wasmtypes.DistributionKeeper  = &StakingAdapter{}
	_ wasmkeeper.ValidatorSetSource = &StakingAdapter{}
	_ evidencetypes.StakingKeeper   = &StakingAdapter{}
	_ minttypes.StakingKeeper       = &StakingAdapter{}
	_ govtypes.StakingKeeper        = &StakingAdapter{}
)

var ErrNotImplemented = errors.New("not implemented")

type poeKeeper interface{}

type StakingAdapter struct {
	k              poeKeeper
	contractKeeper wasmtypes.ContractOpsKeeper
}

func NewStakingAdapter(k poeKeeper, contractKeeper wasmtypes.ContractOpsKeeper) StakingAdapter {
	return StakingAdapter{k: k, contractKeeper: contractKeeper}
}

func (s StakingAdapter) BondDenom(ctx sdk.Context) (res string) {
	log(ctx, "BondDenom")
	return "utgd"
}

func (s StakingAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	log(ctx, "BondDenom")
	return validator, false
}

func (s StakingAdapter) GetBondedValidatorsByPower(ctx sdk.Context) []stakingtypes.Validator {
	log(ctx, "GetBondedValidatorsByPower")
	return nil
}

func (s StakingAdapter) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	log(ctx, "GetAllDelegatorDelegations")
	return nil
}

func (s StakingAdapter) GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool) {
	log(ctx, "GetDelegation")
	return
}

func (s StakingAdapter) HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool {
	log(ctx, "HasReceivingRedelegation")
	return false
}

func (s StakingAdapter) DelegationRewards(stdlibCtx context.Context, req *types.QueryDelegationRewardsRequest) (*types.QueryDelegationRewardsResponse, error) {
	log(sdk.UnwrapSDKContext(stdlibCtx), "DelegationRewards")
	return nil, nil
}

func (s StakingAdapter) ValidatorByConsAddr(ctx sdk.Context, address sdk.ConsAddress) stakingtypes.ValidatorI {
	log(ctx, "ValidatorByConsAddr")
	return nil
}

func (s StakingAdapter) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) (updates []abci.ValidatorUpdate, err error) {
	log(ctx, "ApplyAndReturnValidatorSetUpdates")
	return nil, nil
}

func (s StakingAdapter) IterateValidators(ctx sdk.Context, f func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	log(ctx, "IterateValidators")
}

func (s StakingAdapter) Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI {
	log(ctx, "Validator")
	return nil
}

func (s StakingAdapter) Slash(ctx sdk.Context, address sdk.ConsAddress, i int64, i2 int64, dec sdk.Dec) {
	log(ctx, "Slash")
}

func (s StakingAdapter) Jail(ctx sdk.Context, address sdk.ConsAddress) {
	log(ctx, "Jail")
}

func (s StakingAdapter) Unjail(ctx sdk.Context, address sdk.ConsAddress) {
	log(ctx, "Unjail")
}

func (s StakingAdapter) Delegation(ctx sdk.Context, address sdk.AccAddress, address2 sdk.ValAddress) stakingtypes.DelegationI {
	log(ctx, "Delegation")
	return nil
}

func (s StakingAdapter) MaxValidators(ctx sdk.Context) uint32 {
	log(ctx, "MaxValidators")
	return 100
}

func (s StakingAdapter) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	log(ctx, "StakingTokenSupply")
	return sdk.NewInt(1000000)
}

func (s StakingAdapter) BondedRatio(ctx sdk.Context) sdk.Dec {
	log(ctx, "BondedRatio")
	return sdk.ZeroDec()
}

func (s2 StakingAdapter) IterateBondedValidatorsByPower(ctx sdk.Context, f func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	log(ctx, "IterateBondedValidatorsByPower")
}

func (s2 StakingAdapter) TotalBondedTokens(ctx sdk.Context) sdk.Int {
	log(ctx, "TotalBondedTokens")
	return sdk.NewInt(0)
}

func (s2 StakingAdapter) IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(index int64, delegation stakingtypes.DelegationI) (stop bool)) {
	log(ctx, "IterateDelegations")
}

func log(ctx sdk.Context, msg string) {
	ctx.Logger().Error("NOT IMPLEMENTED: ", "fn", msg)
}
