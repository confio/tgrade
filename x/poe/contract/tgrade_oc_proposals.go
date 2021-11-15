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
