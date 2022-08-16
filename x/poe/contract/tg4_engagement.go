package contract

import (
	"encoding/json"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

// TG4EngagementInitMsg contract init message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementInitMsg struct {
	Admin            string      `json:"admin,omitempty"`
	Members          []TG4Member `json:"members"`
	PreAuthsHooks    uint64      `json:"preauths_hooks,omitempty"`
	PreAuthsSlashing uint64      `json:"preauths_slashing,omitempty"`
	// Halflife is measured in seconds
	Halflife uint64 `json:"halflife,omitempty"`
	// Denom of tokens which may be distributed by this contract.
	Denom string `json:"denom"`
}

// TG4EngagementSudoMsg TG4 group sudo message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementSudoMsg struct {
	UpdateMember *TG4Member `json:"update_member,omitempty"`
}

// TG4EngagementExecute execute message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementExecute struct {
	UpdateMembers      *UpdateMembersMsg      `json:"update_members,omitempty"`
	UpdateAdmin        *TG4UpdateAdminMsg     `json:"update_admin,omitempty"`
	WithdrawRewards    *WithdrawRewardsMsg    `json:"withdraw_rewards,omitempty"`
	DelegateWithdrawal *DelegateWithdrawalMsg `json:"delegate_withdrawal,omitempty"`
}

// UpdateMembersMsg contract execute message to update members
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type UpdateMembersMsg struct {
	Add    []TG4Member `json:"add"`
	Remove []string    `json:"remove"`
}

// WithdrawRewardsMsg contract execute message to claim rewards
type WithdrawRewardsMsg struct {
	Owner    *string `json:"owner,omitempty"`
	Receiver *string `json:"receiver,omitempty"`
}

// DelegateWithdrawalMsg contract sets given address as allowed for senders funds withdrawal
type DelegateWithdrawalMsg struct {
	Delegated string `json:"delegated"`
}

func (m *UpdateMembersMsg) ToJSON(t *testing.T) string {
	t.Helper()
	switch {
	case m.Add == nil:
		m.Add = make([]TG4Member, 0)
	case m.Remove == nil:
		m.Remove = make([]string, 0)
	}
	msg := TG4EngagementExecute{
		UpdateMembers: m,
	}
	r, err := json.Marshal(msg)
	require.NoError(t, err)
	return string(r)
}

type EngagementContractAdapter struct {
	BaseContractAdapter
}

// NewEngagementContractAdapter constructor
func NewEngagementContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *EngagementContractAdapter {
	return &EngagementContractAdapter{
		BaseContractAdapter: NewBaseContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		),
	}
}

func (a EngagementContractAdapter) UpdateAdmin(ctx sdk.Context, newAdmin, sender sdk.AccAddress) error {
	bech32AdminAddr := newAdmin.String()
	msg := TG4EngagementExecute{
		UpdateAdmin: &TG4UpdateAdminMsg{NewAdmin: &bech32AdminAddr},
	}
	return a.doExecute(ctx, msg, sender)
}

// EngagementQuery will create many queries for the engagement contract
// See https://github.com/confio/poe-contracts/blob/v0.5.3-2/contracts/tg4-engagement/src/msg.rs#L77-L123
type EngagementQuery struct {
	Admin                  *struct{}                    `json:"admin,omitempty"`
	TotalPoints            *struct{}                    `json:"total_points,omitempty"`
	ListMembers            *ListMembersQuery            `json:"list_members,omitempty"`
	ListMembersByPoints    *ListMembersByPointsQuery    `json:"list_members_by_points,omitempty"`
	Member                 *MemberQuery                 `json:"member,omitempty"`
	Hooks                  *struct{}                    `json:"hooks,omitempty"`
	Preauths               *struct{}                    `json:"preauths,omitempty"`
	WithdrawableRewards    *WithdrawableRewardsQuery    `json:"withdrawable_rewards,omitempty"`
	DistributedRewards     *struct{}                    `json:"distributed_rewards,omitempty"`
	UndistributedRewards   *struct{}                    `json:"undistributed_rewards,omitempty"`
	Delegated              *DelegatedQuery              `json:"delegated,omitempty"`
	Halflife               *struct{}                    `json:"halflife,omitempty"`
	IsSlasher              *IsSlasher                   `json:"is_slasher,omitempty"`
	ListSlashers           *struct{}                    `json:"list_slashers,omitempty"`
	DistributionData       *struct{}                    `json:"distribution_data,omitempty"`
	WithdrawAdjustmentData *WithdrawAdjustmentDataQuery `json:"withdraw_adjustment_data,omitempty"`
}

type DelegatedQuery struct {
	Owner string `json:"owner"`
}

type IsSlasher struct {
	Addr string `json:"addr"`
}

type WithdrawAdjustmentDataQuery struct {
	Addr string `json:"addr"`
}

type DelegatedResponse struct {
	Delegated string `json:"delegated"`
}

type SlasherResponse struct {
	IsSlasher bool `json:"is_slasher"`
}

type WithdrawAdjustmentResponse struct {
	// PointsCorrection is int128 encoded as a string (use sdk.Int?)
	PointsCorrection string `json:"points_correction"`
	// WithdrawnFunds is uint128 encoded as a string (use sdk.Int?)
	WithdrawnFunds string `json:"withdrawn_funds"`
	Delegated      string `json:"delegated"`
}

func (a EngagementContractAdapter) QueryDelegated(ctx sdk.Context, ownerAddr sdk.AccAddress) (*DelegatedResponse, error) {
	query := EngagementQuery{Delegated: &DelegatedQuery{Owner: ownerAddr.String()}}
	var rsp DelegatedResponse
	err := a.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	return &rsp, err
}

func (a EngagementContractAdapter) QueryWithdrawableRewards(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coin, error) {
	query := EngagementQuery{WithdrawableRewards: &WithdrawableRewardsQuery{Owner: addr.String()}}
	var resp RewardsResponse
	err := a.doQuery(ctx, query, &resp)
	if err != nil {
		return sdk.Coin{}, err
	}
	return resp.Rewards, err
}
