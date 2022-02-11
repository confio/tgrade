package contract

type OCGovProposalMsg struct {
	Propose *OCGovProposalSubmit  `json:"propose,omitempty"`
	Vote    *OCGovProposalVote    `json:"vote,omitempty"`
	Execute *OCGovProposalExecute `json:"execute,omitempty"`
}

type OCGovProposalSubmit struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Proposal    Proposal `json:"proposal"`
}

type Proposal struct {
	GrantEngagement EngagementMember `json:"grant_engagement"`
}

const (
	VoteYes     = "yes"
	VoteNo      = "no"
	VoteAbstain = "abstain"
	VoteVeto    = "veto"
)

type OCGovProposalVote struct {
	ProposalID          uint64 `json:"proposal_id"`
	OCGovProposalOption string `json:"vote"`
}

type OCGovProposalExecute struct {
	ProposalID uint64 `json:"proposal_id"`
}

type EngagementMember struct {
	Addr   string `json:"member"`
	Points uint64 `json:"points"`
}
