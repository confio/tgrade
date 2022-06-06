package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/confio/tgrade/x/poe/types"
)

type TCPaymentsContractAdapter struct {
	BaseContractAdapter
}

// NewTCPaymentsContractAdapter constructor
func NewTCPaymentsContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) *TCPaymentsContractAdapter {
	return &TCPaymentsContractAdapter{
		BaseContractAdapter: NewBaseContractAdapter(
			contractAddr,
			twasmKeeper,
			addressLookupErr,
		),
	}
}

// TcPaymentsInitMsg instantiation message
type TcPaymentsInitMsg struct {
	// Admin (if set) can change the payment amount and period
	Admin string `json:"admin,omitempty"`
	// Trusted Circle / OC contract address
	OcContractAddr string `json:"oc_addr"`
	// Arbiter pool contract address
	ApContractAddr string `json:"ap_addr"`
	// EngagementContractAddress is the engagement contract that contains the members list for engagement rewards.
	// To send the remaining funds after payment
	EngagementContractAddr string `json:"engagement_addr"`
	// The required payment amount, in the payments denom
	Denom string `json:"denom"`
	// The required payment amount, in the TC denom
	PaymentAmount uint64 `json:"payment_amount,string"`
	// Payment period
	PaymentPeriod Period `json:"payment_period"`
}

type Period struct {
	Daily   *struct{} `json:"daily,omitempty"`
	Monthly *struct{} `json:"monthly,omitempty"`
	Yearly  *struct{} `json:"yearly,omitempty"`
}
