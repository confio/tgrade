package contract

import govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

// ExecuteGovProposalFixture text proposal type
func ExecuteGovProposalFixture(mutators ...func(proposal *ExecuteGovProposal)) ExecuteGovProposal {
	r := ExecuteGovProposal{
		Title:       "foo",
		Description: "bar",
		Proposal: GovProposalFixture(func(p *GovProposal) {
			p.Text = &govtypes.TextProposal{}
		}),
	}
	for _, m := range mutators {
		m(&r)
	}
	return r
}

// GovProposalFixture empty gov proposal type
func GovProposalFixture(mutators ...func(proposal *GovProposal)) GovProposal {
	r := GovProposal{}
	for _, m := range mutators {
		m(&r)
	}
	return r
}
