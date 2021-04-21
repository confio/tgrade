package contracts

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

type CW4Member struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
}

type CW4InitMsg struct {
	Members []CW4Member `json:"members"`
}

func (m CW4InitMsg) Json(t *testing.T) string {
	t.Helper()
	r, err := json.Marshal(&m)
	require.NoError(t, err)
	return string(r)
}

type ValsetInitKey struct {
	Operator        string `json:"operator"`
	ValidatorPubkey string `json:"validator_pubkey"`
}
type ValsetInitMsg struct {
	Membership    string          `json:"membership"`
	MinWeight     int             `json:"min_weight"`
	MaxValidators int             `json:"max_validators"`
	EpochLength   int             `json:"epoch_length"`
	InitialKeys   []ValsetInitKey `json:"initial_keys"`
}

func (m ValsetInitMsg) Json(t *testing.T) string {
	t.Helper()
	r, err := json.Marshal(&m)
	require.NoError(t, err)
	return string(r)
}
