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
)

var (
	_ sdk.Msg = &MsgCreateValidator{}
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
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
		Description:      description,
		DelegatorAddress: valAddr.String(),
		Pubkey:           pkAny,
		Value:            selfDelegation,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgCreateValidator) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgCreateValidator) Type() string { return TypeMsgCreateValidator }

// GetSigners implements the sdk.Msg interface. It returns the address(es) that
// must sign over msg.GetSignBytes().
// If the validator address is not same as delegator's, then the validator must
// sign the msg as well.
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
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
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
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
