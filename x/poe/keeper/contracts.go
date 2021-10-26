package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

type DistributionContract interface {
	ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error)
}

func (k Keeper) DistributionContract(ctx sdk.Context) DistributionContract {
	distContractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeDistribution)
	return contract.NewDistributionContractImpl(distContractAddr, k.twasmKeeper, err)
}
