package contract

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TgradeSudoMsg callback message sent to a contract.
// See https://github.com/confio/tgrade-contracts/blob/main/packages/bindings/src/sudo.rs
type TgradeSudoMsg struct {
	PrivilegeChange *PrivilegeChangeMsg `json:"privilege_change,omitempty"`

	BeginBlock *BeginBlock `json:"begin_block,omitempty"`
	// This will be delivered every block if the contract is currently registered for End Block
	// Block height and time is already in Env
	EndBlock *struct{} `json:"end_block,omitempty"`
	// This will be delivered after everything.
	// The data in the Response is (JSON?) encoded diff to the validator set
	EndWithValidatorUpdate *struct{} `json:"end_with_validator_update,omitempty"`

	// Export dump state for genesis export
	Export *struct{} `json:"export,omitempty"`
	// Import genesis state
	Import *ValsetState `json:"import,omitempty"`
}

// PrivilegeChangeMsg is called on a contract when it is made privileged or demoted
type PrivilegeChangeMsg struct {
	/// This is called when a contract gets "privileged status".
	/// It is a proper place to call `RegisterXXX` methods that require this status.
	/// Contracts that require this should be in a "frozen" state until they get this callback.
	Promoted *struct{} `json:"promoted,omitempty"`
	/// This is called when a contract looses "privileged status"
	Demoted *struct{} `json:"demoted,omitempty"`
}

// BeginBlock is delivered every block if the contract is currently registered for Begin Block
type BeginBlock struct {
	Evidence []Evidence `json:"evidence"` // This is key for slashing - let's figure out a standard for these types
}

type EvidenceType string

const (
	EvidenceDuplicateVote     EvidenceType = "duplicate_vote"
	EvidenceLightClientAttack EvidenceType = "light_client_attack"
)

// Evidence See https://github.com/tendermint/tendermint/blob/v0.34.8/proto/tendermint/abci/types.proto#L354-L375
type Evidence struct {
	EvidenceType EvidenceType `json:"evidence_type"`
	Validator    Validator    `json:"validator"`
	Height       uint64       `json:"height"`
	// the time when the offense occurred (in nanosec UNIX time, like env.block.time)
	Time             uint64 `json:"time"`
	TotalVotingPower uint64 `json:"total_voting_power"`
}

// Validator See https://github.com/tendermint/tendermint/blob/v0.34.8/proto/tendermint/abci/types.proto#L336-L340
type Validator struct {
	// The first 20 bytes of SHA256(public key)
	Address []byte `json:"address"`
	Power   uint64 `json:"power"`
}

type ValsetState struct {
	ContractVersion       ContractVersion       `json:"contract_version"`
	Config                ValsetConfigResponse  `json:"config"`
	Epoch                 EpochInfo             `json:"epoch"`
	Operators             []OperatorResponse    `json:"operators"`
	Validators            []ValidatorInfo       `json:"validators"`
	ValidatorsStartHeight []StartHeightResponse `json:"validators_start_height"`
	ValidatorsSlashing    []SlashingResponse    `json:"validators_slashing"`
}

type ContractVersion struct {
	Contract string `json:"contract"`
	Version  string `json:"version"`
}

type EpochInfo struct {
	EpochLength      uint64 `json:"epoch_length"`
	CurrentEpoch     uint64 `json:"current_epoch"`
	LastUpdateTime   uint64 `json:"last_update_time"`
	LastUpdateHeight uint64 `json:"last_update_height"`
}

type OperatorResponse struct {
	Operator        string            `json:"operator"`
	Pubkey          ValidatorPubkey   `json:"pubkey"`
	Metadata        ValidatorMetadata `json:"metadata"`
	JailedUntil     *JailingPeriod    `json:"jailed_until,omitempty"`
	ActiveValidator bool              `json:"active_validator"`
}

type StartHeightResponse struct {
	Validator string `json:"validator"`
	Height    uint64 `json:"height"`
}

type SlashingResponse struct {
	Validator string              `json:"validator"`
	Slashing  []ValidatorSlashing `json:"slashing"`
}

type ValidatorSlashing struct {
	Height  uint64  `json:"slash_height"`
	Portion sdk.Dec `json:"portion"`
}

type JailingPeriod struct {
	Start time.Time  `json:"start,omitempty"`
	End   JailingEnd `json:"end,omitempty"`
}

type JailingEnd struct {
	Forever bool      `json:"forever,omitempty"`
	Until   time.Time `json:"until,omitempty"`
}

type ValidatorPubkey struct {
	Ed25519   []byte `json:"ed25519,omitempty"`
	Secp256k1 []byte `json:"secp256k1,omitempty"`
	Sr25519   []byte `json:"sr25519,omitempty"`
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
	DoubleSignSlashRatio  sdk.Dec                `json:"double_sign_slash_ratio"`
}

type DistributionContract struct {
	Address string `json:"contract"`
	// Ratio of total reward tokens for an epoch to be sent to that contract for further distribution.
	// Range 0 - 1
	Ratio sdk.Dec `json:"ratio"`
}

type ValidatorInfo struct {
	Operator        string          `json:"operator"`
	ValidatorPubkey ValidatorPubkey `json:"validator_pubkey"`
	Power           uint64          `json:"power"`
}
