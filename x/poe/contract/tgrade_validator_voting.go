package contract

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
// Incomplete list as this is used for testing only within tgrade. Clients may use all options defined in the contract.
type ValidatorProposal struct {
	// PinCodes that should be pinned in cache for high performance
	PinCodes        []uint64      `json:"pin_codes,omitempty"`
	RegisterUpgrade *ChainUpgrade `json:"register_upgrade,omitempty"`
}

// ChainUpgrade defines a subset of attributes for testing only
type ChainUpgrade struct {
	Name   string `json:"name"`
	Height uint64 `json:"height"`
	Info   string `json:"info"`
}
