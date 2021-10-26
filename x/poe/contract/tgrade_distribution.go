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

func QueryWithdrawableFunds(ctx sdk.Context, k types.SmartQuerier, contractAddr, owner sdk.AccAddress) (sdk.Coin, error) {
	query := DistributionQuery{WithdrawableFunds: &WithdrawableFundsQuery{Owner: owner.String()}}
	var resp FundsResponse
	err := doQuery(ctx, k, contractAddr, query, &resp)
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
