package keeper

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/rand"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper/poetesting"
	"github.com/confio/tgrade/x/poe/types"
	wasmtesting "github.com/confio/tgrade/x/twasm/testing"
)

func TestCreateValidator(t *testing.T) {
	var (
		myValsetContract  sdk.AccAddress = rand.Bytes(address.Len)
		myStakingContract sdk.AccAddress = rand.Bytes(address.Len)
		myOperatorAddr    sdk.AccAddress = rand.Bytes(address.Len)
	)
	var capturedOpAddr sdk.AccAddress
	var capturedSelfDelegation *sdk.Coin
	poeKeeperMock := PoEKeeperMock{
		GetPoEContractAddressFn: SwitchPoEContractAddressFn(t, myValsetContract, myStakingContract),
		SetValidatorInitialEngagementPointsFn: func(ctx sdk.Context, opAdr sdk.AccAddress, selfDelegation sdk.Coin) error {
			capturedOpAddr = opAdr
			capturedSelfDelegation = &selfDelegation
			return nil
		},
	}

	specs := map[string]struct {
		src               *types.MsgCreateValidator
		expSelfDelegation *sdk.Coin
		expErr            *sdkerrors.Error
	}{
		"all good": {
			src: types.MsgCreateValidatorFixture(
				func(m *types.MsgCreateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Value = sdk.NewInt64Coin(types.DefaultBondDenom, 1)
				},
			),
			expSelfDelegation: &sdk.Coin{Denom: types.DefaultBondDenom, Amount: sdk.NewInt(1)},
		},
		"invalid algo": {
			src: types.MsgCreateValidatorFixture(
				func(m *types.MsgCreateValidator) {
					pkAny, err := codectypes.NewAnyWithValue(secp256k1.GenPrivKey().PubKey())
					require.NoError(t, err)
					m.Pubkey = pkAny
				},
			),
			expErr: stakingtypes.ErrValidatorPubKeyTypeNotSupported,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			fn, execs := wasmtesting.CaptureExecuteFn()
			km := &wasmtesting.ContractOpsKeeperMock{
				ExecuteFn: fn,
			}
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()).WithEventManager(em).WithConsensusParams(&abci.ConsensusParams{
				Validator: &types1.ValidatorParams{PubKeyTypes: []string{"ed25519"}}}))

			// when
			s := NewMsgServerImpl(poeKeeperMock, km, nil)
			gotRes, gotErr := s.CreateValidator(ctx, spec.src)

			// then
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(gotErr), "exp %v but got %#+v", spec.expErr, gotErr)
				assert.Nil(t, gotRes)
				return
			}
			require.NoError(t, gotErr)
			// and contract called
			require.Len(t, *execs, 2)
			assert.Equal(t, myValsetContract, (*execs)[0].ContractAddress)
			assert.Equal(t, myOperatorAddr, (*execs)[0].Caller)
			assert.Nil(t, (*execs)[0].Coins)
			assert.Equal(t, myStakingContract, (*execs)[1].ContractAddress)
			assert.Equal(t, myOperatorAddr, (*execs)[1].Caller)
			assert.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.DefaultBondDenom, 1)), (*execs)[1].Coins)

			assert.Equal(t, myOperatorAddr, capturedOpAddr)
			assert.Equal(t, spec.expSelfDelegation, capturedSelfDelegation)

			// and events emitted
			require.NoError(t, gotErr)
			require.Len(t, em.Events(), 2)
			assert.Equal(t, sdk.EventTypeMessage, em.Events()[0].Type)
			assert.Equal(t, types.EventTypeCreateValidator, em.Events()[1].Type)
		})
	}
}

