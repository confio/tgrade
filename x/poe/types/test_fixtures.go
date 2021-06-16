package types

import (
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
