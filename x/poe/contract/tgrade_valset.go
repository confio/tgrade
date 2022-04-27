package contract

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptosecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/types"
)

func DecimalFromPercentage(percent sdk.Dec) *sdk.Dec {
	if percent.IsZero() {
		return nil
	}
	res := percent.QuoInt64(100)
	return &res
}

func DecimalFromProMille(promille int64) *sdk.Dec {
	res := sdk.NewDec(promille).QuoInt64(1000)
	return &res
}

// ValsetInitMsg Valset contract init message
// See https://github.com/confio/tgrade-contracts/tree/v0.5.0-alpha/contracts/tgrade-valset/src/msg.rs
type ValsetInitMsg struct {
	Admin         string      `json:"admin,omitempty"`
	Membership    string      `json:"membership"`
	MinPoints     uint64      `json:"min_points"`
	MaxValidators uint32      `json:"max_validators"`
	EpochLength   uint64      `json:"epoch_length"`
	EpochReward   sdk.Coin    `json:"epoch_reward"`
	InitialKeys   []Validator `json:"initial_keys"`
	Scaling       uint32      `json:"scaling,omitempty"`
	// Percentage of total accumulated fees which is subtracted from tokens minted as a rewards. A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0
	FeePercentage *sdk.Dec `json:"fee_percentage,omitempty"`
	// If set to true, we will auto-unjail any validator after their jailtime is over.
	AutoUnjail bool `json:"auto_unjail"`
	// This contract receives the rewards that don't go to the validator (set ot tg4-engagement)
	DistributionContracts []DistributionContract `json:"distribution_contracts,omitempty"`
	// This is the code-id of the cw2222-compliant contract used to handle rewards for the validators.
	// Generally, it should the tg4-engagement code id.
	ValidatorGroupCodeID uint64 `json:"validator_group_code_id"`
	// When a validator joins the valset, verify they sign the first block, or jail them for a period otherwise.
	// The verification happens every time the validator becomes an active validator, including when they are unjailed
	// or when they just gain enough power to participate.
	VerifyValidators bool `json:"verify_validators"`
	// The duration in seconds to jail a validator for in case they don't sign their first epoch boundary block.
	// After the period, they have to pass verification again, ad infinitum.
	OfflineJailDuration uint64 `json:"offline_jail_duration"`
}

type DistributionContract struct {
	Address string `json:"contract"`
	// Ratio of total reward tokens for an epoch to be sent to that contract for further distribution.
	// Range 0 - 1
	Ratio sdk.Dec `json:"ratio"`
}

type Validator struct {
	Operator        string          `json:"operator"`
	ValidatorPubkey ValidatorPubkey `json:"validator_pubkey"`
}

// TG4ValsetExecute Valset contract validator key registration
// See https://github.com/confio/tgrade-contracts/tree/v0.5.0-alpha/contracts/tgrade-valset/src/msg.rs
type TG4ValsetExecute struct {
	RegisterValidatorKey *RegisterValidatorKey `json:"register_validator_key,omitempty"`
	UpdateMetadata       *ValidatorMetadata    `json:"update_metadata,omitempty"`
	// Jails validator. Can be executed only by the admin.
	Jail *JailMsg `json:"jail,omitempty"`
	// Unjails validator. Admin can unjail anyone anytime, others can unjail only themselves and
	// only if the jail period passed.
	Unjail *UnjailMsg `json:"unjail,omitempty"`
	// UpdateAdmin set a new admin address
	UpdateAdmin  *TG4UpdateAdminMsg `json:"update_admin,omitempty"`
	UpdateConfig *UpdateConfigMsg   `json:"update_config,omitempty"`
}

type UpdateConfigMsg struct {
	MinPoints     uint64 `json:"min_points,omitempty"`
	MaxValidators uint32 `json:"max_validators,omitempty"`
}

type JailMsg struct {
	Operator string `json:"operator"`
	// Duration for how long validator is jailed (in seconds)
	Duration JailingDuration `json:"duration"`
}

type JailingDuration struct {
	Duration uint64    `json:"duration,omitempty"`
	Forever  *struct{} `json:"forever,omitempty"`
}

