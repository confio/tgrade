package contract

import (
	"encoding/json"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-stake/src/msg.rs
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

// TG4StakeExecute staking contract execute messages
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-stake/src/msg.rs
type TG4StakeExecute struct {
	Bond   *struct{} `json:"bond,omitempty"`
	Unbond *Unbond   `json:"unbond,omitempty"`
	Claim  *struct{} `json:"claim,omitempty"`
}

// Unbond will start the unbonding process for the given number of tokens. The sender immediately loses weight from these tokens, and can claim them back to his wallet after `unbonding_period`",
type Unbond struct {
	// Tokens are the amount to unbond
	Tokens sdk.Int `json:"tokens"`
}

// TG4StakeQuery contains some custom queries for the tg4-stake contract.
// You can also make any generic TG4Query on it.
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-stake/src/msg.rs
type TG4StakeQuery struct {
	UnbondingPeriod *struct{}        `json:"unbonding_period,omitempty"`
	Claims          *ListClaimsQuery `json:"claims,omitempty"`
	Staked          *StakedQuery     `json:"staked,omitempty"`
}

type StakedQuery struct {
	Address string `json:"address"`
}

type ListClaimsQuery struct {
	Address string `json:"address"`
	// Limit for pagination
	Limit uint32 `json:"limit,omitempty"`
	// StartAfter is used for pagination. Take last `claim.ReleaseAt` from last query
	StartAfter uint64 `json:"start_after,string,omitempty"`
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

// QueryStakedAmount query PoE staking contract for bonded self delegation amount
func QueryStakedAmount(ctx sdk.Context, k types.SmartQuerier, stakeAddr sdk.AccAddress, opAddr sdk.AccAddress) (TG4StakedAmountsResponse, error) {
	query := TG4StakeQuery{Staked: &StakedQuery{Address: opAddr.String()}}
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

type StakeContractAdapter struct {
	contractAddr     sdk.AccAddress
	contractQuerier  types.SmartQuerier
	addressLookupErr error
}

// NewStakeContractAdapter constructor
func NewStakeContractAdapter(contractAddr sdk.AccAddress, contractQuerier types.SmartQuerier, addressLookupErr error) *StakeContractAdapter {
	return &StakeContractAdapter{contractAddr: contractAddr, contractQuerier: contractQuerier, addressLookupErr: addressLookupErr}
}

// QueryStakingUnbondingPeriod query the unbonding period from PoE staking contract
func (v StakeContractAdapter) QueryStakingUnbondingPeriod(ctx sdk.Context) (time.Duration, error) {
	if v.addressLookupErr != nil {
		return 0, v.addressLookupErr
	}

	query := TG4StakeQuery{UnbondingPeriod: &struct{}{}}
	var resp UnbondingPeriodResponse
	err := doQuery(ctx, v.contractQuerier, v.contractAddr, query, &resp)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "contract query")
	}
	return time.Duration(resp.UnbondingPeriod) * time.Second, nil
}

func (v StakeContractAdapter) QueryStakedAmount(ctx sdk.Context, opAddr sdk.AccAddress) (*sdk.Int, error) {
	query := TG4Query{Member: &MemberQuery{Addr: opAddr.String()}}
	var resp TG4MemberResponse
	err := doQuery(ctx, v.contractQuerier, v.contractAddr, query, &resp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	if resp.Weight == nil {
		return nil, nil
	}
	amount := sdk.NewInt(int64(*resp.Weight))
	// we should return Coin instead: https://github.com/confio/tgrade-contracts/issues/265
	return &amount, nil
}

// QueryStakingUnbonding query PoE staking contract for unbonded self delegations
// TODO: add pagination support here!
func (v StakeContractAdapter) QueryStakingUnbonding(ctx sdk.Context, opAddr sdk.AccAddress) ([]stakingtypes.UnbondingDelegationEntry, error) {
	if v.addressLookupErr != nil {
		return nil, v.addressLookupErr
	}
	query := TG4StakeQuery{Claims: &ListClaimsQuery{Address: opAddr.String()}}
	var resp TG4StakeClaimsResponse
	err := doQuery(ctx, v.contractQuerier, v.contractAddr, query, &resp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}

	// add all unbonded amounts
	unbodings := make([]stakingtypes.UnbondingDelegationEntry, len(resp.Claims))
	for i, v := range resp.Claims {
		unbodings[i] = stakingtypes.UnbondingDelegationEntry{
			InitialBalance: v.Amount,
			CompletionTime: time.Unix(0, int64(v.ReleaseAt)).UTC(),
			Balance:        v.Amount,
			CreationHeight: int64(v.CreationHeight),
		}
	}
	return unbodings, nil
}
