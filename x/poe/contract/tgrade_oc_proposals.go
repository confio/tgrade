package contract

import sdk "github.com/cosmos/cosmos-sdk/types"

// OCProposalsInitMsg instantiation message
type OCProposalsInitMsg struct {
	VotingRules VotingRules `json:"rules"`
	// GroupContractAddress is the group contract that contains the member list
	GroupContractAddress string `json:"group_addr"`
	// EngagementContractAddress is the engagement contract that contains list for engagement rewards
	EngagementContractAddress string `json:"engagement_addr"`
	// ValsetContractAddress is the valset contract that we execute slashing on
	ValsetContractAddress string `json:"valset_addr"`
}

// VotingRules voting rules
type VotingRules struct {
	// VotingPeriod Voting period in days
	VotingPeriod uint32 `json:"voting_period"`
	// Quorum voting quorum (0.0-1.0)
	Quorum sdk.Dec `json:"quorum"`
	// Threshold voting threshold (0.0-1.0)
	Threshold sdk.Dec `json:"threshold"`
	// AllowEndEarly If true, and absolute threshold and quorum are met, we can end before voting period finished.
	// (Recommended value: true, unless you have special needs)
	AllowEndEarly bool `json:"allow_end_early"`
}

type ExecuteMsg struct {
	Propose *ProposalMsg `json:"propose,omitempty"`
	Vote    *VoteMsg     `json:"vote,omitempty"`
	Execute *ProposalID  `json:"execute,omitempty"`
	Close   *ProposalID  `json:"close,omitempty"`
}

type ProposalMsg struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Proposal    OversightProposal `json:"proposal"`
}

type OversightProposal struct {
	GrantEngagement *GrantEngagementProposal `json:"grant_engagement,omitempty"`
	Slash           *SlashProposal           `json:"slash,omitempty"`
}

type GrantEngagementProposal struct {
	Member string `json:"member"`
	Points uint64 `json:"points"`
}

type SlashProposal struct {
	Member  string  `json:"member"`
	Portion sdk.Dec `json:"portion"`
}

type VoteMsg struct {
	Vote       Vote   `json:"vote"`
	ProposalID uint64 `json:"proposal_id"`
}

type ProposalID struct {
	ProposalID uint64 `json:"proposal_id"`
}

type QueryMsg struct {
	// Returns VotingRules
	Rules *struct{} `json:"rules,omitempty"`
	// Returns OCProposalResponse
	Proposal *ProposalID `json:"proposal,omitempty"`
	// Returns OCProposalListResponse
	ListProposals *ListProposalQuery `json:"list_proposals,omitempty"`
	// Returns OCProposalListResponse
	ReverseProposals *ListProposalQuery `json:"reverse_proposals,omitempty"`
	// Returns VoteResponse
	Vote *VoteQuery `json:"vote,omitempty"`
	// Returns VoteListResponse
	ListVotes *ListVotesQuery `json:"list_votes,omitempty"`
	// Returns VoterResponse
	Voter *VoterQuery `json:"voter,omitempty"`
	// Returns VoterListResponse
	ListVoters *ListVotersQuery `json:"list_voters,omitempty"`
}

type ListProposalQuery struct {
	StartAfter uint64 `json:"start_after,omitempty"`
	Limit      uint32 `json:"limit,omitempty"`
}

type VoteQuery struct {
	ProposalID uint64 `json:"proposal_id"`
	Voter      string `json:"voter"`
}

type ListVotesQuery struct {
	ProposalID uint64 `json:"proposal_id"`
	StartAfter string `json:"start_after,omitempty"`
	Limit      uint32 `json:"limit,omitempty"`
}

type VoterQuery struct {
	Address string `json:"address"`
}

type ListVotersQuery struct {
	StartAfter string `json:"start_after,omitempty"`
	Limit      uint32 `json:"limit,omitempty"`
}

type OCProposalResponse struct {
	ID          uint64            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Proposal    OversightProposal `json:"proposal"`
	Status      Status            `json:"status"`
	// TODO: clarify this format
	// Expires     EXP               `json:"expires"`
	Rules       VotingRules `json:"rules"`
	TotalWeight uint64      `json:"total_weight"`
	Votes       Votes       `json:"votes"`
}

type Votes struct {
	Yes     uint64 `json:"yes"`
	No      uint64 `json:"no"`
	Abstain uint64 `json:"abstain"`
	Veto    uint64 `json:"veto"`
}

type OCProposalListResponse struct {
	Proposals []OCProposalResponse `json:"proposals"`
}

type VoteListResponse struct {
	Votes []VoteInfo `json:"votes"`
}

type VoteInfo struct {
	Voter  string `json:"voter"`
	Vote   Vote   `json:"vote"`
	Weight uint64 `json:"weight"`
}

type VoteResponse struct {
	Vote *VoteInfo `json:"vote,omitempty"`
}

type VoterResponse struct {
	Weight *uint64 `json:"weight"`
}

type VoterListResponse struct {
	Voters []VoterDetail `json:"voters"`
}

type VoterDetail struct {
	Addr   string `json:"addr"`
	Weight uint64 `json:"weight"`
}

type Vote string

const (
	YES_VOTE     Vote = "yes"
	NO_VOTE      Vote = "no"
	ABSTAIN_VOTE Vote = "abstain"
	VETO_VOTE    Vote = "veto"
)

type Status string

const (
	STATUS_PENDING  Status = "pending"
	STATUS_OPEN     Status = "open"
	STATUS_REJECTED Status = "rejected"
	STATUS_PASSED   Status = "passed"
	STATUS_EXECUTED Status = "executed"
)
