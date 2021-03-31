package contract

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

// TgradeSudoMsg callback message sent to a contract
type TgradeSudoMsg struct {
	/// This will be delivered every block if the contract is currently registered for Begin Block
	BeginBlock *BeginBlock `json:"begin_block,omitempty"`
	// This will be delivered every block if the contract is currently registered for End Block
	/// Block height and time is already in Env
	EndBlock *struct{} `json:"end_block,omitempty"`
	/// This will be delivered after everything.
	/// The data in the Response is (JSON?) encoded diff to the validator set
	EndWithValidatorUpdate *struct{}           `json:"end_with_validator_update,omitempty"`
	PrivilegeChange        *PrivilegeChangeMsg `json:"privilege_change,omitempty"`
}

// todo (reviewer): this is what abci expects as result in end block
type ValidatorUpdateResponse = []abci.ValidatorUpdate

/// TODO: define types based on subset of https://github.com/tendermint/tendermint/blob/master/proto/tendermint/abci/types.proto#L71-L76
type BeginBlock struct {
	Hash []byte
	//Header              Header     // ??? do we need more than the height and time already in Env?
	// LastCommitInfo      LastCommitInfo
	ByzantineValidators []abci.Evidence `json:"byzantine_validators"` // This is key for slashing - let's figure out a standard for these types
}

/// These are called on a contract when it is made privileged or demoted
type PrivilegeChangeMsg struct {
	/// This is called when a contract gets "privileged status".
	/// It is a proper place to call `RegisterXXX` methods that require this status.
	/// Contracts that require this should be in a "frozen" state until they get this callback.
	Promoted *struct{} `json:"promoted,omitempty"`
	/// This is called when a contract looses "privileged status"
	Demoted *struct{} `json:"demoted,omitempty"`
}
