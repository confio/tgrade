package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type DistributionQuery struct {
	WithdrawableRewards *WithdrawableRewardsQuery `json:"withdrawable_funds,omitempty"`
}

type WithdrawableRewardsQuery struct {
	Owner string `json:"owner"`
}
type FundsResponse struct {
	Funds sdk.Coin
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
	var resp FundsResponse
	err := doQuery(ctx, d.contractQuerier, d.contractAddr, query, &resp)
	if err != nil {
		return sdk.Coin{}, err
	}
	return resp.Funds, err
}
