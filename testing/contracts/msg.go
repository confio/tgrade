package contracts

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

// CW4InitMsg contract init message
//See https://github.com/CosmWasm/cosmwasm-plus/tree/main/contracts/cw4-group
type CW4InitMsg struct {
	Admin   string      `json:"admin,omitempty"`
	Members []CW4Member `json:"members"`
}

func (m CW4InitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

type CW4Member struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

func SortByWeight(s []CW4Member) []CW4Member {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Weight > s[j].Weight || s[i].Weight == s[j].Weight && s[i].Addr > s[j].Addr
	})
	return s
}

// CW4UpdateMembersMsg contract execute message to update members
//See https://github.com/CosmWasm/cosmwasm-plus/tree/main/contracts/cw4-group
type CW4UpdateMembersMsg struct {
	Add    []CW4Member `json:"add"`
	Remove []string    `json:"remove"`
}

func (m *CW4UpdateMembersMsg) Json(t *testing.T) string {
	switch {
	case m.Add == nil:
		m.Add = make([]CW4Member, 0)
	case m.Remove == nil:
		m.Remove = make([]string, 0)
	}
	x := map[string]interface{}{
		"update_members": m,
	}
	return asJson(t, x)
}

// ValsetInitMsg Valset contract init message
// See https://github.com/confio/tgrade-contracts/tree/main/contracts/tgrade-valset
type ValsetInitMsg struct {
	Membership    string          `json:"membership"`
	MinWeight     int             `json:"min_weight"`
	MaxValidators int             `json:"max_validators"`
	EpochLength   int             `json:"epoch_length"`
	InitialKeys   []ValsetInitKey `json:"initial_keys"`
	Scaling       int             `json:"scaling,omitempty"`
}

func (m ValsetInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

type ValsetInitKey struct {
	Operator        string          `json:"operator"`
	ValidatorPubkey ValidatorPubkey `json:"validator_pubkey"`
}

func NewValsetInitKey(operator, ed25519Pubkey string) ValsetInitKey {
	return ValsetInitKey{Operator: operator, ValidatorPubkey: ValidatorPubkey{Ed25519: ed25519Pubkey}}
}

type ValidatorPubkey struct {
	Ed25519 string `json:"ed25519,omitempty"`
}

func asJson(t *testing.T, m interface{}) string {
	t.Helper()
	r, err := json.Marshal(&m)
	require.NoError(t, err)
	return string(r)
}
