package keeper

import (
	"bytes"
	"errors"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	cosmwasm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

func TestSetPrivileged(t *testing.T) {
	var (
		capturedPinChecksum  *cosmwasm.Checksum
		capturedSudoChecksum *cosmwasm.Checksum
		capturedSudoMsg      []byte
	)
	specs := map[string]struct {
		setup  func(*wasmtesting.MockWasmer)
		expErr bool
	}{
		"all good": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.PinFn = func(checksum cosmwasm.Checksum) error {
					capturedPinChecksum = &checksum
					return nil
				}
				mock.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					capturedSudoChecksum = &codeID
					capturedSudoMsg = sudoMsg
					return &wasmvmtypes.Response{}, 0, nil
				}
			},
		},
		"pin failed": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.PinFn = func(checksum cosmwasm.Checksum) error {
					return errors.New("test, ignore")
				}
			},
			expErr: true,
		},
		"sudo msg failed": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.PinFn = func(checksum cosmwasm.Checksum) error {
					return nil
				}
				mock.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					return nil, 0, errors.New("test, ignore")
				}
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedPinChecksum, capturedSudoMsg, capturedSudoMsg = nil, nil, nil
			mock := NewWasmVMMock()
			spec.setup(mock)

			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(mock))
			k := keepers.TWasmKeeper

			codeID, contractAddr := seedTestContract(t, ctx, k)

			// when
			err := k.SetPrivileged(ctx, contractAddr)

			// then
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var expChecksum cosmwasm.Checksum = k.GetCodeInfo(ctx, codeID).CodeHash

			// then expect pinned to cache
			assert.Equal(t, expChecksum, *capturedPinChecksum)
			// and flag set
			assert.True(t, k.IsPrivileged(ctx, contractAddr))
			// and sudo called
			assert.Equal(t, expChecksum, *capturedSudoChecksum)
			assert.JSONEq(t, `{"privilege_change":{"promoted":{}}}`, string(capturedSudoMsg), "got %s", string(capturedSudoMsg))
		})
	}
}

func TestUnsetPrivileged(t *testing.T) {
	var (
		capturedUnpinChecksum *cosmwasm.Checksum
		capturedSudoChecksum  *cosmwasm.Checksum
		capturedSudoMsg       []byte
	)

	specs := map[string]struct {
		setup  func(*wasmtesting.MockWasmer)
		expErr bool
	}{
		"all good": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.UnpinFn = func(checksum cosmwasm.Checksum) error {
					capturedUnpinChecksum = &checksum
					return nil
				}
				mock.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					capturedSudoChecksum = &codeID
					capturedSudoMsg = sudoMsg
					return &wasmvmtypes.Response{}, 0, nil
				}
			},
		},
		"unpin failed": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.UnpinFn = func(checksum cosmwasm.Checksum) error {
					return errors.New("test, ignore")
				}
				mock.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					return &wasmvmtypes.Response{}, 0, nil
				}
			},
			expErr: true,
		},
		"sudo failed": {
			setup: func(mock *wasmtesting.MockWasmer) {
				mock.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					return nil, 0, errors.New("test, ignore")
				}
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			capturedUnpinChecksum, capturedSudoMsg, capturedSudoMsg = nil, nil, nil
			mock := NewWasmVMMock()
			spec.setup(mock)

			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(mock))
			k := keepers.TWasmKeeper
			codeID, contractAddr := seedTestContract(t, ctx, k)

			// and privileged with a callback
			k.setPrivilegedFlag(ctx, contractAddr)
			k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeEndBlock, contractAddr)
			k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, contractAddr)

			// when
			err := k.UnsetPrivileged(ctx, contractAddr)

			// then
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var expChecksum cosmwasm.Checksum = k.GetCodeInfo(ctx, codeID).CodeHash

			// then expect pinned to cache
			assert.Equal(t, expChecksum, *capturedUnpinChecksum)
			// and flag set
			assert.False(t, k.IsPrivileged(ctx, contractAddr))
			// and callbacks removed
			assert.False(t, k.ExistsPrivilegedContractCallbacks(ctx, types.CallbackTypeEndBlock))
			assert.False(t, k.ExistsPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock))
			// and sudo called
			assert.Equal(t, expChecksum, *capturedSudoChecksum)
			assert.JSONEq(t, `{"privilege_change":{"demoted":{}}}`, string(capturedSudoMsg), "got %s", string(capturedSudoMsg))
		})
	}
}

