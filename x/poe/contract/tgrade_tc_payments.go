package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type TCPaymentsContractAdapter struct {
	ContractAdapter
}

// NewTCPaymentsContractAdapter constructor
func NewTCPaymentsContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *TCPaymentsContractAdapter {
	return &TCPaymentsContractAdapter{
		ContractAdapter: NewContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		),
	}
}

// TCPaymentsInitMsg instantiation message
type TCPaymentsInitMsg struct {
	VotingRules VotingRules `json:"rules"`
	// GroupContractAddress is the group contract that contains the member list
	GroupContractAddress string `json:"group_addr"`
	// EngagementContractAddress is the engagement contract that contains list for engagement rewards
	EngagementContractAddress string `json:"engagement_addr"`
	// ValsetContractAddress is the valset contract that we execute slashing on
	ValsetContractAddress string `json:"valset_addr"`

	// Admin (if set) can change the payment amount and period
	Admin string `json:"admin,omitempty"`
	// Trusted Circle / OC contract address
	OCAddr string `json:"oc_addr"`
	// Arbiter pool contract address
	APAddr string `json:"ap_addr"`
	// Engagement contract address.
	// To send the remaining funds after payment
	EngagementAddr string `json:"engagement_addr"`
	// The required payment amount, in the payments denom
	Denom string `json:"denom"`
	// The required payment amount, in the TC denom
	PaymentAmount uint64 `json:"payment_amount,string"`
	// Payment period
	PaymentPeriod Period `json:"payment_period"`
}

type Period struct {
	Daily   bool `json:"daily,omitempty"`
	Monthly bool `json:"monthly,omitempty"`
	Yearly  bool `json:"yearly,omitempty"`
}
