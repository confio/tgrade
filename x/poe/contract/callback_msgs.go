package contract

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/confio/tgrade/x/poe/types"
)

// ValidatorUpdateSudoMsg callback message sent to a contract.
// See https://github.com/confio/tgrade-contracts/blob/main/packages/bindings/src/sudo.rs
type ValidatorUpdateSudoMsg struct {
	/// This will be delivered after everything.
	/// The data in the Response is (JSON?) encoded diff to the validator set
	EndWithValidatorUpdate *struct{} `json:"end_with_validator_update,omitempty"`
}

// EndWithValidatorUpdateResponse is the response to an `EndWithValidatorUpdate` sudo call.
type EndWithValidatorUpdateResponse struct {
	Diffs []ValidatorUpdate `json:"diffs"`
}

// ValidatorUpdate  is used to update the validator set
// See https://github.com/tendermint/tendermint/blob/v0.34.8/proto/tendermint/abci/types.proto#L343-L346
type ValidatorUpdate struct {
	// PubKey is the ed25519 pubkey used in Tendermint consensus
	PubKey ValidatorPubkey `json:"pubkey"`
	// Power is the new voting power in the consensus rounds
	Power uint64 `json:"power"`
}

type ValidatorPubkey struct {
	Ed25519   []byte `json:"ed25519,omitempty"`
	Secp256k1 []byte `json:"secp256k1,omitempty"`
	Sr25519   []byte `json:"sr25519,omitempty"`
}

func NewValidatorPubkey(pk cryptotypes.PubKey) (ValidatorPubkey, error) {
	switch pk.Type() {
	case "ed25519":
		return ValidatorPubkey{
			Ed25519: pk.Bytes(),
		}, nil
	case "secp256k1":
		return ValidatorPubkey{
			Secp256k1: pk.Bytes(),
		}, nil
	default:
		return ValidatorPubkey{}, errors.Wrap(types.ErrValidatorPubKeyTypeNotSupported, pk.Type())
	}
}
