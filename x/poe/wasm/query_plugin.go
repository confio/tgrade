package wasm

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/keeper"
)

type ViewKeeper interface {
	GetBondDenom(ctx sdk.Context) string
	DistributionContract(ctx sdk.Context) keeper.DistributionContract
	ValsetContract(ctx sdk.Context) keeper.ValsetContract
	StakeContract(ctx sdk.Context) keeper.StakeContract
}

func StakingQuerier(poeKeeper ViewKeeper) func(ctx sdk.Context, request *wasmvmtypes.StakingQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.StakingQuery) ([]byte, error) {
		if request.BondedDenom != nil {
			denom := poeKeeper.GetBondDenom(ctx)
			res := wasmvmtypes.BondedDenomResponse{
				Denom: denom,
			}
			return json.Marshal(res)
		}
		zero := sdk.ZeroDec().String()
		if request.AllValidators != nil {
			validators, err := poeKeeper.ValsetContract(ctx).ListValidators(ctx)
			if err != nil {
				return nil, err
			}
			wasmVals := make([]wasmvmtypes.Validator, len(validators))
			for i, v := range validators {
				wasmVals[i] = wasmvmtypes.Validator{
					Address:       v.OperatorAddress,
					Commission:    zero,
					MaxCommission: zero,
					MaxChangeRate: zero,
				}
			}
			res := wasmvmtypes.AllValidatorsResponse{
				Validators: wasmVals,
			}
			return json.Marshal(res)
		}
		if request.Validator != nil {
			valAddr, err := sdk.AccAddressFromBech32(request.Validator.Address)
			if err != nil {
				return nil, err
			}
			v, err := poeKeeper.ValsetContract(ctx).QueryValidator(ctx, valAddr)
			if err != nil {
				return nil, err
			}
			res := wasmvmtypes.ValidatorResponse{}
			if v != nil {
				res.Validator = &wasmvmtypes.Validator{
					Address:       v.OperatorAddress,
					Commission:    zero,
					MaxCommission: zero,
					MaxChangeRate: zero,
				}
			}
			return json.Marshal(res)
		}
		if request.AllDelegations != nil {
			delegator, err := sdk.AccAddressFromBech32(request.AllDelegations.Delegator)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.AllDelegations.Delegator)
			}
			stakedAmount, err := poeKeeper.StakeContract(ctx).QueryStakedAmount(ctx, delegator)
			if err != nil {
				return nil, err
			}
			var res wasmvmtypes.AllDelegationsResponse
			if stakedAmount != nil {
				res.Delegations = append(res.Delegations, wasmvmtypes.Delegation{
					Delegator: delegator.String(),
					Validator: delegator.String(),
					Amount:    wasmvmtypes.NewCoin(stakedAmount.Uint64(), poeKeeper.GetBondDenom(ctx)),
				})
			}
			return json.Marshal(res)
		}
		if request.Delegation != nil {
			delegator, err := sdk.AccAddressFromBech32(request.Delegation.Delegator)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Delegation.Delegator)
			}
			validator, err := sdk.AccAddressFromBech32(request.Delegation.Validator)
			if err != nil {
				return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, request.Delegation.Validator)
			}

			var res wasmvmtypes.DelegationResponse
			if !delegator.Equals(validator) { // no match
				return json.Marshal(res)
			}
			stakeContract := poeKeeper.StakeContract(ctx)
			stakedAmount, err := stakeContract.QueryStakedAmount(ctx, delegator)
			if err != nil {
				return nil, err
			}
			if stakedAmount != nil {
				stakedCoin := wasmvmtypes.NewCoin(stakedAmount.Uint64(), poeKeeper.GetBondDenom(ctx))
				res.Delegation = &wasmvmtypes.FullDelegation{
					Delegator:          delegator.String(),
					Validator:          delegator.String(),
					Amount:             stakedCoin,
					CanRedelegate:      stakedCoin,
					AccumulatedRewards: nil,
				}
				reward, err := poeKeeper.DistributionContract(ctx).ValidatorOutstandingReward(ctx, delegator)
				if err != nil {
					return nil, err
				}
				res.Delegation.AccumulatedRewards = wasmvmtypes.Coins{wasmvmtypes.NewCoin(reward.Amount.Uint64(), reward.Denom)}
			}
			return json.Marshal(res)
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown Staking variant"}
	}
}
