package contract

import (
	"encoding/json"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValsetQuery will create many queries for the valset contract
// https://github.com/confio/tgrade-contracts/blob/v0.3.0/contracts/tgrade-valset/src/msg.rs
type ValsetQuery struct {
	Config                   *struct{}            `json:"config,omitempty"`
	Epoch                    *struct{}            `json:"epoch,omitempty"`
	Validator                *ValidatorQuery      `json:"validator,omitempty"`
	ListValidators           *ListValidatorsQuery `json:"list_validators,omitempty"`
	ListActiveValidators     *struct{}            `json:"list_active_validators,omitempty"`
	SimulateActiveValidators *struct{}            `json:"simulate_active_validators,omitempty"`
}

type ValidatorQuery struct {
	Operator string `json:"operator"`
}

type ListValidatorsQuery struct {
	StartAfter string `json:"start_after,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// ValsetConfigResponse Response to `config` query
type ValsetConfigResponse struct {
	Membership    string `json:"membership"`
	MinWeight     int    `json:"min_weight"`
	MaxValidators int    `json:"max_validators"`
	Scaling       int    `json:"scaling,omitempty"`
}

// ValsetEpochQueryResponse Response to `config` query
type ValsetEpochResponse struct {
	// Number of seconds in one epoch. We update the Tendermint validator set only once per epoch.
	EpochLength int `json:"epoch_length"`
	// The current epoch # (block.time/epoch_length, rounding down)
	CurrentEpoch int `json:"current_epoch"`
	// The last time we updated the validator set - block time (in seconds)
	LastUpdateTime int `json:"last_update_time"`
	// The last time we updated the validator set - block height
	LastUpdateHeight int `json:"last_update_height"`
	/// Seconds (UTC UNIX time) of next timestamp that will trigger a validator recalculation
	NextUpdateTime int `json:"next_update_time"`
}

type OperatorResponse struct {
	Operator string            `json:"operator"`
	Pubkey   ValidatorPubkey   `json:"pubkey"`
	Metadata ValidatorMetadata `json:"metadata"`
}

type ValidatorInfo struct {
	Operator        string          `json:"operator"`
	ValidatorPubkey ValidatorPubkey `json:"validator_pubkey"`
	Power           int             `json:"power"`
}

type ValidatorResponse struct {
	Validator *OperatorResponse `json:"validator"`
}

type ListValidatorsResponse struct {
	Validators []OperatorResponse `json:"validators"`
}

type ListActiveValidatorsResponse struct {
	Validators []ValidatorInfo `json:"validators"`
}

type SimulateActiveValidatorsResponse = ListActiveValidatorsResponse

func QueryValsetConfig(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress) (*ValsetConfigResponse, error) {
	query := ValsetQuery{Config: &struct{}{}}
	var response ValsetConfigResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValsetEpoch(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress) (*ValsetEpochResponse, error) {
	query := ValsetQuery{Epoch: &struct{}{}}
	var response ValsetEpochResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValidator(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress, operator sdk.AccAddress) (*OperatorResponse, error) {
	query := ValsetQuery{Validator: &ValidatorQuery{Operator: operator.String()}}
	var response ValidatorResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validator, err
}

// TODO: add auto-pagination support
func ListValidators(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress) ([]OperatorResponse, error) {
	query := ValsetQuery{ListValidators: &ListValidatorsQuery{Limit: 30}}
	var response ListValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func ListActiveValidators(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{ListActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func SimulateActiveValidators(ctx sdk.Context, k *twasmkeeper.Keeper, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{SimulateActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

// These queries apply to all tg4 types - stake, group, and mixer
type TG4Query struct {
	TotalWeight         *struct{}                 `json:"total_weight,omitempty"`
	ListMembers         *ListMembersQuery         `json:"list_members,omitempty"`
	ListMembersByWeight *ListMembersByWeightQuery `json:"list_members_by_weight,omitempty"`
	Member              *MemberQuery              `json:"member,omitempty"`
}

type ListMembersQuery struct {
	StartAfter string `json:"start_after,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type ListMembersByWeightQuery struct {
	StartAfter *TG4Member `json:"start_after,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

type MemberQuery struct {
	Addr     string `json:"addr"`
	AtHeight int    `json:"at_height,omitempty"`
}

// TG4MemberListResponse response to a list members query.
type TG4MemberListResponse struct {
	Members []TG4Member `json:"members"`
}

type TG4MemberResponse struct {
	// nil means not a member, 0 means member with no voting power... this can be a very important distinction
	Weight *int `json:"weight"`
}

type TG4TotalWeightResponse struct {
	Weight int `json:"weight"`
}

// TODO: add pagination
func QueryTG4MembersByWeight(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress) ([]TG4Member, error) {
	query := TG4Query{ListMembersByWeight: &ListMembersByWeightQuery{Limit: 30}}
	var response TG4MemberListResponse
	err := doQuery(ctx, k, groupAddr, query, &response)
	return response.Members, err
}

func QueryTG4Members(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress) ([]TG4Member, error) {
	query := TG4Query{ListMembers: &ListMembersQuery{Limit: 30}}
	var response TG4MemberListResponse
	err := doQuery(ctx, k, groupAddr, query, &response)
	return response.Members, err
}

// TODO: expose at height (if we care)
// Returns the weight of this member. (nil, nil) means not present
func QueryTG4Member(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress, member sdk.AccAddress) (*int, error) {
	query := TG4Query{Member: &MemberQuery{Addr: member.String()}}
	var response TG4MemberResponse
	err := doQuery(ctx, k, groupAddr, query, &response)
	return response.Weight, err
}

// TODO: expose at height (if we care)
// Returns the weight of this member. (nil, nil) means not present
func QueryTG4TotalWeight(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress) (int, error) {
	query := TG4Query{TotalWeight: &struct{}{}}
	var response TG4TotalWeightResponse
	err := doQuery(ctx, k, groupAddr, query, &response)
	return response.Weight, err
}

func doQuery(ctx sdk.Context, k *twasmkeeper.Keeper, contractAddr sdk.AccAddress, query interface{}, result interface{}) error {
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