type UnjailMsg struct {
	// Address to unjail. Optional, as if not provided it is assumed to be the sender of the
	// message (for convenience when unjailing self after the jail period).
	Operator string `json:"operator,omitempty"`
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
// See https://github.com/confio/tgrade-contracts/tree/v0.5.0-alpha/contracts/tgrade-valset/src/msg.rs
type ValsetQuery struct {
	Config                   *struct{}            `json:"configuration,omitempty"`
	Epoch                    *struct{}            `json:"epoch,omitempty"`
	Validator                *ValidatorQuery      `json:"validator,omitempty"`
	ListValidators           *ListValidatorsQuery `json:"list_validators,omitempty"`
	ListActiveValidators     *ListValidatorsQuery `json:"list_active_validators,omitempty"`
	SimulateActiveValidators *struct{}            `json:"simulate_active_validators,omitempty"`
	ListValidatorSlashing    *ValidatorQuery      `json:"list_validator_slashing,omitempty"`
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
	Membership    string   `json:"membership"`
	MinPoints     uint64   `json:"min_points"`
	MaxValidators uint32   `json:"max_validators"`
	Scaling       uint32   `json:"scaling,omitempty"`
	EpochReward   sdk.Coin `json:"epoch_reward"`
	// Percentage of total accumulated fees which is subtracted from tokens minted as a rewards. A fixed-point decimal value with 18 fractional digits, i.e. Decimal(1_000_000_000_000_000_000) == 1.0
	FeePercentage         sdk.Dec                `json:"fee_percentage"`
	DistributionContracts []DistributionContract `json:"distribution_contracts,omitempty"`
	ValidatorGroup        string                 `json:"validator_group"`
	AutoUnjail            bool                   `json:"auto_unjail"`
}

// ValsetEpochQueryResponse Response to `config` query
type ValsetEpochResponse struct {
	// Number of seconds in one epoch. We update the Tendermint validator set only once per epoch.
	EpochLength uint64 `json:"epoch_length"`
	// The current epoch # (block.time/epoch_length, rounding down)
	CurrentEpoch uint64 `json:"current_epoch"`
	// The last time we updated the validator set - block time (in seconds)
	LastUpdateTime uint64 `json:"last_update_time"`
	// The last time we updated the validator set - block height
	LastUpdateHeight uint64 `json:"last_update_height"`
	// TODO: add this if you want it, not in current code
	// Seconds (UTC UNIX time) of next timestamp that will trigger a validator recalculation
	//NextUpdateTime int `json:"next_update_time"`
}

type OperatorResponse struct {
	Operator        string            `json:"operator"`
	Pubkey          ValidatorPubkey   `json:"pubkey"`
	Metadata        ValidatorMetadata `json:"metadata"`
	JailedUntil     *JailingPeriod    `json:"jailed_until,omitempty"`
	ActiveValidator bool              `json:"active_validator"`
}

type JailingPeriod struct {
	Start time.Time  `json:"start,omitempty"`
	End   JailingEnd `json:"end,omitempty"`
}

type JailingEnd struct {
	Forever bool      `json:"forever,omitempty"`
	Until   time.Time `json:"until,omitempty"`
}

func (j *JailingPeriod) UnmarshalJSON(data []byte) error {
	var r struct {
		Start string `json:"start,omitempty"`
		End   struct {
			Until   string    `json:"until,omitempty"`
			Forever *struct{} `json:"forever,omitempty"`
		} `json:"end,omitempty"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	if r.Start != "" {
		start, err := strconv.ParseInt(r.Start, 10, 64)
		if err != nil {
			return err
		}
		j.Start = time.Unix(0, start).UTC()
	}

	switch {
	case r.End.Forever != nil:
		j.End.Forever = true
	case r.End.Until != "":
		until, err := strconv.ParseInt(r.End.Until, 10, 64)
		if err != nil {
			return err
		}
		j.End.Until = time.Unix(0, until).UTC()
	default:
		return errors.New("unknown json data: ")
	}
	return nil
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

	status := stakingtypes.Bonded
	if !v.ActiveValidator {
		status = stakingtypes.Unbonded
	}

	return stakingtypes.Validator{
		OperatorAddress: v.Operator,
		ConsensusPubkey: any,
		Description:     v.Metadata.ToDescription(),
		DelegatorShares: sdk.OneDec(),
		Status:          status,
		Jailed:          v.JailedUntil != nil,
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
	Power           uint64          `json:"power"`
}

type ValidatorResponse struct {
	Validator *OperatorResponse `json:"validator"`
}

type ListValidatorsResponse struct {
	Validators []OperatorResponse `json:"validators"`
}

func (l ListValidatorsResponse) PaginationCursor() PaginationCursor {
	if len(l.Validators) == 0 {
		return nil
	}
	return PaginationCursor(l.Validators[len(l.Validators)-1].Operator)
}

type ListActiveValidatorsResponse struct {
	Validators []ValidatorInfo `json:"validators"`
}

type ListValidatorSlashingResponse struct {
	Operator    string              `json:"operator"`
	StartHeight uint64              `json:"start_height"`
	Slashing    []ValidatorSlashing `json:"slashing"`
}
type ValidatorSlashing struct {
	Height  uint64  `json:"slash_height"`
	Portion sdk.Dec `json:"portion"`
}

type SimulateActiveValidatorsResponse = ListActiveValidatorsResponse

func QueryValsetEpoch(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) (*ValsetEpochResponse, error) {
	query := ValsetQuery{Epoch: &struct{}{}}
	var response ValsetEpochResponse
	err := doQuery(ctx, k, valset, query, &response)
	return &response, err
}

// TODO: add pagination support
func ListActiveValidators(ctx sdk.Context, k types.SmartQuerier, valset sdk.AccAddress) ([]ValidatorInfo, error) {
	// TODO: this is just a placeholder trying to get 100
	query := ValsetQuery{ListActiveValidators: &ListValidatorsQuery{Limit: 100}}
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

type ValsetContractAdapter struct {
	ContractAdapter
}

// NewValsetContractAdapter constructor
func NewValsetContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *ValsetContractAdapter {
	return &ValsetContractAdapter{
		ContractAdapter: NewContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		)}
}

// QueryValidator query a single validator and map to the sdk type. returns nil when not found
func (v ValsetContractAdapter) QueryValidator(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
	query := ValsetQuery{Validator: &ValidatorQuery{Operator: opAddr.String()}}
	var rsp ValidatorResponse
	err := v.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	if rsp.Validator == nil {
		return nil, nil
	}
	val, err := rsp.Validator.ToValidator()
	return &val, err
}

// QueryRawValidator query a single validator as the contract returns it. returns nil when not found
func (v ValsetContractAdapter) QueryRawValidator(ctx sdk.Context, opAddr sdk.AccAddress) (ValidatorResponse, error) {
	query := ValsetQuery{Validator: &ValidatorQuery{Operator: opAddr.String()}}
	var rsp ValidatorResponse
	err := v.doQuery(ctx, query, &rsp)
	return rsp, sdkerrors.Wrap(err, "contract query")
}

// ListValidators query all validators
func (v ValsetContractAdapter) ListValidators(ctx sdk.Context, pagination *Paginator) ([]stakingtypes.Validator, PaginationCursor, error) {
	var startAfter string
	var limit int
	if pagination != nil {
		startAfter = string(pagination.StartAfter)
		limit = int(pagination.Limit)
	}
	query := ValsetQuery{ListValidators: &ListValidatorsQuery{StartAfter: startAfter, Limit: limit}}
	var rsp ListValidatorsResponse
	cursor, err := v.doPageableQuery(ctx, query, &rsp)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "contract query")
	}
	vals := make([]stakingtypes.Validator, len(rsp.Validators))
	for i, v := range rsp.Validators {
		vals[i], err = v.ToValidator()
		if err != nil {
			return nil, nil, err
		}
	}
	// always return the cursor and let the client figure out if they want to do another call
	// a simple len(res.validators) < limit check to clear the cursor would not work because
	// the contract also has a max limit that may be < our limit
	return vals, cursor, nil
}

func (v ValsetContractAdapter) ListValidatorSlashing(ctx sdk.Context, opAddr sdk.AccAddress) ([]ValidatorSlashing, error) {
	query := ValsetQuery{ListValidatorSlashing: &ValidatorQuery{Operator: opAddr.String()}}
	var rsp ListValidatorSlashingResponse
	err := v.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	return rsp.Slashing, nil
}

// QueryConfig query contract configuration
func (v ValsetContractAdapter) QueryConfig(ctx sdk.Context) (*ValsetConfigResponse, error) {
	if v.addressLookupErr != nil {
		return nil, v.addressLookupErr
	}
	query := ValsetQuery{Config: &struct{}{}}
	var rsp ValsetConfigResponse
	err := v.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	return &rsp, err
}

// UpdateAdmin sets a new admin address
func (v ValsetContractAdapter) UpdateAdmin(ctx sdk.Context, newAdmin sdk.AccAddress, sender sdk.AccAddress) error {
	bech32AdminAddr := newAdmin.String()
	msg := TG4ValsetExecute{
		UpdateAdmin: &TG4UpdateAdminMsg{NewAdmin: &bech32AdminAddr},
	}
	return v.doExecute(ctx, msg, sender)
}

// JailValidator is for testing propose only. On a chain the OC does this
func (v ValsetContractAdapter) JailValidator(ctx sdk.Context, nodeOperator sdk.AccAddress, duration time.Duration, forever bool, sender sdk.AccAddress) error {
	if time.Duration(int64(duration.Seconds()))*time.Second != duration {
		return sdkerrors.Wrap(types.ErrInvalid, "must fit into seconds")
	}
	var jailDuration JailingDuration
	switch {
	case forever && duration > time.Nanosecond:
		return sdkerrors.Wrap(types.ErrInvalid, "either duration or forever")
	case forever:
		jailDuration = JailingDuration{Forever: &struct{}{}}
	case duration > time.Nanosecond:
		jailDuration = JailingDuration{Duration: uint64(duration.Seconds())}
	default:
		return types.ErrEmpty
	}
	msg := TG4ValsetExecute{
		Jail: &JailMsg{
			Operator: nodeOperator.String(),
			Duration: jailDuration,
		},
	}
	return v.doExecute(ctx, msg, sender)
}

func (v ValsetContractAdapter) UnjailValidator(ctx sdk.Context, sender sdk.AccAddress) error {
	msg := TG4ValsetExecute{
		Unjail: &UnjailMsg{},
	}
	return v.doExecute(ctx, msg, sender)
}
