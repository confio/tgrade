package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var ( // leave enough space for wasm error types

	ErrValidatorPubKeyTypeNotSupported = sdkerrors.Register(ModuleName, 100, "validator pubkey type is not supported")
)
