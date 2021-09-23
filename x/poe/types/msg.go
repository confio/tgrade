package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	TypeMsgCreateValidator = "create_validator"
	TypeMsgUpdateValidator = "update_validator"
	TypeMsgUndelegate      = "begin_unbonding"
	TypeMsgDelegate        = "delegate"
)

var (
	_ sdk.Msg = &MsgCreateValidator{}
	_ sdk.Msg = &MsgUpdateValidator{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Operator address and validator address are the same.
func NewMsgCreateValidator(
	valAddr sdk.AccAddress,
	pubKey cryptotypes.PubKey,
	selfDelegation sdk.Coin,
	description stakingtypes.Description,
) (*MsgCreateValidator, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateValidator{
		Description:     description,
		OperatorAddress: valAddr.String(),
		Pubkey:          pkAny,
		Value:           selfDelegation,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCreateValidator) Type() string { return TypeMsgCreateValidator }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgCreateValidator) ValidateBasic() error {
	// note that unmarshaling from bech32 ensures either empty or valid
	delAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return err
	}
	if delAddr.Empty() {
		return stakingtypes.ErrEmptyDelegatorAddr
	}

	if msg.Pubkey == nil {
		return stakingtypes.ErrEmptyValidatorPubKey
	}

	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
		return stakingtypes.ErrBadDelegationAmount
	}

	if msg.Description == (stakingtypes.Description{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if len(msg.Description.Moniker) < 3 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "moniker must be at least 3 characters")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.Pubkey, &pubKey)
}

// NewMsgUpdateValidator creates a new MsgUpdateValidator instance.
// Operator address and validator address are the same.
func NewMsgUpdateValidator(
	valAddr sdk.AccAddress,
	description stakingtypes.Description,
) *MsgUpdateValidator {
	return &MsgUpdateValidator{
		Description:     description,
		OperatorAddress: valAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUpdateValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUpdateValidator) Type() string { return TypeMsgUpdateValidator }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
func (msg MsgUpdateValidator) GetSigners() []sdk.AccAddress {
	opAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{opAddr}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpdateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUpdateValidator) ValidateBasic() error {
	// note that unmarshaling from bech32 ensures either empty or valid
	_, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return err
	}

	if msg.Description == (stakingtypes.Description{}) ||
		msg.Description == stakingtypes.NewDescription(
			stakingtypes.DoNotModifyDesc,
			stakingtypes.DoNotModifyDesc,
			stakingtypes.DoNotModifyDesc,
			stakingtypes.DoNotModifyDesc,
			stakingtypes.DoNotModifyDesc,
		) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if len(msg.Description.Moniker) < 3 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "moniker must be at least 3 characters")
	}

	return nil
}

// NewMsgDelegate constructor
func NewMsgDelegate(delAddr sdk.AccAddress, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		OperatorAddress: delAddr.String(),
		Amount:          amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgDelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgDelegate) Type() string { return TypeMsgDelegate }

// GetSigners implements the sdk.Msg interface.
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgDelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.OperatorAddress); err != nil {
		return sdkerrors.Wrap(ErrEmpty, "operator address")
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(ErrInvalid, "delegation amount")
	}

	return nil
}

// NewMsgUndelegate constructor
func NewMsgUndelegate(delAddr sdk.AccAddress, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		OperatorAddress: delAddr.String(),
		Amount:          amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUndelegate) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUndelegate) Type() string { return TypeMsgUndelegate }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress {
	delAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUndelegate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUndelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.OperatorAddress); err != nil {
		return sdkerrors.Wrap(ErrEmpty, "operator address")
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return sdkerrors.Wrap(ErrInvalid, "delegation amount")
	}

	return nil
}
