package types_test

import (
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	coinPos     = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	coinZero    = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	pk1         = ed25519.GenPrivKey().PubKey()
	valAddr1    = sdk.AccAddress(pk1.Address())
	emptyAddr   sdk.AccAddress
	emptyPubkey cryptotypes.PubKey
)

func TestMsgDecode(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// firstly we start testing the pubkey serialization

	pk1bz, err := cdc.MarshalInterface(pk1)
	require.NoError(t, err)
	var pkUnmarshaled cryptotypes.PubKey
	err = cdc.UnmarshalInterface(pk1bz, &pkUnmarshaled)
	require.NoError(t, err)
	require.True(t, pk1.Equals(pkUnmarshaled.(*ed25519.PubKey)))

	// now let's try to serialize the whole message

	msg, err := types.NewMsgCreateValidator(valAddr1, pk1, coinPos, stakingtypes.Description{})
	require.NoError(t, err)
	msgSerialized, err := cdc.MarshalInterface(msg)
	require.NoError(t, err)

	var msgUnmarshaled sdk.Msg
	err = cdc.UnmarshalInterface(msgSerialized, &msgUnmarshaled)
	require.NoError(t, err)
	msg2, ok := msgUnmarshaled.(*types.MsgCreateValidator)
	require.True(t, ok)
	require.True(t, msg.Value.IsEqual(msg2.Value))
	require.True(t, msg.Pubkey.Equal(msg2.Pubkey))
}

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		delegatorAddr                                              sdk.AccAddress
		pubkey                                                     cryptotypes.PubKey
		bond                                                       sdk.Coin
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", valAddr1, pk1, coinPos, true},
		{"partial description", "", "", "c", "", "", valAddr1, pk1, coinPos, true},
		{"empty description", "", "", "", "", "", valAddr1, pk1, coinPos, false},
		{"empty address", "a", "b", "c", "d", "e", emptyAddr, pk1, coinPos, false},
		{"empty pubkey", "a", "b", "c", "d", "e", valAddr1, emptyPubkey, coinPos, false},
		{"empty bond", "a", "b", "c", "d", "e", valAddr1, pk1, coinZero, false},
		{"nil bond", "a", "b", "c", "d", "e", valAddr1, pk1, sdk.Coin{}, false},
	}

	for _, tc := range tests {
		description := stakingtypes.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		msg, err := types.NewMsgCreateValidator(tc.delegatorAddr, tc.pubkey, tc.bond, description)
		require.NoError(t, err)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
