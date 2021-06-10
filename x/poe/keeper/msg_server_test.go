package keeper

import (
	"context"
	"github.com/confio/tgrade/x/poe/types"
	wasmtesting "github.com/confio/tgrade/x/twasm/testing"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/rand"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestCreateValidator(t *testing.T) {
	var myValsetContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myStakingContract sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	var myDelegatorAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	specs := map[string]struct {
		src    *types.MsgCreateValidator
		expErr *sdkerrors.Error
	}{
		"all good": {
			src: types.MsgCreateValidatorFixture(
				func(m *types.MsgCreateValidator) {
					m.DelegatorAddress = myDelegatorAddr.String()
					m.Value = sdk.NewInt64Coin(types.DefaultBondDenom, 1)
				},
			),
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
			cm := ContractSourceMock{
				GetPoEContractAddressFn: func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
					switch ctype {
					case types.PoEContractTypeValset:
						return myValsetContract, nil
					case types.PoEContractTypeStaking:
						return myStakingContract, nil
					default:
						t.Fatalf("unexpected type: %s", ctype)
						return nil, nil
					}
				},
			}
			fn, execs := wasmtesting.CaptureExecuteFn()
			km := &wasmtesting.ContractOpsKeeperMock{
				ExecuteFn: fn,
			}
			em := sdk.NewEventManager()
			ctx := sdk.WrapSDKContext(sdk.Context{}.WithContext(context.Background()).WithEventManager(em).WithConsensusParams(&abci.ConsensusParams{
				Validator: &types1.ValidatorParams{PubKeyTypes: []string{"ed25519"}}}))

			// when
			s := NewMsgServerImpl(cm, km)
			gotRes, gotErr := s.CreateValidator(ctx, spec.src)

			// then
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(gotErr), "exp %v but got %#+v", spec.expErr, gotErr)
				assert.Nil(t, gotRes)
				return
			}
			require.NoError(t, gotErr)
			// and contract called
			assert.Len(t, *execs, 2)
			assert.Equal(t, myValsetContract, (*execs)[0].ContractAddress)
			assert.Equal(t, myDelegatorAddr, (*execs)[0].Caller)
			assert.Nil(t, (*execs)[0].Coins)
			assert.Equal(t, myStakingContract, (*execs)[1].ContractAddress)
			assert.Equal(t, myDelegatorAddr, (*execs)[1].Caller)
			assert.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.DefaultBondDenom, 1)), (*execs)[1].Coins)

			// and events emitted
			require.NoError(t, gotErr)
			require.Len(t, em.Events(), 2)
			assert.Equal(t, types.EventTypeCreateValidator, em.Events()[0].Type)
			assert.Equal(t, sdk.EventTypeMessage, em.Events()[1].Type)
		})
	}

}
