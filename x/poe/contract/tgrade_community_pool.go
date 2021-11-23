package contract

type CommunityPoolInitMsg struct {
	VotingRules  VotingRules `json:"rules"`
	GroupAddress string      `json:"group_addr"`
	DELME        string      `json:"engagement_addr"` // remove when https://github.com/confio/tgrade-contracts/issues/348 is done
}
