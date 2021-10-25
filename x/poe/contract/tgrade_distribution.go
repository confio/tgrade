package contract

import (
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
	Funds sdk.DecCoin
}

func QueryWithdrawableFunds(ctx sdk.Context, k types.SmartQuerier, contractAddr, owner sdk.AccAddress) (sdk.DecCoin, error) {
	query := DistributionQuery{WithdrawableFunds: &WithdrawableFundsQuery{Owner: owner.String()}}
	var resp FundsResponse
	err := doQuery(ctx, k, contractAddr, query, &resp)
	return resp.Funds, err
}
