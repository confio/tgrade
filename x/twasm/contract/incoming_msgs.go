package contract

// TgradeMsg messages coming from a contract
type TgradeMsg struct {
	Hooks *Hooks `json:"hooks"`
}

// Hooks contains method to interact with system callbacks
type Hooks struct {
	RegisterBeginBlock   *struct{} `json:"register_begin_block"`
	UnregisterBeginBlock *struct{} `json:"unregister_begin_block"`
	// these are called the end of every block
	RegisterEndBlock   *struct{} `json:"register_end_block"`
	UnregisterEndBlock *struct{} `json:"unregister_end_block"`
	// only max 1 contract can be registered here, this is called in EndBlock (after everything else) and can change the validator set.
	RegisterValidatorSetUpdate   *struct{} `json:"register_validator_set_update"`
	UnregisterValidatorSetUpdate *struct{} `json:"unregister_validator_set_update"`
}
