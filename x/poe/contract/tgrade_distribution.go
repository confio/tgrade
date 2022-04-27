package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type DistributionQuery struct {
	WithdrawableRewards *WithdrawableRewardsQuery `json:"withdrawable_rewards,omitempty"`
}

type WithdrawableRewardsQuery struct {
	Owner string `json:"owner"`
}
type RewardsResponse struct {
	Rewards sdk.Coin
}

type DistributionContractAdapter struct {
	contractAddr     sdk.AccAddress
	contractQuerier  types.SmartQuerier
	addressLookupErr error
}

// NewDistributionContractAdapter constructor
func NewDistributionContractAdapter(contractAddr sdk.AccAddress, contractQuerier types.SmartQuerier, addressLookupErr error) *DistributionContractAdapter {
	return &DistributionContractAdapter{contractAddr: contractAddr, contractQuerier: contractQuerier, addressLookupErr: addressLookupErr}
}

func (d DistributionContractAdapter) ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
	if d.addressLookupErr != nil {
		return sdk.Coin{}, d.addressLookupErr
	}
	query := DistributionQuery{WithdrawableRewards: &WithdrawableRewardsQuery{Owner: addr.String()}}
	var resp RewardsResponse
	err := doQuery(ctx, d.contractQuerier, d.contractAddr, query, &resp)
	if err != nil {
		return sdk.Coin{}, err
	}
	return resp.Rewards, err
}

func (d DistributionContractAdapter) Address() (sdk.AccAddress, error) {
	return d.contractAddr, d.addressLookupErr
}
