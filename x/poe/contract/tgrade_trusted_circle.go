package contract

import sdk "github.com/cosmos/cosmos-sdk/types"

type TrustedCircleInitMsg struct { //nolint:maligned
	// Name of trusted circle
	Name string `json:"name"`
	// EscrowAmount The required escrow amount, in the default denom (utgd)
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
	// DenyList is an optional cw4 contract with list of addresses denied to be part of TrustedCircle
	DenyList string `json:"deny_list,omitempty"`
	// EditTrustedCircleDisabled If true, no further adjustments may happen.
	EditTrustedCircleDisabled bool `json:"edit_trusted_circle_disabled"`
	// RewardDenom is the token denom we can distribute to the trusted circle
	RewardDenom string `json:"reward_denom"`
}
