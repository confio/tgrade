package poe

import (
	"encoding/json"
	"github.com/confio/tgrade/x/poe/contract"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	"testing"
)

func TestEndBlock(t *testing.T) {
	var (
		capturedSudoCalls []tuple
		myAddr            sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	)

	specs := map[string]struct {
		setup           func(m *MockSudoer)
		expSudoCalls    []tuple
		expPanic        bool
		expCommitted    []bool
		expValsetUpdate []abci.ValidatorUpdate
	}{
		"valset update - empty response": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
					_, err := captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					return &sdk.Result{}, err
				}
				m.IterateContractCallbacksByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expSudoCalls: []tuple{{addr: myAddr, msg: []byte(`{"end_with_validator_update":{}}`)}},
			expCommitted: []bool{true},
		},
		"valset update - empty list": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
					captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					bz, err := json.Marshal(&contract.EndWithValidatorUpdateResponse{})
					require.NoError(t, err)
					return &sdk.Result{Data: bz}, err
				}
				m.IterateContractCallbacksByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expSudoCalls: []tuple{{addr: myAddr, msg: []byte(`{"end_with_validator_update":{}}`)}},
			expCommitted: []bool{true},
		},
		"valset update - response list": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
					captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					bz, err := json.Marshal(&contract.EndWithValidatorUpdateResponse{
						Diffs: []contract.ValidatorUpdate{
							{PubKey: contract.ValidatorPubkey{Ed25519: []byte("my key")}, Power: 1},
							{PubKey: contract.ValidatorPubkey{Ed25519: []byte("my other key")}, Power: 2}},
					})
					require.NoError(t, err)
					return &sdk.Result{Data: bz}, err
				}
				m.IterateContractCallbacksByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expSudoCalls: []tuple{{addr: myAddr, msg: []byte(`{"end_with_validator_update":{}}`)}},
			expCommitted: []bool{true},
			expValsetUpdate: []abci.ValidatorUpdate{{
				PubKey: crypto.PublicKey{Sum: &crypto.PublicKey_Ed25519{Ed25519: []byte("my key")}},
				Power:  1,
			}, {
				PubKey: crypto.PublicKey{Sum: &crypto.PublicKey_Ed25519{Ed25519: []byte("my other key")}},
				Power:  2,
			}},
		},
		"valset update - panic not handled": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
					if contractAddress.Equals(myAddr) {
						panic("testing")
					}
					return captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
				}
				m.IterateContractCallbacksByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expPanic: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedSudoCalls = nil
			mock := MockSudoer{}
			spec.setup(&mock)
			commitMultistore := mockCommitMultiStore{}
			ctx := sdk.Context{}.WithLogger(log.TestingLogger()).
				WithMultiStore(&commitMultistore)

			// when
			if spec.expPanic {
				require.Panics(t, func() {
					_ = EndBlocker(ctx, &mock)
				})
				return
			}
			gotValsetUpdate := EndBlocker(ctx, &mock)
			assert.Equal(t, spec.expValsetUpdate, gotValsetUpdate)

			// then
			require.Len(t, capturedSudoCalls, len(spec.expSudoCalls))
			for i, v := range spec.expSudoCalls {
				require.Equal(t, v.addr, capturedSudoCalls[i].addr)
				exp, got := string(v.msg), string(capturedSudoCalls[i].msg)
				assert.JSONEq(t, exp, got, "expected %q but got %q", exp, got)
			}
			// and tx committed
			for i, v := range spec.expCommitted {
				assert.Equal(t, v, commitMultistore.committed[i], "tx number %d", i)
			}
		})
	}
}

func iterateContractsFn(t *testing.T, expType twasmtypes.PrivilegedCallbackType, addrs ...sdk.AccAddress) func(ctx sdk.Context, callbackType twasmtypes.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	return func(ctx sdk.Context, callbackType twasmtypes.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
		require.Equal(t, expType, callbackType)
		for i, a := range addrs {
			if cb(uint8(i+1), a) {
				return
			}
		}
	}
}

// helper function to handle both types in end block
func endBlockTypeIterateContractsFn(t *testing.T, end []sdk.AccAddress, valset []sdk.AccAddress) func(sdk.Context, twasmtypes.PrivilegedCallbackType, func(uint8, sdk.AccAddress) bool) {
	return func(ctx sdk.Context, callbackType twasmtypes.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
		switch callbackType {
		case twasmtypes.CallbackTypeEndBlock:
			iterateContractsFn(t, twasmtypes.CallbackTypeEndBlock, end...)(ctx, callbackType, cb)
		case twasmtypes.CallbackTypeValidatorSetUpdate:
			iterateContractsFn(t, twasmtypes.CallbackTypeValidatorSetUpdate, valset...)(ctx, callbackType, cb)
		default:
			t.Errorf("unexpected callback type: %q", callbackType.String())
		}
	}
}

type tuple struct {
	addr sdk.AccAddress
	msg  []byte
}

func captureSudos(capturedSudoCalls *[]tuple) func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
	return func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
		*capturedSudoCalls = append(*capturedSudoCalls, tuple{addr: contractAddress, msg: msg})
		return nil, nil
	}
}

type MockSudoer struct {
	SudoFn                           func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
	IterateContractCallbacksByTypeFn func(ctx sdk.Context, callbackType twasmtypes.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func (m MockSudoer) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error) {
	if m.SudoFn == nil {
		panic("not expected to be called")
	}
	return m.SudoFn(ctx, contractAddress, msg)
}

func (m MockSudoer) IterateContractCallbacksByType(ctx sdk.Context, callbackType twasmtypes.PrivilegedCallbackType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	if m.IterateContractCallbacksByTypeFn == nil {
		panic("not expected to be called")
	}
	m.IterateContractCallbacksByTypeFn(ctx, callbackType, cb)
}

type mockCommitMultiStore struct {
	sdk.CommitMultiStore
	committed []bool
}

func (m *mockCommitMultiStore) CacheMultiStore() storetypes.CacheMultiStore {
	m.committed = append(m.committed, false)
	return &mockCMS{m, &m.committed[len(m.committed)-1]}
}

type mockCMS struct {
	sdk.CommitMultiStore
	committed *bool
}

func (m *mockCMS) Write() {
	*m.committed = true
}
