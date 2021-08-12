package poe

import (
	"context"
	"github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestHandler(t *testing.T) {
	specs := map[string]struct {
		src       sdk.Msg
		mock      MsgServerMock
		expErr    *sdkerrors.Error
		expResult *sdk.Result
	}{
		"MsgCreateValidator": {
			src: types.MsgCreateValidatorFixture(),
			mock: MsgServerMock{
				CreateValidatorFn: func(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
					return &types.MsgCreateValidatorResponse{}, nil
				},
			},
			expResult: &sdk.Result{Data: []byte{}, Events: []abcitypes.Event{}},
		},
		"MsgCreateValidator with events": {
			src: types.MsgCreateValidatorFixture(),
			mock: MsgServerMock{
				CreateValidatorFn: func(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
					sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.Event{Type: "foo"})
					return nil, nil
				},
			},
			expResult: &sdk.Result{Data: []byte{}, Events: []abcitypes.Event{{
				Type: "foo",
			}}},
		},
		"error returned": {
			src: types.MsgCreateValidatorFixture(),
			mock: MsgServerMock{
				CreateValidatorFn: func(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
					return nil, sdkerrors.ErrInvalidAddress
				},
			},
			expErr: sdkerrors.ErrInvalidAddress,
		},
		"unknown message": {
			src:    &banktypes.MsgSend{},
			expErr: sdkerrors.ErrUnknownRequest,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			h := newHandler(spec.mock)
			ctx := sdk.Context{}.WithContext(context.Background())
			gotRes, gotErr := h(ctx, spec.src)
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(gotErr), "exp %v but got %#+v", spec.expErr, gotErr)
				assert.Nil(t, gotRes)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expResult, gotRes)
		})
	}
}

var _ types.MsgServer = MsgServerMock{}

type MsgServerMock struct {
	CreateValidatorFn func(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error)
	UpdateValidatorFn func(ctx context.Context, validator *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error)
}

func (m MsgServerMock) CreateValidator(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	if m.CreateValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.CreateValidatorFn(ctx, validator)
}

func (m MsgServerMock) UpdateValidator(ctx context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
	if m.UpdateValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateValidatorFn(ctx, msg)
}
