package keeper

import (
	"context"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryPrivilegedContracts(t *testing.T) {
	addr1 := RandomAddress(t)
	addr2 := RandomAddress(t)

	specs := map[string]struct {
		state  []sdk.AccAddress
		expRsp *types.QueryPrivilegedContractsResponse
	}{
		"none found": {
			expRsp: &types.QueryPrivilegedContractsResponse{},
		},
		"single found": {
			state: []sdk.AccAddress{addr1},
			expRsp: &types.QueryPrivilegedContractsResponse{
				Addresses: []string{addr1.String()},
			},
		},
		"multiple found": {
			state: []sdk.AccAddress{addr1, addr2},
			expRsp: &types.QueryPrivilegedContractsResponse{
				Addresses: []string{addr1.String(), addr2.String()},
			},
		},
	}
	ctx := sdk.Context{}.WithContext(context.Background())
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			mock := MockQueryKeeper{
				IteratePrivilegedFn: func(ctx sdk.Context, cb func(sdk.AccAddress) bool) {
					for _, a := range spec.state {
						if cb(a) {
							return
						}
					}
				},
			}

			q := NewGrpcQuerier(mock)
			// when
			gotRsp, gotErr := q.PrivilegedContracts(sdk.WrapSDKContext(ctx), nil)
			// then
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expRsp, gotRsp)
		})
	}
}

func TestQueryContractsByCallbackType(t *testing.T) {
	addr1 := RandomAddress(t)
	addr2 := RandomAddress(t)

	specs := map[string]struct {
		state  []sdk.AccAddress
		src    types.QueryContractsByCallbackTypeRequest
		expRsp *types.QueryContractsByCallbackTypeResponse
		expErr bool
	}{
		"none found": {
			src: types.QueryContractsByCallbackTypeRequest{
				CallbackType: types.CallbackTypeEndBlock.String(),
			},
			expRsp: &types.QueryContractsByCallbackTypeResponse{},
		},
		"single found": {
			src: types.QueryContractsByCallbackTypeRequest{
				CallbackType: types.CallbackTypeEndBlock.String(),
			},
			state: []sdk.AccAddress{addr1},
			expRsp: &types.QueryContractsByCallbackTypeResponse{
				Contracts: []types.QueryContractsByCallbackTypeResponse_ContractPosition{
					{Addresses: addr1.String()},
				},
			},
		},
		"multiple found": {
			src: types.QueryContractsByCallbackTypeRequest{
				CallbackType: types.CallbackTypeEndBlock.String(),
			},
			state: []sdk.AccAddress{addr1, addr2},
			expRsp: &types.QueryContractsByCallbackTypeResponse{
				Contracts: []types.QueryContractsByCallbackTypeResponse_ContractPosition{
					{Addresses: addr1.String()},
					{Addresses: addr2.String()},
				},
			},
		},
		"unknown callback type": {
			src: types.QueryContractsByCallbackTypeRequest{
				CallbackType: "unknown",
			},
			expErr: true,
		},
		"empty callback type": {
			src:    types.QueryContractsByCallbackTypeRequest{},
			expErr: true,
		},
	}
	ctx := sdk.Context{}.WithContext(context.Background())
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			mock := MockQueryKeeper{
				IterateContractCallbacksByTypeFn: func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
					for i, a := range spec.state {
						if cb(uint8(i+1), a) {
							return
						}
					}
				},
			}

			q := NewGrpcQuerier(mock)
			// when
			gotRsp, gotErr := q.ContractsByCallbackType(sdk.WrapSDKContext(ctx), &spec.src)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				assert.Nil(t, gotRsp)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expRsp, gotRsp)
		})
	}
}

type MockQueryKeeper struct {
	IteratePrivilegedFn              func(ctx sdk.Context, cb func(sdk.AccAddress) bool)
	IterateContractCallbacksByTypeFn func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func (m MockQueryKeeper) IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool) {
	if m.IteratePrivilegedFn == nil {
		panic("not expected to be called")
	}
	m.IteratePrivilegedFn(ctx, cb)
}

func (m MockQueryKeeper) IterateContractCallbacksByType(ctx sdk.Context, callbackType types.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	if m.IterateContractCallbacksByTypeFn == nil {
		panic("not expected to be called")
	}
	m.IterateContractCallbacksByTypeFn(ctx, callbackType, cb)
}
