package contract

import "github.com/confio/tgrade/x/twasm/contract"

// ValidatorVotingInitMsg setup contract on instantiation
type ValidatorVotingInitMsg struct {
	VotingRules  VotingRules `json:"rules"`
	GroupAddress string      `json:"group_addr"`
}

// ValidatorVotingExecuteMsg executable contract message
type ValidatorVotingExecuteMsg struct {
	Propose *ValidatorVotingPropose `json:"propose,omitempty"`
	Vote    *struct{}               `json:"vote,omitempty"`
	Execute *struct{}               `json:"execute,omitempty"`
	// Close   *struct{} `json:"close,omitempty"`
}

// ValidatorVotingPropose submit a new gov proposal
type ValidatorVotingPropose struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Proposal    ValidatorProposal `json:"proposal"`
}

// ValidatorProposal proposal options.
type ValidatorProposal struct {
	RegisterUpgrade *ChainUpgrade `json:"register_upgrade,omitempty"`
	CancelUpgrade   *struct{}     `json:"cancel_upgrade,omitempty"`
	// PinCodes that should be pinned in cache for high performance
	PinCodes []uint64 `json:"pin_codes,omitempty"`
	/// UnpinCodes that should be removed from cache to free space
	UnpinCodes                    []uint64                       `json:"unpin_codes,omitempty"`
	UpdateConsensusBlockParams    *ConsensusBlockParamsUpdate    `json:"update_consensus_block_params,omitempty"`
	UpdateConsensusEvidenceParams *ConsensusEvidenceParamsUpdate `json:"update_consensus_evidence_params,omitempty"`
	MigrateContract               *Migration                     `json:"migrate_contract,omitempty"`
}

type ChainUpgrade struct {
	Name   string `json:"name"`
	Height uint64 `json:"height"`
	Info   string `json:"info"`
}

type ConsensusBlockParamsUpdate = contract.BlockParams

type ConsensusEvidenceParamsUpdate = contract.EvidenceParams

type Migration struct {
	/// the contract address to be migrated
	Contract string `json:"contract"`
	/// a reference to the new WASM code that it should be migrated to
	CodeId uint64 `json:"code_id"`
	/// encoded message to be passed to perform the migration
	MigrateMsg []byte `json:"migrate_msg"`
}
