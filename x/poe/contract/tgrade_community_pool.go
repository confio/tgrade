package contract

type CommunityPoolInitMsg struct {
	VotingRules  VotingRules `json:"rules"`
	GroupAddress string      `json:"group_addr"`
}
