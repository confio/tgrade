package contract

// TgradeSudoMsg callback message sent to a contract.
// See https://github.com/confio/tgrade-contracts/blob/main/packages/bindings/src/sudo.rs
type TgradeSudoMsg struct {
	PrivilegeChange *PrivilegeChangeMsg `json:"privilege_change,omitempty"`

	BeginBlock *BeginBlock `json:"begin_block,omitempty"`
	// This will be delivered every block if the contract is currently registered for End Block
	/// Block height and time is already in Env
	EndBlock *struct{} `json:"end_block,omitempty"`
	/// This will be delivered after everything.
	/// The data in the Response is (JSON?) encoded diff to the validator set
	EndWithValidatorUpdate *struct{} `json:"end_with_validator_update,omitempty"`
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

const EvidenceDuplicateVote EvidenceType = "DuplicateVote"
const EvidenceLightClientAttack EvidenceType = "LightClientAttack"

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
