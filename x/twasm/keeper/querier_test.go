package keeper

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/twasm/types"
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
				Contracts: []string{addr1.String()},
			},
		},
		"multiple found": {
			state: []sdk.AccAddress{addr1, addr2},
			expRsp: &types.QueryPrivilegedContractsResponse{
				Contracts: []string{addr1.String(), addr2.String()},
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

			q := NewQuerier(mock)
			// when
			gotRsp, gotErr := q.PrivilegedContracts(sdk.WrapSDKContext(ctx), nil)
			// then
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expRsp, gotRsp)
		})
	}
}

func TestQueryContractsByPrivilegeType(t *testing.T) {
	addr1 := RandomAddress(t)
	addr2 := RandomAddress(t)

	specs := map[string]struct {
		state  []sdk.AccAddress
		src    types.QueryContractsByPrivilegeTypeRequest
		expRsp *types.QueryContractsByPrivilegeTypeResponse
		expErr bool
	}{
		"none found": {
			src: types.QueryContractsByPrivilegeTypeRequest{
				PrivilegeType: types.PrivilegeTypeEndBlock.String(),
			},
			expRsp: &types.QueryContractsByPrivilegeTypeResponse{},
		},
		"single found": {
			src: types.QueryContractsByPrivilegeTypeRequest{
				PrivilegeType: types.PrivilegeTypeEndBlock.String(),
			},
			state: []sdk.AccAddress{addr1},
			expRsp: &types.QueryContractsByPrivilegeTypeResponse{
				Contracts: []string{addr1.String()},
			},
		},
		"multiple found": {
			src: types.QueryContractsByPrivilegeTypeRequest{
				PrivilegeType: types.PrivilegeTypeEndBlock.String(),
			},
			state: []sdk.AccAddress{addr1, addr2},
			expRsp: &types.QueryContractsByPrivilegeTypeResponse{
				Contracts: []string{addr1.String(), addr2.String()},
			},
		},
		"unknown privilege type": {
			src: types.QueryContractsByPrivilegeTypeRequest{
				PrivilegeType: "unknown",
			},
			expErr: true,
		},
		"empty privilege type": {
			src:    types.QueryContractsByPrivilegeTypeRequest{},
			expErr: true,
		},
	}
	ctx := sdk.Context{}.WithContext(context.Background())
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			mock := MockQueryKeeper{
				IterateContractCallbacksByTypeFn: func(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
					for i, a := range spec.state {
						if cb(uint8(i+1), a) {
							return
						}
					}
				},
			}

			q := NewQuerier(mock)
			// when
			gotRsp, gotErr := q.ContractsByPrivilegeType(sdk.WrapSDKContext(ctx), &spec.src)
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
	IterateContractCallbacksByTypeFn func(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func (m MockQueryKeeper) IteratePrivileged(ctx sdk.Context, cb func(sdk.AccAddress) bool) {
	if m.IteratePrivilegedFn == nil {
		panic("not expected to be called")
	}
	m.IteratePrivilegedFn(ctx, cb)
}

func (m MockQueryKeeper) IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType types.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	if m.IterateContractCallbacksByTypeFn == nil {
		panic("not expected to be called")
	}
	m.IterateContractCallbacksByTypeFn(ctx, privilegeType, cb)
}
