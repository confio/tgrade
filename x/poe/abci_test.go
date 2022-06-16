package poe

import (
	"encoding/json"
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/confio/tgrade/x/poe/contract"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

func TestEndBlock(t *testing.T) {
	var (
		capturedSudoCalls []tuple
		myAddr            sdk.AccAddress = rand.Bytes(address.Len)
	)

	specs := map[string]struct {
		setup           func(m *MockSudoer)
		expSudoCalls    []tuple
		expCommitted    []bool
		expValsetUpdate []abci.ValidatorUpdate
	}{
		"valset update - empty response": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
					_, err := captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					return []byte{}, err
				}
				m.IteratePrivilegedContractsByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expSudoCalls: []tuple{{addr: myAddr, msg: []byte(`{"end_with_validator_update":{}}`)}},
			expCommitted: []bool{true},
		},
		"valset update - empty list": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
					captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					bz, err := json.Marshal(&contract.EndWithValidatorUpdateResponse{})
					require.NoError(t, err)
					return bz, err
				}
				m.IteratePrivilegedContractsByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
			expSudoCalls: []tuple{{addr: myAddr, msg: []byte(`{"end_with_validator_update":{}}`)}},
			expCommitted: []bool{true},
		},
		"valset update - response list": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
					captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
					bz, err := json.Marshal(&contract.EndWithValidatorUpdateResponse{
						Diffs: []contract.ValidatorUpdate{
							{PubKey: contract.ValidatorPubkey{Ed25519: []byte("my key")}, Power: 1},
							{PubKey: contract.ValidatorPubkey{Ed25519: []byte("my other key")}, Power: 2},
						},
					})
					require.NoError(t, err)
					return bz, err
				}
				m.IteratePrivilegedContractsByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
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
		"valset update - panic should be handled": {
			setup: func(m *MockSudoer) {
				m.SudoFn = func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
					if contractAddress.Equals(myAddr) {
						panic("testing")
					}
					return captureSudos(&capturedSudoCalls)(ctx, contractAddress, msg)
				}
				m.IteratePrivilegedContractsByTypeFn = endBlockTypeIterateContractsFn(t, nil, []sdk.AccAddress{myAddr})
			},
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

func iterateContractsFn(t *testing.T, expType twasmtypes.PrivilegeType, addrs ...sdk.AccAddress) func(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	return func(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
		require.Equal(t, expType, privilegeType)
		for i, a := range addrs {
			if cb(uint8(i+1), a) {
				return
			}
		}
	}
}

// helper function to handle both types in end block
func endBlockTypeIterateContractsFn(t *testing.T, end []sdk.AccAddress, valset []sdk.AccAddress) func(sdk.Context, twasmtypes.PrivilegeType, func(uint8, sdk.AccAddress) bool) {
	return func(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
		switch privilegeType {
		case twasmtypes.PrivilegeTypeEndBlock:
			iterateContractsFn(t, twasmtypes.PrivilegeTypeEndBlock, end...)(ctx, privilegeType, cb)
		case twasmtypes.PrivilegeTypeValidatorSetUpdate:
			iterateContractsFn(t, twasmtypes.PrivilegeTypeValidatorSetUpdate, valset...)(ctx, privilegeType, cb)
		default:
			t.Errorf("unexpected privileged type: %q", privilegeType.String())
		}
	}
}

type tuple struct {
	addr sdk.AccAddress
	msg  []byte
}

func captureSudos(capturedSudoCalls *[]tuple) func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	return func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
		*capturedSudoCalls = append(*capturedSudoCalls, tuple{addr: contractAddress, msg: msg})
		return nil, nil
	}
}

type MockSudoer struct {
	SudoFn                             func(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
	IteratePrivilegedContractsByTypeFn func(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool)
}

func (m MockSudoer) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if m.SudoFn == nil {
		panic("not expected to be called")
	}
	return m.SudoFn(ctx, contractAddress, msg)
}

func (m MockSudoer) IteratePrivilegedContractsByType(ctx sdk.Context, privilegeType twasmtypes.PrivilegeType, cb func(prio uint8, contractAddr sdk.AccAddress) bool) {
	if m.IteratePrivilegedContractsByTypeFn == nil {
		panic("not expected to be called")
	}
	m.IteratePrivilegedContractsByTypeFn(ctx, privilegeType, cb)
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
