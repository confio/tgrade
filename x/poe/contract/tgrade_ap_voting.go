package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// APVotingInitMsg instantiation message
type APVotingInitMsg struct {
	VotingRules VotingRules `json:"rules"`
	// GroupContractAddress is the group contract that contains the member list
	GroupContractAddress string `json:"group_addr"`
	// Dispute cost on this contract
	DisputeCost sdk.Coin `json:"dispute_cost"`
	// Waiting period for this contract
	WaitingPeriod uint64 `json:"waiting_period"`
}
