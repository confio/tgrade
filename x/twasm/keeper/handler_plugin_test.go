package keeper

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTgradeHandlesDispatchMsg(t *testing.T) {
	contractAddr := RandomAddress(t)
	specs := map[string]struct {
		setup  func(m *handlerTgradeKeeperMock)
		src    wasmvmtypes.CosmosMsg
		expErr *sdkerrors.Error
	}{
		"handle hook msg": {
			src: wasmvmtypes.CosmosMsg{
				Custom: []byte(`{"hooks":{"register_begin_block":{}}}`),
			},
			setup: func(m *handlerTgradeKeeperMock) { // noopRegisterHook
				noopRegisterHook(m)
			},
		},
		"non custom msg rejected": {
			src:    wasmvmtypes.CosmosMsg{},
			setup:  func(m *handlerTgradeKeeperMock) {},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"non privileged contracts rejected": {
			src: wasmvmtypes.CosmosMsg{Custom: []byte(`{}`)},
			setup: func(m *handlerTgradeKeeperMock) {
				m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
					return false
				}
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
		"invalid json rejected": {
			src: wasmvmtypes.CosmosMsg{Custom: []byte(`not json`)},
			setup: func(m *handlerTgradeKeeperMock) {
				m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
					return true
				}
			},
			expErr: sdkerrors.ErrJSONUnmarshal,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			mock := handlerTgradeKeeperMock{}
			spec.setup(&mock)
			h := NewTgradeHandler(mock)
			var ctx sdk.Context
			_, _, gotErr := h.DispatchMsg(ctx, contractAddr, "", spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
		})
	}
}

type registration struct {
	cb   types.PrivilegedCallbackType
	addr sdk.AccAddress
}
type unregistration struct {
	cb   types.PrivilegedCallbackType
	pos  uint8
	addr sdk.AccAddress
}

func TestTgradeHandlesHooks(t *testing.T) {
	myContractAddr := RandomAddress(t)

	var capturedDetails *types.TgradeContractDetails
	captureContractDetails := func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
		require.Equal(t, myContractAddr, contract)
		capturedDetails = details
		return nil
	}

	var capturedRegistrations []registration
	captureRegistrations := func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) uint8 {
		capturedRegistrations = append(capturedRegistrations, registration{cb: callbackType, addr: contractAddress})
		return 1
	}
	var capturedUnRegistrations []unregistration
	captureUnRegistrations := func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddress sdk.AccAddress) bool {
		capturedUnRegistrations = append(capturedUnRegistrations, unregistration{cb: callbackType, pos: pos, addr: contractAddress})
		return true
	}

	captureWithMock := func(mutators ...func(*wasmtypes.ContractInfo)) func(mock *handlerTgradeKeeperMock) {
		return func(m *handlerTgradeKeeperMock) {
			m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
				f := wasmtypes.ContractInfoFixture(mutators...)
				return &f
			}
			m.setContractDetailsFn = captureContractDetails
			m.appendToPrivilegedContractCallbacksFn = captureRegistrations
			m.removePrivilegedContractCallbacksFn = captureUnRegistrations
		}
	}

	specs := map[string]struct {
		setup              func(m *handlerTgradeKeeperMock)
		src                contract.Hooks
		expDetails         *types.TgradeContractDetails
		expRegistrations   []registration
		expUnRegistrations []unregistration
		expErr             *sdkerrors.Error
	}{
		"register begin block": {
			src:   contract.Hooks{RegisterBeginBlock: &struct{}{}},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "begin_block"}},
			},
			expRegistrations: []registration{{cb: types.CallbackTypeBeginBlock, addr: myContractAddr}},
		},
		"register end block": {
			src:   contract.Hooks{RegisterEndBlock: &struct{}{}},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "end_block"}},
			},
			expRegistrations: []registration{{cb: types.CallbackTypeEndBlock, addr: myContractAddr}},
		},
		"register validator set update block": {
			src:   contract.Hooks{RegisterValidatorSetUpdate: &struct{}{}},
			setup: captureWithMock(),
			expDetails: &types.TgradeContractDetails{
				RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "validator_set_update"}},
			},
			expRegistrations: []registration{{cb: types.CallbackTypeValidatorSetUpdate, addr: myContractAddr}},
		},
		"unregister begin block": {
			src: contract.Hooks{UnregisterBeginBlock: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "begin_block"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredCallbacks: []*types.RegisteredCallback{}},
			expUnRegistrations: []unregistration{{cb: types.CallbackTypeBeginBlock, pos: 1, addr: myContractAddr}},
		},
		"unregister end block": {
			src: contract.Hooks{UnregisterEndBlock: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "end_block"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredCallbacks: []*types.RegisteredCallback{}},
			expUnRegistrations: []unregistration{{cb: types.CallbackTypeEndBlock, pos: 1, addr: myContractAddr}},
		},
		"unregister validator set update block": {
			src: contract.Hooks{UnregisterValidatorSetUpdate: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "validator_set_update"}},
				}
				info.SetExtension(ext)
			}),
			expDetails:         &types.TgradeContractDetails{RegisteredCallbacks: []*types.RegisteredCallback{}},
			expUnRegistrations: []unregistration{{cb: types.CallbackTypeValidatorSetUpdate, pos: 1, addr: myContractAddr}},
		},
		"register begin block with existing registration": {
			src: contract.Hooks{RegisterBeginBlock: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				info.SetExtension(&types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "begin_block"}},
				})
			}),
		},
		"register appends to existing callback list": {
			src: contract.Hooks{RegisterBeginBlock: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				info.SetExtension(&types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 100, CallbackType: "end_block"}},
				})
			}),
			expDetails: &types.TgradeContractDetails{
				RegisteredCallbacks: []*types.RegisteredCallback{{Position: 100, CallbackType: "end_block"}, {Position: 1, CallbackType: "begin_block"}},
			},
			expRegistrations: []registration{{cb: types.CallbackTypeBeginBlock, addr: myContractAddr}},
		},
		"unregister removed from existing callback list": {
			src: contract.Hooks{UnregisterBeginBlock: &struct{}{}},
			setup: captureWithMock(func(info *wasmtypes.ContractInfo) {
				ext := &types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{
						{Position: 3, CallbackType: "validator_set_update"},
						{Position: 1, CallbackType: "begin_block"},
						{Position: 100, CallbackType: "end_block"},
					},
				}
				info.SetExtension(ext)
			}),
			expDetails: &types.TgradeContractDetails{RegisteredCallbacks: []*types.RegisteredCallback{
				{Position: 3, CallbackType: "validator_set_update"},
				{Position: 100, CallbackType: "end_block"},
			}},
			expUnRegistrations: []unregistration{{cb: types.CallbackTypeBeginBlock, pos: 1, addr: myContractAddr}},
		},

		"unregister begin block with without existing registration": {
			src:   contract.Hooks{UnregisterBeginBlock: &struct{}{}},
			setup: captureWithMock(),
		},
		"empty hook msg rejected": {
			setup: func(m *handlerTgradeKeeperMock) {
				noopRegisterHook(m)
			},
			expErr: wasmtypes.ErrUnknownMsg,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedDetails, capturedRegistrations, capturedUnRegistrations = nil, nil, nil
			mock := handlerTgradeKeeperMock{}
			spec.setup(&mock)
			h := NewTgradeHandler(mock)
			var ctx sdk.Context
			gotErr := h.handleHooks(ctx, myContractAddr, &spec.src)
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
			if spec.expErr != nil {
				return
			}
			assert.Equal(t, spec.expDetails, capturedDetails)
			assert.Equal(t, spec.expRegistrations, capturedRegistrations)
			assert.Equal(t, spec.expUnRegistrations, capturedUnRegistrations)
		})
	}
}

