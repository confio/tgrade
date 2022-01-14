package contract

import (
	"encoding/json"
	"sort"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type TG4Member struct {
	Addr   string `json:"addr"`
	Weight uint64 `json:"weight"`
}

func SortByWeightDesc(s []TG4Member) []TG4Member {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Weight > s[j].Weight || s[i].Weight == s[j].Weight && s[i].Addr < s[j].Addr
	})
	return s
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

func QueryTG4MembersByWeight(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress, pagination *Paginator) ([]TG4Member, error) {
	var sa TG4Member
	var startAfter *TG4Member
	var limit int
	if pagination != nil {
		if pagination.StartAfter != nil {
			if err := json.Unmarshal(pagination.StartAfter, &sa); err != nil {
				return nil, sdkerrors.Wrap(err, "query tg4 members by weight")
			}
			startAfter = &sa
		}
		limit = int(pagination.Limit)
	}
	query := TG4Query{ListMembersByWeight: &ListMembersByWeightQuery{StartAfter: startAfter, Limit: limit}}
	var response TG4MemberListResponse
	err := doQuery(ctx, k, tg4Addr, query, &response)
	return response.Members, err
}

func QueryTG4Members(ctx sdk.Context, k types.SmartQuerier, tg4Addr sdk.AccAddress, pagination *Paginator) ([]TG4Member, error) {
	var startAfter string
	var limit int
	if pagination != nil {
		startAfter = string(pagination.StartAfter)
		limit = int(pagination.Limit)
	}
	query := TG4Query{ListMembers: &ListMembersQuery{StartAfter: startAfter, Limit: limit}}
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

// TG4UpdateAdminMsg update admin message
type TG4UpdateAdminMsg struct {
	NewAdmin *string `json:"admin,omitempty"`
}
