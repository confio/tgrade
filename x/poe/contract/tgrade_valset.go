package contract

import (
	"encoding/json"
	"testing"

	"github.com/confio/tgrade/x/poe/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// This is used by bootstrap module (what number = 1%)
const ValsetInitPercentageFactor = 10_000_000_000_000_000

// ValsetInitMsg Valset contract init message
// See https://github.com/confio/tgrade-contracts/tree/main/contracts/tgrade-valset
type ValsetInitMsg struct {
	Membership    string      `json:"membership"`
	MinWeight     uint64      `json:"min_weight"`
	MaxValidators uint32      `json:"max_validators"`
	EpochLength   uint64      `json:"epoch_length"`
	EpochReward   sdk.Coin    `json:"epoch_reward"`
	InitialKeys   []Validator `json:"initial_keys"`
	Scaling       uint32      `json:"scaling,omitempty"`
	// Percentage of total accumulated fees which is substracted from tokens minted as a rewards. A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0
	FeePercentage uint64 `json:"fee_percentage,string,omitempty"`
}

func (m ValsetInitMsg) Json(t *testing.T) string {
	return asJson(t, m)
}

type Validator struct {
	Operator        string          `json:"operator"`
	ValidatorPubkey ValidatorPubkey `json:"validator_pubkey"`
}

func asJson(t *testing.T, m interface{}) string {
	t.Helper()
	r, err := json.Marshal(&m)
	require.NoError(t, err)
	return string(r)
}

// TG4ValsetExecute Valset contract validator key registration
// See https://github.com/confio/tgrade-contracts/blob/main/contracts/tgrade-valset/schema/execute_msg.json
type TG4ValsetExecute struct {
	RegisterValidatorKey *RegisterValidatorKey `json:"register_validator_key,omitempty"`
	UpdateMetadata       *ValidatorMetadata    `json:"update_metadata,omitempty"`
}

type RegisterValidatorKey struct {
	PubKey   ValidatorPubkey   `json:"pubkey"`
	Metadata ValidatorMetadata `json:"metadata"`
}

type ValidatorMetadata struct {
	// moniker defines a human-readable name for the validator.
	Moniker string `json:"moniker"`
	// identity defines an optional identity signature (ex. UPort or Keybase).
	Identity string `json:"identity,omitempty"`
	// website defines an optional website link.
	Website string `json:"website,omitempty"`
	// security_contact defines an optional email for security contact.
	SecurityContact string `json:"security_contact,omitempty"`
	// details define other optional details.
	Details string `json:"details,omitempty"`
}

func MetadataFromDescription(description stakingtypes.Description) ValidatorMetadata {
	return ValidatorMetadata{
		Moniker:         description.Moniker,
		Identity:        description.Identity,
		Website:         description.Website,
		SecurityContact: description.SecurityContact,
		Details:         description.Details,
	}
}

func (m ValidatorMetadata) ToDescription() stakingtypes.Description {
	return stakingtypes.Description{
		Moniker:         m.Moniker,
		Identity:        m.Identity,
		Website:         m.Website,
		SecurityContact: m.SecurityContact,
		Details:         m.Details,
	}

}

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
	// Percentage of total accumulated fees which is substracted from tokens minted as a rewards. A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0
	FeePercentage uint64 `json:"fee_percentage,string,omitempty"`
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

func (v OperatorResponse) ToValidator() (stakingtypes.Validator, error) {
	pubKey, err := toCosmosPubKey(v.Pubkey)
	if err != nil {
		return stakingtypes.Validator{}, sdkerrors.Wrap(err, "convert to cosmos key")
	}
	any, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return stakingtypes.Validator{}, sdkerrors.Wrap(err, "convert to any type")
	}

	return stakingtypes.Validator{
		OperatorAddress: v.Operator,
		ConsensusPubkey: any,
		Description:     v.Metadata.ToDescription(),
		DelegatorShares: sdk.OneDec(),
		Status:          stakingtypes.Bonded,
	}, nil
}

func toCosmosPubKey(key ValidatorPubkey) (cryptotypes.PubKey, error) {
	switch {
	case key.Ed25519 != nil:
		return &ed25519.PubKey{Key: key.Ed25519}, nil
	case key.Secp256k1 != nil:
		return &cryptosecp256k1.PubKey{Key: key.Secp256k1}, nil
	default:
		return nil, types.ErrValidatorPubKeyTypeNotSupported
	}
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

func QueryValsetConfig(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) (*ValsetConfigResponse, error) {
	query := ValsetQuery{Config: &struct{}{}}
	var response ValsetConfigResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValsetEpoch(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) (*ValsetEpochResponse, error) {
	query := ValsetQuery{Epoch: &struct{}{}}
	var response ValsetEpochResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

func QueryValidator(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress, operator sdk.AccAddress) (*OperatorResponse, error) {
	query := ValsetQuery{Validator: &ValidatorQuery{Operator: operator.String()}}
	var response ValidatorResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validator, err
}

func ListValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]OperatorResponse, error) {
	query := ValsetQuery{ListValidators: &ListValidatorsQuery{Limit: 30}}
	var response ListValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func ListActiveValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{ListActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}

func SimulateActiveValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	query := ValsetQuery{SimulateActiveValidators: &struct{}{}}
	var response ListActiveValidatorsResponse
	err := doQuery(ctx, k, valset, query, &response)
	return response.Validators, err
}
