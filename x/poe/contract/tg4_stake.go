package contract

import (
	"encoding/json"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TG4StakeInitMsg struct {
	Admin           string `json:"admin,omitempty"`
	Denom           string `json:"denom"`
	MinBond         uint64 `json:"min_bond,string"`
	TokensPerWeight uint64 `json:"tokens_per_weight,string"`
	// UnbondingPeriod unbonding period in seconds
	UnbondingPeriod uint64 `json:"unbonding_period"`
	// AutoReturnLimit limits how much claims would be automatically returned at end of block, 20 by default. Setting this to 0 disables auto returning claims.
	AutoReturnLimit *uint64 `json:"auto_return_limit,omitempty"`
	Preauths        uint64  `json:"preauths,omitempty"`
}

func (m TG4StakeInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

// TG4StakeQuery contains some custom queries for the tg4-stake contract.
// You can also make any generic TG4Query on it.
type TG4StakeQuery struct {
	UnbondingPeriod *struct{}             `json:"unbonding_period,omitempty"`
	Claims          *TG4StakeQueryAddress `json:"claims,omitempty"`
	Staked          *TG4StakeQueryAddress `json:"staked,omitempty"`
}

type TG4StakeQueryAddress struct {
	Address string `json:"address"`
}
type TG4StakeClaimsResponse struct {
	Claims []TG4StakeClaim `json:"claims"`
}

type TG4StakedAmountsResponse struct {
	Stake wasmvmtypes.Coin `json:"stake"`
}

type TG4StakeClaim struct {
	// Addr A human readable address
	Addr string `json:"addr"`
	// Amount of tokens in claim
	Amount sdk.Int `json:"amount"`
	// CreationHeight Height of a blockchain in a moment of creation of this claim
	CreationHeight uint64 `json:"creation_height"`
	// ReleaseAt is the release time of the claim as timestamp in nanoseconds
	ReleaseAt uint64 `json:"release_at,string,omitempty"`
}

type UnbondingPeriodResponse struct {
	// Time is the number of seconds that must pass
	UnbondingPeriod uint64 `json:"unbonding_period"`
}

// QueryStakingUnbondingPeriod query the unbonding period from PoE staking contract
func QueryStakingUnbondingPeriod(ctx sdk.Context, k types.SmartQuerier, stakeAddr sdk.AccAddress) (uint64, error) {
	query := TG4StakeQuery{UnbondingPeriod: &struct{}{}}
	var response UnbondingPeriodResponse
	err := doQuery(ctx, k, stakeAddr, query, &response)
	return response.UnbondingPeriod, err
}

// QueryStakingUnbonding query PoE staking contract for unbonded self delegations
func QueryStakingUnbonding(ctx sdk.Context, k types.SmartQuerier, stakeAddr sdk.AccAddress, opAddr sdk.AccAddress) (TG4StakeClaimsResponse, error) {
	query := TG4StakeQuery{Claims: &TG4StakeQueryAddress{Address: opAddr.String()}}
	var response TG4StakeClaimsResponse
	err := doQuery(ctx, k, stakeAddr, query, &response)
	return response, err
}

// QueryStakedAmount query PoE staking contract for bonded self delegation amount
func QueryStakedAmount(ctx sdk.Context, k types.SmartQuerier, stakeAddr sdk.AccAddress, opAddr sdk.AccAddress) (TG4StakedAmountsResponse, error) {
	query := TG4StakeQuery{Staked: &TG4StakeQueryAddress{Address: opAddr.String()}}
	var response TG4StakedAmountsResponse
	err := doQuery(ctx, k, stakeAddr, query, &response)
	return response, err
}

func doQuery(ctx sdk.Context, k types.SmartQuerier, contractAddr sdk.AccAddress, query interface{}, result interface{}) error {
	bz, err := json.Marshal(query)
	if err != nil {
		return err
	}
	res, err := k.QuerySmart(ctx, contractAddr, bz)
	if err != nil {
		return err
	}
	return json.Unmarshal(res, result)
}
