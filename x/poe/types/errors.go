package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrDeliverGenTXFailed              = sdkerrors.Register(ModuleName, 2, "tx failed")
	ErrValidatorPubKeyTypeNotSupported = sdkerrors.Register(ModuleName, 3, "validator pubkey type is not supported")
)
