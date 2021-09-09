package types

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func MsgCreateValidatorFixture(mutators ...func(m *MsgCreateValidator)) *MsgCreateValidator {
	pk1 := ed25519.GenPrivKey().PubKey()
	desc := stakingtypes.NewDescription("testname", "", "", "", "")

	r, err := NewMsgCreateValidator(sdk.AccAddress(pk1.Address()), pk1, sdk.NewInt64Coin(DefaultBondDenom, 50), desc)
	if err != nil {
		panic(err)
	}
	for _, m := range mutators {
		m(r)
	}
	return r
}

func MsgUpdateValidatorFixture(mutators ...func(m *MsgUpdateValidator)) *MsgUpdateValidator {
	desc := stakingtypes.NewDescription("other-name", "foo", "http://example.com", "bar", "my details")

	r := NewMsgUpdateValidator(RandomAccAddress(), desc)
	for _, m := range mutators {
		m(r)
	}
	return r
}

func GenesisStateFixture(mutators ...func(m *GenesisState)) GenesisState {
	r := DefaultGenesisState()
	r.Engagement = []TG4Member{{
		Address: RandomAccAddress().String(),
		Weight:  10,
	}}

	for _, m := range mutators {
		m(&r)
	}
	return r
}

func ValidatorFixture(mutators ...func(m *stakingtypes.Validator)) stakingtypes.Validator {
	pkAny, err := codectypes.NewAnyWithValue(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(fmt.Sprintf("failed to encode any type: %s", err.Error()))
	}
	desc := stakingtypes.Description{
		Moniker:         "myMoniker",
		Identity:        "myIdentity",
		Website:         "https://example.com",
		SecurityContact: "myContact",
		Details:         "myDetails",
	}
	r := stakingtypes.Validator{
		OperatorAddress: RandomAccAddress().String(),
		ConsensusPubkey: pkAny,
		Description:     desc,
		DelegatorShares: sdk.OneDec(),
	}

	for _, m := range mutators {
		m(&r)
	}
	return r
}
