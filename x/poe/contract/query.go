package contract

// ValsetConfigQueryResponse Response to `config` query
// https://github.com/confio/tgrade-contracts/blob/v0.1.5/contracts/tgrade-valset/schema/query_msg.json#L3
type ValsetConfigQueryResponse struct {
	Membership    string `json:"membership"`
	MinWeight     int    `json:"min_weight"`
	MaxValidators int    `json:"max_validators"`
	Scaling       int    `json:"scaling,omitempty"`
}

// QueryTG4GroupResponse response to a list members query.
type QueryTG4GroupResponse struct {
	Members []TG4Member `json:"members"`
}
