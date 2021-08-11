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

func QueryValsetConfig(ctx sdk.Context, k *twasmkeeper.Keeper, valsetAddr sdk.AccAddress) (*ValsetConfigResponse, error) {
	query := ValsetQuery{Config: &struct{}{}}
	bz, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	res, err := k.QuerySmart(ctx, valsetAddr, bz)
	if err != nil {
		return nil, err
	}
	var gotValsetConfig ValsetConfigResponse
	err = json.Unmarshal(res, &gotValsetConfig)
	return &gotValsetConfig, err
}

// QueryTG4GroupResponse response to a list members query.
type QueryTG4GroupResponse struct {
	Members []TG4Member `json:"members"`
}

// TODO: add pagination
func QueryGroupMembers(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress) ([]TG4Member, error) {
	res, err := k.QuerySmart(ctx, groupAddr, []byte(`{"list_members_by_weight":{"limit":30}}`))
	if err != nil {
		return nil, err
	}
	var gotMemberResponse QueryTG4GroupResponse
	err = json.Unmarshal(res, &gotMemberResponse)
	return gotMemberResponse.Members, err
}
