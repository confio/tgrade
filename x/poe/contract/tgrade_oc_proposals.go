package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type OCProposalsContractAdapter struct {
	VotingContractAdapter
}

// NewOCProposalsContractAdapter constructor
func NewOCProposalsContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *OCProposalsContractAdapter {
	return &OCProposalsContractAdapter{
		VotingContractAdapter: NewVotingContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		)}
}

// ProposeSlash creates a proposal to slash this account
// Use LatestProposal after to get the ProposalID
func (v OCProposalsContractAdapter) ProposeSlash(ctx sdk.Context, member sdk.AccAddress, portion sdk.Dec, sender sdk.AccAddress) error {
	msg := OCProposalsExecuteMsg{
		Propose: &ProposalMsg{
			Title:       "Slash them",
			Description: "Slash them harder!",
			Proposal: OversightProposal{
				Slash: &SlashProposal{
					Member:  member.String(),
					Portion: portion,
				},
			},
		},
	}
	return v.doExecute(ctx, msg, sender)
}

// ProposeGrant creates a proposal to grant engagement to this account
// Use LatestProposal after to get the ProposalID
func (v OCProposalsContractAdapter) ProposeGrant(ctx sdk.Context, grantee sdk.AccAddress, points uint64, sender sdk.AccAddress) error {
	msg := OCProposalsExecuteMsg{
		Propose: &ProposalMsg{
			Title:       "Slash them",
			Description: "Slash them harder!",
			Proposal: OversightProposal{
				GrantEngagement: &GrantEngagementProposal{
					Member: grantee.String(),
					Points: points,
				},
			},
		},
	}
	return v.doExecute(ctx, msg, sender)
}

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

type OCProposalsExecuteMsg struct {
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
	Slash           *SlashProposal           `json:"punish,omitempty"`
}

type GrantEngagementProposal struct {
	Member string `json:"member"`
	Points uint64 `json:"points"`
}

type SlashProposal struct {
	Member  string  `json:"member"`
	Portion sdk.Dec `json:"portion"`
}

type OCProposalResponse struct {
	ID          uint64            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Proposal    OversightProposal `json:"proposal"`
	Status      ProposalStatus    `json:"status"`
	CreatedBy   string            `json:"created_by"`
	// TODO: clarify this format
	// Expires     EXP               `json:"expires"`
	Rules       VotingRules `json:"rules"`
	TotalWeight uint64      `json:"total_weight"`
	Votes       Votes       `json:"votes"`
}

type OCProposalListResponse struct {
	Proposals []OCProposalResponse `json:"proposals"`
}
