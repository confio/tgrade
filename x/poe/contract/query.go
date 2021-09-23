package contract

import (
	"encoding/json"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/poe/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptosecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	// Percentage of total accumulated fees which is substracted from tokens minted as a rewards. A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0
	FeePercentage uint64 `json:"fee_percentage,string,omitempty"`
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

func (v OperatorResponse) ToValidator() (stakingtypes.Validator, error) {
	pubKey, err := toCosmosPubKey(v.Pubkey)
	if err != nil {
		return stakingtypes.Validator{}, sdkerrors.Wrap(err, "convert to cosmos key")
	}
	any, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return stakingtypes.Validator{}, sdkerrors.Wrap(err, "convert to any type")
	}

	return stakingtypes.Validator{
		OperatorAddress: v.Operator,
		ConsensusPubkey: any,
		Description:     v.Metadata.ToDescription(),
		DelegatorShares: sdk.OneDec(),
		Status:          stakingtypes.Bonded,
	}, nil
}

func toCosmosPubKey(key ValidatorPubkey) (cryptotypes.PubKey, error) {
	switch {
	case key.Ed25519 != nil:
		return &ed25519.PubKey{Key: key.Ed25519}, nil
	case key.Secp256k1 != nil:
		return &cryptosecp256k1.PubKey{Key: key.Secp256k1}, nil
	default:
		return nil, types.ErrValidatorPubKeyTypeNotSupported
	}
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

func QueryValsetConfig(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) (*ValsetConfigResponse, error) {
	query := ValsetQuery{Config: &struct{}{}}
	var response ValsetConfigResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValsetEpoch(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) (*ValsetEpochResponse, error) {
	query := ValsetQuery{Epoch: &struct{}{}}
	var response ValsetEpochResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValidator(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress, operator sdk.AccAddress) (*OperatorResponse, error) {
	query := ValsetQuery{Validator: &ValidatorQuery{Operator: operator.String()}}
	var response ValidatorResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validator, err
}

func ListValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]OperatorResponse, error) {
	query := ValsetQuery{ListValidators: &ListValidatorsQuery{Limit: 30}}
	var response ListValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func ListActiveValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{ListActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func SimulateActiveValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{SimulateActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

// TG4Query applies to all tg4 types - stake, group, and mixer
type TG4Query struct {
	Admin               *struct{}                 `json:"admin,omitempty"`
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

type TG4AdminResponse struct {
	Admin string `json:"admin,omitempty"`
}

// TG4MemberListResponse response to a list members query.
type TG4MemberListResponse struct {
	Members []TG4Member `json:"members"`
}

type TG4MemberResponse struct {
	// Weight nil means not a member, 0 means member with no voting power... this can be a very important distinction
	Weight *int `json:"weight"`
}

type TG4TotalWeightResponse struct {
	Weight int `json:"weight"`
}

func QueryTG4MembersByWeight(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress) ([]TG4Member, error) {
	query := TG4Query{ListMembersByWeight: &ListMembersByWeightQuery{Limit: 30}}
	var response TG4MemberListResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	return response.Members, err
}

func QueryTG4Members(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress) ([]TG4Member, error) {
	query := TG4Query{ListMembers: &ListMembersQuery{Limit: 30}}
	var response TG4MemberListResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	return response.Members, err
}

// QueryTG4Member returns the weight of this member. (nil, nil) means not present, (&0, nil) means member with no votes
func QueryTG4Member(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress, member sdk.AccAddress) (*int, error) {
	query := TG4Query{Member: &MemberQuery{Addr: member.String()}}
	var response TG4MemberResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	return response.Weight, err
}

// QueryTG4TotalWeight returns the weight of this member. (nil, nil) means not present
func QueryTG4TotalWeight(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress) (int, error) {
	query := TG4Query{TotalWeight: &struct{}{}}
	var response TG4TotalWeightResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	return response.Weight, err
}

// QueryTG4Admin returns admin of this contract, if any. Will return nil, err if no admin
func QueryTG4Admin(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress) (sdk.AccAddress, error) {
	query := TG4Query{Admin: &struct{}{}}
	var response TG4AdminResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	if err != nil {
		return nil, err
	}
	if response.Admin == "" {
		return nil, nil
	}
	return sdk.AccAddressFromBech32(response.Admin)
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