func TestUpdateValidator(t *testing.T) {
	var myValsetContract sdk.AccAddress = rand.Bytes(address.Len)
	var myOperatorAddr sdk.AccAddress = rand.Bytes(address.Len)

	poeKeeperMock := PoEKeeperMock{
		GetPoEContractAddressFn: SwitchPoEContractAddressFn(t, myValsetContract, nil),
		SetValidatorInitialEngagementPointsFn: func(ctx sdk.Context, address sdk.AccAddress, value sdk.Coin) error {
			return nil
		},
		ValsetContractFn: func(ctx sdk.Context) ValsetContract {
			return poetesting.ValsetContractMock{QueryValidatorFn: func(ctx sdk.Context, opAddr sdk.AccAddress) (*stakingtypes.Validator, error) {
				v := types.ValidatorFixture(func(m *stakingtypes.Validator) {
					m.OperatorAddress = myOperatorAddr.String()
				})
				return &v, nil
			}}
		},
	}

	desc := contract.MetadataFromDescription(stakingtypes.Description{
		Moniker:         "myMoniker",
		Identity:        "myIdentity",
		Website:         "https://example.com",
		SecurityContact: "myContact",
		Details:         "myDetails",
	})
	querier := TwasmKeeperMock{QuerySmartFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
		return json.Marshal(contract.ValidatorResponse{Validator: &contract.OperatorResponse{
			Operator: RandomAddress(t).String(),
			Pubkey:   contract.ValidatorPubkey{Ed25519: ed25519.GenPrivKey().PubKey().Bytes()},
			Metadata: desc,
		}})
	}}
	specs := map[string]struct {
		src    *types.MsgUpdateValidator
		mock   poetesting.ValsetContractMock
		exp    *contract.ValidatorMetadata
		expErr *sdkerrors.Error
	}{
		"update all": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						"otherMoniker",
						"otherIdentity",
						"https://otherWebsite",
						"otherContact",
						"otherDetails",
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "otherMoniker",
				Identity:        "otherIdentity",
				Website:         "https://otherWebsite",
				SecurityContact: "otherContact",
				Details:         "otherDetails",
			},
		},
		"partial update: moniker": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						"otherMoniker",
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "otherMoniker",
				Identity:        "myIdentity",
				Website:         "https://example.com",
				SecurityContact: "myContact",
				Details:         "myDetails",
			},
		},
		"partial update: Identity": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						stakingtypes.DoNotModifyDesc,
						"otherIdentity",
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "myMoniker",
				Identity:        "otherIdentity",
				Website:         "https://example.com",
				SecurityContact: "myContact",
				Details:         "myDetails",
			},
		},
		"partial update: Website": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						"otherWebsite",
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "myMoniker",
				Identity:        "myIdentity",
				Website:         "otherWebsite",
				SecurityContact: "myContact",
				Details:         "myDetails",
			},
		},
		"partial update: Details": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						"otherDetails",
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "myMoniker",
				Identity:        "myIdentity",
				Website:         "https://example.com",
				SecurityContact: "myContact",
				Details:         "otherDetails",
			},
		},
		"partial update: SecurityContact": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription(
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						stakingtypes.DoNotModifyDesc,
						"otherContact",
						stakingtypes.DoNotModifyDesc,
					)
				},
			),
			exp: &contract.ValidatorMetadata{
				Moniker:         "myMoniker",
				Identity:        "myIdentity",
				Website:         "https://example.com",
				SecurityContact: "otherContact",
				Details:         "myDetails",
			},
		},
		"set empty values": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.OperatorAddress = myOperatorAddr.String()
					m.Description = stakingtypes.NewDescription("otherMoniker", "", "", "", "")
				},
			),
			exp: &contract.ValidatorMetadata{Moniker: "otherMoniker"},
		},
		"invalid description": {
			src: types.MsgUpdateValidatorFixture(
				func(m *types.MsgUpdateValidator) {
					m.Description.Moniker = strings.Repeat("x", 71)
				},
			),
			expErr: sdkerrors.ErrInvalidRequest,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			fn, execs := wasmtesting.CaptureExecuteFn()
			km := &wasmtesting.ContractOpsKeeperMock{
				ExecuteFn: fn,
			}
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()).WithEventManager(em))

			// when
			s := NewMsgServerImpl(poeKeeperMock, km, querier)
			gotRes, gotErr := s.UpdateValidator(ctx, spec.src)

			// then
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(gotErr), "exp %v but got %#+v", spec.expErr, gotErr)
				assert.Nil(t, gotRes)
				return
			}
			require.NoError(t, gotErr)
			// and contract called
			assert.Len(t, *execs, 1)
			assert.Equal(t, myValsetContract, (*execs)[0].ContractAddress)
			assert.Equal(t, myOperatorAddr, (*execs)[0].Caller)
			// ensure we pass the right data
			var op contract.TG4ValsetExecute
			require.NoError(t, json.Unmarshal((*execs)[0].Msg, &op))
			assert.Equal(t, spec.exp, op.UpdateMetadata)
			assert.Nil(t, (*execs)[0].Coins)

			// and events emitted
			require.NoError(t, gotErr)
			require.Len(t, em.Events(), 2)
			assert.Equal(t, sdk.EventTypeMessage, em.Events()[0].Type)
			assert.Equal(t, types.EventTypeUpdateValidator, em.Events()[1].Type)
		})
	}
}

func TestDelegate(t *testing.T) {
	var (
		myStakingContract sdk.AccAddress = rand.Bytes(address.Len)
		myOperatorAddr    sdk.AccAddress = rand.Bytes(address.Len)
	)
	poeKeeperMock := PoEKeeperMock{
		GetPoEContractAddressFn: SwitchPoEContractAddressFn(t, nil, myStakingContract),
	}

	specs := map[string]struct {
		src               *types.MsgDelegate
		expSelfDelegation *sdk.Coin
		expErr            *sdkerrors.Error
	}{
		"all good": {
			src:               types.NewMsgDelegate(myOperatorAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt())),
			expSelfDelegation: &sdk.Coin{Denom: types.DefaultBondDenom, Amount: sdk.OneInt()},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			fn, execs := wasmtesting.CaptureExecuteFn()
			km := &wasmtesting.ContractOpsKeeperMock{
				ExecuteFn: fn,
			}
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()).WithEventManager(em))
			// when
			s := NewMsgServerImpl(poeKeeperMock, km, nil)
			gotRes, gotErr := s.Delegate(ctx, spec.src)

			// then
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(gotErr), "exp %v but got %#+v", spec.expErr, gotErr)
				assert.Nil(t, gotRes)
				return
			}
			require.NoError(t, gotErr)
			// and contract called
			require.Len(t, *execs, 1)
			assert.Equal(t, myStakingContract, (*execs)[0].ContractAddress)
			assert.Equal(t, myOperatorAddr, (*execs)[0].Caller)
			assert.Equal(t, sdk.NewCoins(sdk.NewCoin(types.DefaultBondDenom, sdk.OneInt())), (*execs)[0].Coins)

			// and events emitted
			require.NoError(t, gotErr)
			require.Len(t, em.Events(), 2)
			assert.Equal(t, sdk.EventTypeMessage, em.Events()[0].Type)
			assert.Equal(t, types.EventTypeDelegate, em.Events()[1].Type)
			assert.Equal(t, "amount", string(em.Events()[1].Attributes[0].Key))
			assert.Equal(t, "1utgd", string(em.Events()[1].Attributes[0].Value))
		})
	}
}
