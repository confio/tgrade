package contract

import (
	"encoding/json"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValsetConfigQueryResponse Response to `config` query
// https://github.com/confio/tgrade-contracts/blob/v0.1.5/contracts/tgrade-valset/schema/query_msg.json#L3
type ValsetConfigQueryResponse struct {
	Membership    string `json:"membership"`
	MinWeight     int    `json:"min_weight"`
	MaxValidators int    `json:"max_validators"`
	Scaling       int    `json:"scaling,omitempty"`
}

func QueryValsetConfig(ctx sdk.Context, k *twasmkeeper.Keeper, valsetAddr sdk.AccAddress) (*ValsetConfigQueryResponse, error) {
	res, err := k.QuerySmart(ctx, valsetAddr, []byte(`{"config":{}}`))
	if err != nil {
		return nil, err
	}
	var gotValsetConfig ValsetConfigQueryResponse
	err = json.Unmarshal(res, &gotValsetConfig)
	return &gotValsetConfig, err
}

// QueryTG4GroupResponse response to a list members query.
type QueryTG4GroupResponse struct {
	Members []TG4Member `json:"members"`
}

func QueryGroupMembers(ctx sdk.Context, k *twasmkeeper.Keeper, groupAddr sdk.AccAddress) ([]TG4Member, error) {
	res, err := k.QuerySmart(ctx, groupAddr, []byte(`{"list_members_by_weight":{"limit":30}}`))
	if err != nil {
		return nil, err
	}
	var gotMemberResponse QueryTG4GroupResponse
	err = json.Unmarshal(res, &gotMemberResponse)
	return gotMemberResponse.Members, err
}
