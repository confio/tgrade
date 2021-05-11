package keeper

import (
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestGovHandler(t *testing.T) {
	var (
		myAddr                sdk.AccAddress = rand.Bytes(sdk.AddrLen)
		capturedContractAddrs []sdk.AccAddress
	)
	notHandler := func(ctx sdk.Context, content govtypes.Content) error {
		return sdkerrors.ErrUnknownRequest
	}

	specs := map[string]struct {
		wasmHandler           govtypes.Handler
		setupGovKeeper        func(*MockGovKeeper)
		srcProposal           govtypes.Content
		expErr                *sdkerrors.Error
		expCapturedAddrs      []sdk.AccAddress
		expCapturedGovContent []govtypes.Content
	}{
		"handled in wasm": {
			wasmHandler: func(ctx sdk.Context, content govtypes.Content) error {
				return nil
			},
		},
		"fails in wasm": {
			wasmHandler: func(ctx sdk.Context, content govtypes.Content) error {
				return sdkerrors.ErrJSONMarshal
			},
			expErr: sdkerrors.ErrJSONMarshal,
		},
		"not handled": {
			wasmHandler: notHandler,
			srcProposal: &govtypes.TextProposal{},
			expErr:      sdkerrors.ErrUnknownRequest,
		},
		"promote proposal": {
			wasmHandler: notHandler,
			setupGovKeeper: func(m *MockGovKeeper) {
				m.SetPrivilegedFn = func(ctx sdk.Context, contractAddr sdk.AccAddress) error {
					capturedContractAddrs = append(capturedContractAddrs, contractAddr)
					return nil
				}
			},
			srcProposal: types.PromoteProposalFixture(func(proposal *types.PromoteToPrivilegedContractProposal) {
				proposal.Contract = myAddr.String()
			}),
			expCapturedAddrs: []sdk.AccAddress{myAddr},
		},
		"invalid promote proposal rejected": {
			wasmHandler: notHandler,
			srcProposal: &types.PromoteToPrivilegedContractProposal{},
			expErr:      govtypes.ErrInvalidProposalContent,
		},
		"demote proposal": {
			wasmHandler: notHandler,
			setupGovKeeper: func(m *MockGovKeeper) {
				m.UnsetPrivilegedFn = func(ctx sdk.Context, contractAddr sdk.AccAddress) error {
					capturedContractAddrs = append(capturedContractAddrs, contractAddr)
					return nil
				}
			},
			srcProposal: types.DemoteProposalFixture(func(proposal *types.DemotePrivilegedContractProposal) {
				proposal.Contract = myAddr.String()
			}),
			expCapturedAddrs: []sdk.AccAddress{myAddr},
		},
		"invalid demote proposal rejected": {
			wasmHandler: notHandler,
			srcProposal: &types.DemotePrivilegedContractProposal{},
			expErr:      govtypes.ErrInvalidProposalContent,
		},
		"nil content": {
			wasmHandler: notHandler,
			expErr:      sdkerrors.ErrUnknownRequest,
		},
	}
	var ctx sdk.Context
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedContractAddrs = nil
			var mock MockGovKeeper
			if spec.setupGovKeeper != nil {
				spec.setupGovKeeper(&mock)
			}
			// when
			router := &CapturingGovRouter{}
			h := NewProposalHandlerX(&mock, spec.wasmHandler)
			gotErr := h(ctx, spec.srcProposal)
			// then
			require.True(t, spec.expErr.Is(gotErr), "exp %v but got #+v", spec.expErr, gotErr)
			assert.Equal(t, spec.expCapturedAddrs, capturedContractAddrs)
			assert.Equal(t, spec.expCapturedGovContent, router.captured)
		})
	}
}

type MockGovKeeper struct {
	SetPrivilegedFn   func(ctx sdk.Context, contractAddr sdk.AccAddress) error
	UnsetPrivilegedFn func(ctx sdk.Context, contractAddr sdk.AccAddress) error
}

func (m MockGovKeeper) SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	if m.SetPrivilegedFn == nil {
		panic("not expected to be called")
	}
	return m.SetPrivilegedFn(ctx, contractAddr)
}

func (m MockGovKeeper) UnsetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	if m.UnsetPrivilegedFn == nil {
		panic("not expected to be called")
	}
	return m.UnsetPrivilegedFn(ctx, contractAddr)
}

type CapturingGovRouter struct {
	govtypes.Router
	captured []govtypes.Content
}

func (m CapturingGovRouter) HasRoute(r string) bool {
	return true
}

func (m *CapturingGovRouter) GetRoute(path string) (h govtypes.Handler) {
	return func(ctx sdk.Context, content govtypes.Content) error {
		m.captured = append(m.captured, content)
		return nil
	}
}
