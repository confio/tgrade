package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/types"
)

type ProposalID struct {
	ProposalID uint64 `json:"proposal_id"`
}

// VotingRules voting rules
type VotingRules struct {
	// VotingPeriod Voting period in days
	VotingPeriod uint32 `json:"voting_period"`
	// Quorum voting quorum (0.0-1.0)
	Quorum sdk.Dec `json:"quorum"`
	// Threshold voting threshold (0.5-1.0)
	Threshold sdk.Dec `json:"threshold"`
	// AllowEndEarly If true, and absolute threshold and quorum are met, we can end before voting period finished.
	// (Recommended value: true, unless you have special needs)
	AllowEndEarly bool `json:"allow_end_early"`
}

type Vote string

const (
	YesVote     Vote = "yes"
	NoVote      Vote = "no"
	AbstainVote Vote = "abstain"
	VetoVote    Vote = "veto"
)

type ProposalStatus string

const (
	ProposalStatusPending  ProposalStatus = "pending"
	ProposalStatusOpen     ProposalStatus = "open"
	ProposalStatusRejected ProposalStatus = "rejected"
	ProposalStatusPassed   ProposalStatus = "passed"
	ProposalStatusExecuted ProposalStatus = "executed"
)

type ProposalsQuery struct {
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

type Votes struct {
	Yes     uint64 `json:"yes"`
	No      uint64 `json:"no"`
	Abstain uint64 `json:"abstain"`
	Veto    uint64 `json:"veto"`
}

type VoteMsg struct {
	Vote       Vote   `json:"vote"`
	ProposalID uint64 `json:"proposal_id"`
}

type VoteInfo struct {
	Voter  string `json:"voter"`
	Vote   Vote   `json:"vote"`
	Weight uint64 `json:"weight"`
}

type VoteListResponse struct {
	Votes []VoteInfo `json:"votes"`
}

type VoteResponse struct {
	Vote *VoteInfo `json:"vote,omitempty"`
}

type VoterResponse struct {
	Weight *uint64 `json:"weight"`
}

type VoterDetail struct {
	Addr   string `json:"addr"`
	Weight uint64 `json:"weight"`
}

type VoterListResponse struct {
	Voters []VoterDetail `json:"voters"`
}

type VotingContractAdapter struct {
	BaseContractAdapter
}

// NewVotingContractAdapter  constructor
func NewVotingContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) VotingContractAdapter {
	return VotingContractAdapter{
		BaseContractAdapter: NewBaseContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		),
	}
}

// LatestProposal gets info on the last proposal made, easy way to get the ProposalID
func (v VotingContractAdapter) LatestProposal(ctx sdk.Context) (*OCProposalResponse, error) {
	query := ProposalsQuery{ReverseProposals: &ListProposalQuery{Limit: 1}}
	var rsp OCProposalListResponse
	err := v.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	if len(rsp.Proposals) == 0 {
		return nil, nil
	}
	return &rsp.Proposals[0], nil
}

// QueryProposal query a proposal by id
func (v VotingContractAdapter) QueryProposal(ctx sdk.Context, id uint64) (*OCProposalResponse, error) {
	query := ProposalsQuery{Proposal: &ProposalID{ProposalID: id}}
	var rsp OCProposalResponse
	err := v.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	return &rsp, nil
}

// VoteProposal votes on a proposal
func (v VotingContractAdapter) VoteProposal(ctx sdk.Context, proposalID uint64, vote Vote, sender sdk.AccAddress) error {
	msg := OCProposalsExecuteMsg{
		Vote: &VoteMsg{
			ProposalID: proposalID,
			Vote:       vote,
		},
	}
	return v.doExecute(ctx, msg, sender)
}

// ExecuteProposal executes a previously passed proposal
func (v VotingContractAdapter) ExecuteProposal(ctx sdk.Context, proposalID uint64, sender sdk.AccAddress) error {
	msg := OCProposalsExecuteMsg{
		Execute: &ProposalID{
			ProposalID: proposalID,
		},
	}
	return v.doExecute(ctx, msg, sender)
}