func TestIteratePrivileged(t *testing.T) {
	ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(NewWasmVMMock()))
	k := keepers.TWasmKeeper

	var (
		addr1 = sdk.AccAddress(bytes.Repeat([]byte{1}, sdk.AddrLen))
		addr2 = sdk.AccAddress(bytes.Repeat([]byte{2}, sdk.AddrLen))
		addr3 = sdk.AccAddress(bytes.Repeat([]byte{3}, sdk.AddrLen))
	)
	for _, a := range []sdk.AccAddress{addr2, addr1, addr3} {
		k.setPrivilegedFlag(ctx, a)
	}

	var captured []sdk.AccAddress

	specs := map[string]struct {
		callback func(addr sdk.AccAddress) bool
		exp      []sdk.AccAddress
	}{
		"capture all": {
			callback: func(addr sdk.AccAddress) bool {
				captured = append(captured, addr)
				return false
			},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
		"capture first": {
			callback: func(addr sdk.AccAddress) bool {
				captured = append(captured, addr)
				return true
			},
			exp: []sdk.AccAddress{addr1},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			captured = nil
			// when
			k.IteratePrivileged(ctx, spec.callback)
			assert.Equal(t, spec.exp, captured)
		})
	}

}
func TestAppendToPrivilegedContractCallbacks(t *testing.T) {
	var (
		addr1 = sdk.AccAddress(bytes.Repeat([]byte{1}, sdk.AddrLen))
		addr2 = sdk.AccAddress(bytes.Repeat([]byte{2}, sdk.AddrLen))
		addr3 = sdk.AccAddress(bytes.Repeat([]byte{3}, sdk.AddrLen))
	)

	type tuple struct {
		a sdk.AccAddress
		p uint8
	}

	specs := map[string]struct {
		setup        func(sdk.Context, *Keeper)
		expPersisted []tuple
	}{
		"first callback": {
			setup:        func(ctx sdk.Context, k *Keeper) {},
			expPersisted: []tuple{{p: 1, a: addr1}},
		},
		"second callback - ordered by position": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, addr3)
			},
			expPersisted: []tuple{{p: 1, a: addr3}, {p: 2, a: addr1}},
		},
		"second callback with same address": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, addr1)
			},
			expPersisted: []tuple{{p: 1, a: addr1}, {p: 2, a: addr1}},
		},
		"other callback type - separate group": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeEndBlock, addr2)
			},
			expPersisted: []tuple{{p: 1, a: addr1}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(NewWasmVMMock()))
			k := keepers.TWasmKeeper
			spec.setup(ctx, k)
			// when
			k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, addr1)

			// then
			var captured []tuple
			k.IterateContractCallbacksByType(ctx, types.CallbackTypeBeginBlock, func(prio uint8, contractAddr sdk.AccAddress) bool {
				captured = append(captured, tuple{p: prio, a: contractAddr})
				return false
			})
			assert.Equal(t, spec.expPersisted, captured)
		})
	}
}

