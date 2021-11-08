package contract

import (
	"testing"

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
	Token    string `json:"token"`
}

func (m TG4EngagementInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

// TG4EngagementSudoMsg TG4 group sudo message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementSudoMsg struct {
	UpdateMember *TG4Member `json:"update_member,omitempty"`
}

// TG4EngagementExecute execute message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementExecute struct {
	UpdateMembers *UpdateMembersMsg  `json:"update_members,omitempty"`
	UpdateAdmin   *TG4UpdateAdminMsg `json:"update_admin,omitempty"`
}

// UpdateMembersMsg contract execute message to update members
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type UpdateMembersMsg struct {
	Add    []TG4Member `json:"add"`
	Remove []string    `json:"remove"`
}

func (m *UpdateMembersMsg) Json(t *testing.T) string {
	switch {
	case m.Add == nil:
		m.Add = make([]TG4Member, 0)
	case m.Remove == nil:
		m.Remove = make([]string, 0)
	}
	msg := TG4EngagementExecute{
		UpdateMembers: m,
	}
	return asJson(t, msg)
}

type EngagementContractAdapter struct {
	ContractAdapter
}

// NewEngagementContractAdapter constructor
func NewEngagementContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *EngagementContractAdapter {
	return &EngagementContractAdapter{
		ContractAdapter: NewContractAdapter(
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
