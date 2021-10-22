package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

// GetEngagementPoints read engagement points from contract
func (k Keeper) GetEngagementPoints(ctx sdk.Context, opAddr sdk.AccAddress) (uint64, error) {
	engagementContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "get contract addr")
	}
	weight, err := contract.QueryTG4Member(ctx, k.twasmKeeper, engagementContractAddr, opAddr)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "query staking contract")
	}
	if weight == nil {
		return 0, nil
	}
	return uint64(*weight), nil
}

// setEngagementPoints set new engagement point value.
func (k Keeper) setEngagementPoints(ctx sdk.Context, opAddr sdk.AccAddress, points uint64) error {
	engagementContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	if err != nil {
		return sdkerrors.Wrap(err, "get contract addr")
	}
	return contract.SetEngagementPoints(ctx, engagementContractAddr, k.twasmKeeper, opAddr, points)
}

// SetValidatorInitialEngagementPoints set an initial amount of engagement points for a validator when it matches self delegation preconditions
func (k Keeper) SetValidatorInitialEngagementPoints(ctx sdk.Context, opAddr sdk.AccAddress, selfDelegation sdk.Coin) error {
	// distribute engagement points enabled ?
	newPoints := k.GetInitialValidatorEngagementPoints(ctx)
	if newPoints == 0 {
		return nil
	}
	// qualifies for initial engagement points?
	min := k.MinimumDelegationAmounts(ctx)
	if !sdk.NewCoins(selfDelegation).IsAllGTE(min) {
		// we could also fail here to communicate the lack of self delegation but this makes it more complicated
		// for genesis validators as this condition is checked on chain
		return nil
	}
	currentPoints, err := k.GetEngagementPoints(ctx, opAddr)
	if err != nil {
		return sdkerrors.Wrap(err, "get engagement points")
	}
	if currentPoints >= newPoints {
		return nil
	}
	err = k.setEngagementPoints(ctx, opAddr, newPoints)
	return sdkerrors.Wrap(err, "set engagement points")
}