func TestRemovePrivilegedContractCallbacks(t *testing.T) {
	t.Skip("temporary deactivated during refactorings")
	var (
		myAddr      = sdk.AccAddress(bytes.Repeat([]byte{1}, sdk.AddrLen))
		otherAddr   = sdk.AccAddress(bytes.Repeat([]byte{2}, sdk.AddrLen))
		anotheraddr = sdk.AccAddress(bytes.Repeat([]byte{3}, sdk.AddrLen))
	)

	type tuple struct {
		a sdk.AccAddress
		p uint8
	}

	specs := map[string]struct {
		setup        func(sdk.Context, *Keeper)
		expRemaining []tuple
	}{
		"one callback": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
			},
		},
		"multiple callback - same address": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
			},
		},
		"multiple callback - different address": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, otherAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, anotheraddr)
			},
			expRemaining: []tuple{{p: 1, a: otherAddr}, {p: 3, a: anotheraddr}},
		},
		"different address only": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, otherAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, anotheraddr)
			},
			expRemaining: []tuple{{p: 1, a: otherAddr}, {p: 2, a: anotheraddr}},
		},
		"no callbacks": {
			setup: func(ctx sdk.Context, k *Keeper) {},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(NewWasmVMMock()))
			k := keepers.TWasmKeeper
			spec.setup(ctx, k)

			// when
			//k.removePrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)

			// then
			var captured []tuple
			k.IterateContractCallbacksByType(ctx, types.CallbackTypeBeginBlock, func(prio uint8, contractAddr sdk.AccAddress) bool {
				captured = append(captured, tuple{p: prio, a: contractAddr})
				return false
			})
			assert.Equal(t, spec.expRemaining, captured)
		})
	}
}

func TestIterateContractCallbacksByContract(t *testing.T) {
	var (
		myAddr    = sdk.AccAddress(bytes.Repeat([]byte{1}, sdk.AddrLen))
		otherAddr = sdk.AccAddress(bytes.Repeat([]byte{2}, sdk.AddrLen))
	)
	type tuple struct {
		t types.PrivilegedCallbackType
		p uint8
	}

	specs := map[string]struct {
		setup        func(sdk.Context, *Keeper)
		captureFirst bool
		expPersisted []tuple
	}{
		"no callbacks - no empty store": {
			setup: func(ctx sdk.Context, k *Keeper) {},
		},
		"no callbacks for address": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, otherAddr)
			},
		},
		"one callback": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeEndBlock, myAddr)
			},
			expPersisted: []tuple{{p: 1, t: types.CallbackTypeEndBlock}},
		},
		"multiple callbacks - ordered by type": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeEndBlock, myAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
			},
			expPersisted: []tuple{{p: 1, t: types.CallbackTypeBeginBlock}, {p: 1, t: types.CallbackTypeEndBlock}},
		},
		"multiple callbacks same type - ordered by position": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
			},
			expPersisted: []tuple{{p: 1, t: types.CallbackTypeBeginBlock}, {p: 2, t: types.CallbackTypeBeginBlock}},
		},
		"multiple callbacks - abort after first": {
			setup: func(ctx sdk.Context, k *Keeper) {
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
				k.appendToPrivilegedContractCallbacks(ctx, types.CallbackTypeBeginBlock, myAddr)
			},
			captureFirst: true,
			expPersisted: []tuple{{p: 1, t: types.CallbackTypeBeginBlock}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(NewWasmVMMock()))
			k := keepers.TWasmKeeper
			spec.setup(ctx, k)

			// when
			var captured []tuple
			k.iterateContractCallbacksByContract(ctx, myAddr, func(t types.PrivilegedCallbackType, prio uint8) bool {
				captured = append(captured, tuple{t: t, p: prio})
				return spec.captureFirst
			})

			// then
			assert.Equal(t, spec.expPersisted, captured)
		})
	}
}

func seedTestContract(t *testing.T, ctx sdk.Context, k *Keeper) (uint64, sdk.AccAddress) {
	t.Helper()
	creatorAddr := rand.Bytes(sdk.AddrLen)
	codeID, err := k.contractKeeper.Create(ctx, creatorAddr, []byte{}, "", "", nil)
	require.NoError(t, err)
	contractAddr, _, err := k.contractKeeper.Instantiate(ctx, codeID, creatorAddr, creatorAddr, nil, "", nil)
	require.NoError(t, err)
	return codeID, contractAddr
}
