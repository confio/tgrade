package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// APVotingInitMsg instantiation message
type APVotingInitMsg struct {
	VotingRules VotingRules `json:"rules"`
	// GroupContractAddress is the group contract that contains the member list
	GroupContractAddress string `json:"group_addr"`
	// Dispute cost on this contract
	DisputeCost sdk.Coin `json:"dispute_cost"`
	// Waiting period in seconds for this contract
	WaitingPeriod uint64 `json:"waiting_period"`
	// Cw3-fixed-multisig contract code id.
	// This code id will be used to instantiate a group contract with the proposed arbiters list.
	MultisigCodeID uint64 `json:"multisig_code_id"`
}

// APVotingExecute ap-voting contract execute messages
// See https://github.com/confio/tgrade-contracts/blob/v0.9.0/contracts/tgrade-ap-voting/src/msg.rs
type APVotingExecute struct {
	Propose           *Propose           `json:"propose,omitempty"`
	Vote              *VoteProposal      `json:"vote,omitempty"`
	Execute           *ExecuteProposal   `json:"execute,omitempty"`
	Close             *CloseProposal     `json:"close,omitempty"`
	RegisterComplaint *RegisterComplaint `json:"register_complaint,omitempty"`
	AcceptComplaint   *AcceptComplaint   `json:"accept_complaint,omitempty"`
	WithdrawComplaint *WithdrawComplaint `json:"withdraw_complaint,omitempty"`
	RenderDecision    *RenderDecision    `json:"render_decision,omitempty"`
}

// Propose arbiters for a given dispute
type Propose struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	APProposal  APProposal `json:"arbiter_pool_proposal"`
}

type APProposal struct {
	// An open text proposal with no actual logic executed when it passes
	Text *struct{} `json:"text,omitempty"`
	// Proposes arbiters for existing complaint
	ProposeArbiters *ProposeArbiters `json:"propose_arbiters,omitempty"`
}

type ProposeArbiters struct {
	CaseID   uint64           `json:"case_id"`
	Arbiters []sdk.AccAddress `json:"arbiters"`
}

type VoteProposal struct {
	ProposalID uint64 `json:"proposal_id"`
	Vote       Vote   `json:"vote"`
}

type ExecuteProposal struct {
	ProposalID uint64 `json:"proposal_id"`
}

type CloseProposal struct {
	ProposalID uint64 `json:"proposal_id"`
}

type RegisterComplaint struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Defendant   string `json:"defendant"`
}

type AcceptComplaint struct {
	ComplaintID uint64 `json:"complaint_id"`
}

type WithdrawComplaint struct {
	ComplaintID uint64 `json:"complaint_id"`
	Reason      string `json:"reason"`
}

type RenderDecision struct {
	ComplaintID uint64 `json:"complaint_id"`
	Summary     string `json:"summary"`
	IpfsLink    string `json:"ipfs_link"`
}