// noopRegisterHook does nothing and but all methods for registration
func noopRegisterHook(m *handlerTgradeKeeperMock) {
	m.IsPrivilegedFn = func(ctx sdk.Context, contract sdk.AccAddress) bool {
		return true
	}
	m.GetContractInfoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
		v := wasmtypes.ContractInfoFixture(func(info *wasmtypes.ContractInfo) {
			info.SetExtension(&types.TgradeContractDetails{})
		})
		return &v
	}
	m.appendToPrivilegedContractCallbacksFn = func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) uint8 {
		return 1
	}
	m.setContractDetailsFn = func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
		return nil
	}
}

type handlerTgradeKeeperMock struct {
	IsPrivilegedFn                        func(ctx sdk.Context, contract sdk.AccAddress) bool
	appendToPrivilegedContractCallbacksFn func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) uint8
	removePrivilegedContractCallbacksFn   func(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) bool
	setContractDetailsFn                  func(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error
	GetContractInfoFn                     func(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

func (m handlerTgradeKeeperMock) IsPrivileged(ctx sdk.Context, contract sdk.AccAddress) bool {
	if m.IsPrivilegedFn == nil {
		panic("not expected to be called")
	}
	return m.IsPrivilegedFn(ctx, contract)
}

func (m handlerTgradeKeeperMock) appendToPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) uint8 {
	if m.appendToPrivilegedContractCallbacksFn == nil {
		panic("not expected to be called")
	}
	return m.appendToPrivilegedContractCallbacksFn(ctx, callbackType, contractAddress)
}

func (m handlerTgradeKeeperMock) removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) bool {
	if m.removePrivilegedContractCallbacksFn == nil {
		panic("not expected to be called")
	}
	return m.removePrivilegedContractCallbacksFn(ctx, callbackType, pos, contractAddr)
}

func (m handlerTgradeKeeperMock) setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
	if m.setContractDetailsFn == nil {
		panic("not expected to be called")
	}
	return m.setContractDetailsFn(ctx, contract, details)
}

func (m handlerTgradeKeeperMock) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo {
	if m.GetContractInfoFn == nil {
		panic("not expected to be called")
	}
	return m.GetContractInfoFn(ctx, contractAddress)
}
