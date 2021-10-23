package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
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
	RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// firstly we start testing the pubkey serialization

	pk1bz, err := cdc.MarshalInterface(pk1)
	require.NoError(t, err)
	var pkUnmarshaled cryptotypes.PubKey
	err = cdc.UnmarshalInterface(pk1bz, &pkUnmarshaled)
	require.NoError(t, err)
	require.True(t, pk1.Equals(pkUnmarshaled.(*ed25519.PubKey)))

	// now let's try to serialize the whole message

	msg, err := NewMsgCreateValidator(valAddr1, pk1, coinPos, stakingtypes.Description{})
	require.NoError(t, err)
	msgSerialized, err := cdc.MarshalInterface(msg)
	require.NoError(t, err)

	var msgUnmarshaled sdk.Msg
	err = cdc.UnmarshalInterface(msgSerialized, &msgUnmarshaled)
	require.NoError(t, err)
	msg2, ok := msgUnmarshaled.(*MsgCreateValidator)
	require.True(t, ok)
	require.True(t, msg.Value.IsEqual(msg2.Value))
	require.True(t, msg.Pubkey.Equal(msg2.Pubkey))
}

// test ValidateBasic for MsgCreateValidator
func TestMsgCreateValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		operatorAddr                                               sdk.AccAddress
		pubkey                                                     cryptotypes.PubKey
		bond                                                       sdk.Coin
		expectPass                                                 bool
	}{
		{"basic good", "hello", "b", "c", "d", "e", valAddr1, pk1, coinPos, true},
		{"partial description", "hello", "", "c", "", "", valAddr1, pk1, coinPos, true},
		{"short moniker", "a", "", "", "", "", valAddr1, pk1, coinPos, false},
		{"empty description", "", "", "", "", "", valAddr1, pk1, coinPos, false},
		{"empty address", "hello", "b", "c", "d", "e", emptyAddr, pk1, coinPos, false},
		{"empty pubkey", "hello", "b", "c", "d", "e", valAddr1, emptyPubkey, coinPos, false},
		{"empty bond", "hello", "b", "c", "d", "e", valAddr1, pk1, coinZero, false},
		{"nil bond", "hello", "b", "c", "d", "e", valAddr1, pk1, sdk.Coin{}, false},
	}

	for _, tc := range tests {
		description := stakingtypes.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		msg, err := NewMsgCreateValidator(tc.operatorAddr, tc.pubkey, tc.bond, description)
		require.NoError(t, err)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

func TestMsgUpdateValidatorValidateBasic(t *testing.T) {
	specs := map[string]struct {
		src    *MsgUpdateValidator
		expErr bool
	}{
		"all good": {
			src: MsgUpdateValidatorFixture(),
		},
		"empty description": {
			src: MsgUpdateValidatorFixture(func(m *MsgUpdateValidator) {
				m.Description = stakingtypes.Description{}
			}),
			expErr: true,
		},
		"do not modify description": {
			src: MsgUpdateValidatorFixture(func(m *MsgUpdateValidator) {
				m.Description = stakingtypes.NewDescription(
					stakingtypes.DoNotModifyDesc,
					stakingtypes.DoNotModifyDesc,
					stakingtypes.DoNotModifyDesc,
					stakingtypes.DoNotModifyDesc,
					stakingtypes.DoNotModifyDesc,
				)
			}),
			expErr: true,
		},
		"invalid address": {
			src: MsgUpdateValidatorFixture(func(m *MsgUpdateValidator) {
				m.OperatorAddress = "notAValidAddress"
			}),
			expErr: true,
		},
		"empty address": {
			src: MsgUpdateValidatorFixture(func(m *MsgUpdateValidator) {
				bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
				bech32Addr, _ := bech32.ConvertAndEncode(bech32PrefixAccAddr, []byte{})
				m.OperatorAddress = bech32Addr
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestMsgDelegate(t *testing.T) {
	tests := []struct {
		name         string
		operatorAddr sdk.AccAddress
		bond         sdk.Coin
		expectPass   bool
	}{
		{"basic good", sdk.AccAddress(valAddr1), coinPos, true},
		{"empty operator", sdk.AccAddress(emptyAddr), coinPos, false},
		{"empty bond", sdk.AccAddress(valAddr1), coinZero, false},
		{"nil bold", sdk.AccAddress(valAddr1), sdk.Coin{}, false},
	}
	for _, tc := range tests {
		msg := NewMsgDelegate(tc.operatorAddr, tc.bond)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}

// test ValidateBasic for MsgUnbond
func TestMsgUndelegate(t *testing.T) {
	tests := []struct {
		name         string
		operatorAddr sdk.AccAddress
		amount       sdk.Coin
		expectPass   bool
	}{
		{"regular", sdk.AccAddress(valAddr1), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1), true},
		{"zero amount", sdk.AccAddress(valAddr1), sdk.NewInt64Coin(sdk.DefaultBondDenom, 0), false},
		{"nil amount", sdk.AccAddress(valAddr1), sdk.Coin{}, false},
		{"empty operator", sdk.AccAddress(emptyAddr), sdk.NewInt64Coin(sdk.DefaultBondDenom, 1), false},
	}

	for _, tc := range tests {
		msg := NewMsgUndelegate(tc.operatorAddr, tc.amount)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
