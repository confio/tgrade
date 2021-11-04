package contract

import (
	"testing"
)

// TG4EngagementInitMsg contract init message
// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagementInitMsg struct {
	Admin            string      `json:"admin,omitempty"`
	Members          []TG4Member `json:"members"`
	PreAuths         uint64      `json:"preauths,omitempty"`
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

// See https://github.com/confio/tgrade-contracts/blob/v0.5.0-alpha/contracts/tg4-engagement/src/msg.rs
type TG4EngagmentExecute struct {
	UpdateMembers *UpdateMembersMsg `json:"update_members,omitempty"`
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
	msg := TG4EngagmentExecute{
		UpdateMembers: m,
	}
	return asJson(t, msg)
}
