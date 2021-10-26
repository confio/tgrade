package contract

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type DistributionQuery struct {
	WithdrawableFunds *WithdrawableFundsQuery `json:"withdrawable_funds,omitempty"`
}

type WithdrawableFundsQuery struct {
	Owner string `json:"owner"`
}
type FundsResponse struct {
	Funds sdk.Coin
}

type DistributionContractImpl struct {
	contractAddr     sdk.AccAddress
	contractQuerier  types.SmartQuerier
	addressLookupErr error
}

// NewDistributionContractImpl constructor
func NewDistributionContractImpl(contractAddr sdk.AccAddress, contractQuerier types.SmartQuerier, addressLookupErr error) *DistributionContractImpl {
	return &DistributionContractImpl{contractAddr: contractAddr, contractQuerier: contractQuerier, addressLookupErr: addressLookupErr}
}

func (d DistributionContractImpl) ValidatorOutstandingReward(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
	if d.addressLookupErr != nil {
		return sdk.Coin{}, d.addressLookupErr
	}
	query := DistributionQuery{WithdrawableFunds: &WithdrawableFundsQuery{Owner: addr.String()}}
	var resp FundsResponse
	err := doQuery(ctx, d.contractQuerier, d.contractAddr, query, &resp)
	if err != nil {
		return sdk.Coin{}, castError(err)
	}
	return resp.Funds, err
}

func castError(err error) error {
	const notFound = "tg4_engagement::state::WithdrawAdjustment not found"
	if strings.HasPrefix(err.Error(), notFound) {
		return types.ErrNotFound
	}
	return err
}
