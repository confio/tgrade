package contract

import (
	"testing"
)

// TG4GroupInitMsg contract init message
//See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-group/schema/instantiate_msg.json
type TG4GroupInitMsg struct {
	Admin    string      `json:"admin,omitempty"`
	Members  []TG4Member `json:"members"`
	Preauths uint64      `json:"preauths,omitempty"`
}

func (m TG4GroupInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

// TG4GroupSudoMsg TG4 group sudo message
// See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-group/schema/sudo_msg.json
type TG4GroupSudoMsg struct {
	UpdateMember *TG4Member `json:"update_member,omitempty"`
}

// TG4UpdateMembersMsg contract execute message to update members
// See https://github.com/CosmWasm/cosmwasm-plus/tree/main/contracts/cw4-group
// https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-group/schema/execute_msg.json
type TG4UpdateMembersMsg struct {
	Add    []TG4Member `json:"add"`
	Remove []string    `json:"remove"`
}

func (m *TG4UpdateMembersMsg) Json(t *testing.T) string {
	switch {
	case m.Add == nil:
		m.Add = make([]TG4Member, 0)
	case m.Remove == nil:
		m.Remove = make([]string, 0)
	}
	x := map[string]interface{}{
		"update_members": m,
	}
	return asJson(t, x)
}
