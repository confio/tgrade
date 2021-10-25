package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrDeliverGenTXFailed              = sdkerrors.Register(ModuleName, 2, "tx failed")
	ErrValidatorPubKeyTypeNotSupported = sdkerrors.Register(ModuleName, 3, "validator pubkey type is not supported")
	ErrInvalidHistoricalInfo           = sdkerrors.Register(ModuleName, 4, "invalid historical info")
	ErrEmpty                           = sdkerrors.Register(ModuleName, 5, "empty")
	ErrInvalid                         = sdkerrors.Register(ModuleName, 6, "invalid")
	ErrNotFound                        = sdkerrors.Register(ModuleName, 7, "not found")
)
