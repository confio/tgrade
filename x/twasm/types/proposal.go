package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"strings"
)

type ProposalType string

const (
	ProposalTypePromoteContract ProposalType = "PromoteToPrivilegedContract"
	ProposalTypeDemoteContract  ProposalType = "DemotePrivilegedContract"
)

// EnableAllProposals contains all twasm gov types as keys.
var EnableAllProposals = []ProposalType{
	ProposalTypePromoteContract,
	ProposalTypeDemoteContract,
}

func init() { // register new content types with the sdk
	govtypes.RegisterProposalType(string(ProposalTypePromoteContract))
	govtypes.RegisterProposalType(string(ProposalTypeDemoteContract))

	govtypes.RegisterProposalTypeCodec(&PromoteToPrivilegedContractProposal{}, "twasm/PromoteToPrivilegedContractProposal")
	govtypes.RegisterProposalTypeCodec(&DemotePrivilegedContractProposal{}, "twasm/DemotePrivilegedContractProposal")
}

// ProposalRoute returns the routing key of a parameter change proposal.
func (p PromoteToPrivilegedContractProposal) ProposalRoute() string { return RouterKey }

// GetTitle returns the title of the proposal
func (p *PromoteToPrivilegedContractProposal) GetTitle() string { return p.Title }

// GetDescription returns the human readable description of the proposal
func (p PromoteToPrivilegedContractProposal) GetDescription() string { return p.Description }

// ProposalType returns the type
func (p PromoteToPrivilegedContractProposal) ProposalType() string {
	return string(ProposalTypePromoteContract)
}

// ValidateBasic validates the proposal
func (p PromoteToPrivilegedContractProposal) ValidateBasic() error {
	if err := validateProposalCommons(p.Title, p.Description); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(p.Contract); err != nil {
		return sdkerrors.Wrap(err, "contract")
	}
	return nil
}

// String implements the Stringer interface.
func (p PromoteToPrivilegedContractProposal) String() string {
	return fmt.Sprintf(`Store Code Proposal:
  Title:       %s
  Description: %s
  Contract:    %s
`, p.Title, p.Description, p.Contract)
}

// MarshalYAML pretty prints the wasm byte code
func (p PromoteToPrivilegedContractProposal) MarshalYAML() (interface{}, error) {
	return p, nil
}

// ProposalRoute returns the routing key of a parameter change proposal.
func (p DemotePrivilegedContractProposal) ProposalRoute() string { return RouterKey }

// GetTitle returns the title of the proposal
func (p *DemotePrivilegedContractProposal) GetTitle() string { return p.Title }

// GetDescription returns the human readable description of the proposal
func (p DemotePrivilegedContractProposal) GetDescription() string { return p.Description }

// ProposalType returns the type
func (p DemotePrivilegedContractProposal) ProposalType() string {
	return string(ProposalTypeDemoteContract)
}

// ValidateBasic validates the proposal
func (p DemotePrivilegedContractProposal) ValidateBasic() error {
	if err := validateProposalCommons(p.Title, p.Description); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(p.Contract); err != nil {
		return sdkerrors.Wrap(err, "contract")
	}
	return nil
}

// String implements the Stringer interface.
func (p DemotePrivilegedContractProposal) String() string {
	return fmt.Sprintf(`Store Code Proposal:
  Title:       %s
  Description: %s
  Contract:    %s
`, p.Title, p.Description, p.Contract)
}

// MarshalYAML pretty prints the wasm byte code
func (p DemotePrivilegedContractProposal) MarshalYAML() (interface{}, error) {
	return p, nil
}

// common validations
func validateProposalCommons(title, description string) error {
	if strings.TrimSpace(title) != title {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal title must not start/end with white spaces")
	}
	if len(title) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > govtypes.MaxTitleLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal title is longer than max length of %d", govtypes.MaxTitleLength)
	}
	if strings.TrimSpace(description) != description {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal description must not start/end with white spaces")
	}
	if len(description) == 0 {
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > govtypes.MaxDescriptionLength {
		return sdkerrors.Wrapf(govtypes.ErrInvalidProposalContent, "proposal description is longer than max length of %d", govtypes.MaxDescriptionLength)
	}
	return nil
}
