package keeper

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	cosmwasm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitGenesis(t *testing.T) {
	mock := NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
		m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
		m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{}, 0, nil
		}
	})

	specs := map[string]struct {
		state  types.GenesisState
		expErr bool
	}{
		"import with privileged contract": {
			state: types.GenesisStateFixture(t),
		},
		"import without privileged contract": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
			}),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(mock))
			k := keepers.TWasmKeeper
			b, _ := json.Marshal(spec.state)
			t.Logf("%s", string(b))
			msgHandler := wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(k))
			valset, gotErr := InitGenesis(ctx, k, spec.state, keepers.StakingKeeper, msgHandler)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Nil(t, valset)
			for _, v := range spec.state.PrivilegedContractAddresses {
				addr, _ := sdk.AccAddressFromBech32(v)
				require.True(t, k.IsPrivileged(ctx, addr))
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	wasmCodes := make(map[string]cosmwasm.WasmCode)
	mock := NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
		m.CreateFn = func(code cosmwasm.WasmCode) (cosmwasm.Checksum, error) {
			hash := sha256.Sum256(code)
			wasmCodes[string(hash[:])] = code
			return hash[:], nil
		}
		m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
		m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{}, 0, nil
		}
		m.GetCodeFn = func(checksum cosmwasm.Checksum) (cosmwasm.WasmCode, error) {
			r, ok := wasmCodes[string(checksum)]
			require.True(t, ok)
			return r, nil
		}
	})

	specs := map[string]struct {
		state  types.GenesisState
		expErr bool
	}{
		"export with privileged contract": {
			state: types.GenesisStateFixture(t),
		},
		"export without privileged contracts": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
			}),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(mock))
			k := keepers.TWasmKeeper

			msgHandler := wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(k))
			_, err := InitGenesis(ctx, k, spec.state, keepers.StakingKeeper, msgHandler)
			require.NoError(t, err)

			// when & then
			newState := ExportGenesis(ctx, k)
			assert.Equal(t, spec.state, *newState)
		})
	}
}
