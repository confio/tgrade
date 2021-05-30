package contracts

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"sort"
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

type TG4Member struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

func SortByWeight(s []TG4Member) []TG4Member {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Weight > s[j].Weight || s[i].Weight == s[j].Weight && s[i].Addr > s[j].Addr
	})
	return s
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

// TG4MixerInitMsg contract init message
//See https://github.com/confio/tgrade-contracts/blob/main/contracts/tg4-mixer/schema/instantiate_msg.json
type TG4MixerInitMsg struct {
	Admin      string `json:"admin,omitempty"`
	LeftGroup  string `json:"left_group"`
	RightGroup string `json:"right_group"`
	Preauths   uint64 `json:"preauths,omitempty"`
}

func (m TG4MixerInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

type TG4StakeInitMsg struct {
	Admin           string         `json:"admin,omitempty"`
	Denom           Denom          `json:"denom"`
	MinBond         string         `json:"min_bond"`
	TokensPerWeight string         `json:"tokens_per_weight"`
	UnbondingPeriod UnbodingPeriod `json:"unbonding_period"`
	Preauths        uint64         `json:"preauths,omitempty"`
}

func (m TG4StakeInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

type Denom struct {
	Native string `json:"native,omitempty"`
	CW20   string `json:"cw20,omitempty"`
}

type UnbodingPeriod struct {
	Height    uint64 `json:"height,omitempty"`
	TimeInSec uint64 `json:"time,omitempty"`
}

type TG4StakeExecute struct {
	Bond *struct{} `json:"bond,omitempty"`
}

func (m TG4StakeExecute) Json(t *testing.T) string {
	return asJson(t, m)
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
