package types

func PromoteProposalFixture(mutators ...func(*PromoteToPrivilegedContractProposal)) *PromoteToPrivilegedContractProposal {
	const anyAddress = "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du"
	p := &PromoteToPrivilegedContractProposal{
		Title:       "Foo",
		Description: "Bar",
		Contract:    anyAddress,
	}
	for _, m := range mutators {
		m(p)
	}
	return p
}

func DemoteProposalFixture(mutators ...func(proposal *DemotePrivilegedContractProposal)) *DemotePrivilegedContractProposal {
	const anyAddress = "cosmos1qyqszqgpqyqszqgpqyqszqgpqyqszqgpjnp7du"
	p := &DemotePrivilegedContractProposal{
		Title:       "Foo",
		Description: "Bar",
		Contract:    anyAddress,
	}
	for _, m := range mutators {
		m(p)
	}
	return p
}
