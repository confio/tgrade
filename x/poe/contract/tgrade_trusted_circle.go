package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/types"
)

type TrustedCircleInitMsg struct { //nolint:maligned
	// Name Trusted circle name
	Name string `json:"name"`
	// Denom Trusted circle denom
	Denom string `json:"denom"`
	// EscrowAmount The required escrow amount, in the Denom
	EscrowAmount sdk.Int `json:"escrow_amount"`
	// VotingPeriod Voting period in days
	VotingPeriod uint32 `json:"voting_period"`
	// Quorum voting quorum (0.0-1.0)
	Quorum sdk.Dec `json:"quorum"`
	// Threshold voting threshold (0.0-1.0)
	Threshold sdk.Dec `json:"threshold"`
	// AllowEndEarly If true, and absolute threshold and quorum are met, we can end before voting period finished.
	// (Recommended value: true, unless you have special needs)
	AllowEndEarly bool `json:"allow_end_early"`
	// InitialMembers is a list of non-voting members to be added to the TRUSTED_CIRCLE upon creation
	InitialMembers []string `json:"initial_members"`
	// DenyList is an optional cw4 contract with list of addresses denied being part of the trusted circle
	DenyList string `json:"deny_list,omitempty"`
	// EditTrustedCircleDisabled If true, no further adjustments may happen
	EditTrustedCircleDisabled bool `json:"edit_trusted_circle_disabled"`
	// RewardDenom is the token denom we can distribute to the trusted circle
	RewardDenom string `json:"reward_denom"`
}

type TrustedCircleExecute struct {
	DepositEscrow *struct{}   `json:"deposit_escrow,omitempty"`
	Propose       *ProposeMsg `json:"propose,omitempty"`
	Execute       *ExecuteMsg `json:"execute,omitempty"`
}

type ExecuteMsg struct {
	ProposalID uint64 `json:"proposal_id"`
}

type ProposeMsg struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Proposal    ProposalContent `json:"proposal"`
}

type ProposalContent struct {
	AddVotingMembers          *AddVotingMembers          `json:"add_voting_members,omitempty"`
	AddRemoveNonVotingMembers *AddRemoveNonVotingMembers `json:"add_remove_non_voting_members,omitempty"`
}

type AddVotingMembers struct {
	Voters []string `json:"voters"`
}

type AddRemoveNonVotingMembers struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

type TrustedCircleContractAdapter struct {
	VotingContractAdapter
}

// NewTrustedCircleContractAdapter  constructor
func NewTrustedCircleContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) TrustedCircleContractAdapter {
	return TrustedCircleContractAdapter{
		VotingContractAdapter: NewVotingContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		)}
}

// AddVotingMembersProposal set up the proposal for adding voting members
func (a TrustedCircleContractAdapter) AddVotingMembersProposal(ctx sdk.Context, members []string, sender sdk.AccAddress) error {
	msg := TrustedCircleExecute{
		Propose: &ProposeMsg{
			Title:       "Add voting members",
			Description: "Add voting members",
			Proposal: ProposalContent{
				AddVotingMembers: &AddVotingMembers{
					Voters: members,
				},
			},
		},
	}
	return a.doExecute(ctx, msg, sender)
}

// DepositEscrow deposits escrow for the given member
func (a TrustedCircleContractAdapter) DepositEscrow(ctx sdk.Context, deposit sdk.Coin, sender sdk.AccAddress) error {
	msg := TrustedCircleExecute{
		DepositEscrow: &struct{}{},
	}
	return a.doExecute(ctx, msg, sender, deposit)
}

// QueryListVoters query the list of voters
func (a TrustedCircleContractAdapter) QueryListVoters(ctx sdk.Context) (*TG4MemberListResponse, error) {
	query := ProposalsQuery{ListVoters: &ListVotersQuery{}}
	var rsp TG4MemberListResponse
	err := a.doQuery(ctx, query, &rsp)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "contract query")
	}
	return &rsp, nil
}
