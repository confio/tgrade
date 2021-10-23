package poe

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/poe/types"
)

func TestHandler(t *testing.T) {
	now := time.Now().UTC()
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
		"MsgCreateValidator error returned": {
			src: types.MsgCreateValidatorFixture(),
			mock: MsgServerMock{
				CreateValidatorFn: func(ctx context.Context, validator *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
					return nil, sdkerrors.ErrInvalidAddress
				},
			},
			expErr: sdkerrors.ErrInvalidAddress,
		},
		"MsgUpdateValidator": {
			src: types.MsgUpdateValidatorFixture(),
			mock: MsgServerMock{
				UpdateValidatorFn: func(ctx context.Context, validator *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
					return &types.MsgUpdateValidatorResponse{}, nil
				},
			},
			expResult: &sdk.Result{Data: []byte{}, Events: []abcitypes.Event{}},
		},
		"MsgUpdateValidator error returned": {
			src: types.MsgUpdateValidatorFixture(),
			mock: MsgServerMock{
				UpdateValidatorFn: func(ctx context.Context, validator *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
					return nil, sdkerrors.ErrInvalidAddress
				},
			},
			expErr: sdkerrors.ErrInvalidAddress,
		},
		"MsgDelegate": {
			src: &types.MsgDelegate{},
			mock: MsgServerMock{
				DelegateFn: func(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
					return &types.MsgDelegateResponse{}, nil
				},
			},
			expResult: &sdk.Result{Data: []byte{}, Events: []abcitypes.Event{}},
		},
		"MsgDelegate error returned": {
			src: &types.MsgDelegate{},
			mock: MsgServerMock{
				DelegateFn: func(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
					return nil, types.ErrInvalid
				},
			},
			expErr: types.ErrInvalid,
		},
		"MsgUndelegate": {
			src: &types.MsgUndelegate{},
			mock: MsgServerMock{
				UndelegateFn: func(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
					return &types.MsgUndelegateResponse{
						CompletionTime: now,
					}, nil
				},
			},
			expResult: &sdk.Result{Data: mustMarshalProto(&types.MsgUndelegateResponse{CompletionTime: now}), Events: []abcitypes.Event{}},
		},
		"MsgUndelegate error returned": {
			src: &types.MsgUndelegate{},
			mock: MsgServerMock{
				UndelegateFn: func(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
					return nil, types.ErrInvalid
				},
			},
			expErr: types.ErrInvalid,
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

func mustMarshalProto(m *types.MsgUndelegateResponse) []byte {
	r, err := m.Marshal()
	if err != nil {
		panic(err)
	}
	return r
}

var _ types.MsgServer = MsgServerMock{}

type MsgServerMock struct {
	CreateValidatorFn func(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error)
	UpdateValidatorFn func(ctx context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error)
	DelegateFn        func(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error)
	UndelegateFn      func(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error)
}

func (m MsgServerMock) CreateValidator(ctx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	if m.CreateValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.CreateValidatorFn(ctx, msg)
}

func (m MsgServerMock) UpdateValidator(ctx context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
	if m.UpdateValidatorFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateValidatorFn(ctx, msg)
}

func (m MsgServerMock) Delegate(ctx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	if m.DelegateFn == nil {
		panic("not expected to be called")
	}
	return m.DelegateFn(ctx, msg)
}

func (m MsgServerMock) Undelegate(ctx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	if m.UndelegateFn == nil {
		panic("not expected to be called")
	}
	return m.UndelegateFn(ctx, msg)
}
