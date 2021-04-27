package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cosmwasm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

func TestInitGenesis(t *testing.T) {
	noopMock := NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
		m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
		m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{}, 0, nil
		}
	})

	type registeredCallback struct {
		addr sdk.AccAddress
		pos  uint8
		cbt  types.PrivilegedCallbackType
	}

	specs := map[string]struct {
		state          types.GenesisState
		wasmvm         *wasmtesting.MockWasmer
		expCallbackReg []registeredCallback
		expErr         bool
	}{
		"privileged contract": {
			state:  types.GenesisStateFixture(t),
			wasmvm: noopMock,
		},
		"without privileged contracts": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
			}),
			wasmvm: noopMock,
		},
		"unknown privileged contract address": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{RandomAddress(t).String()}
			}),
			wasmvm: noopMock,
			expErr: true,
		},
		"callback set for dumped contract": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(2, 2).String()}
				state.Wasm.Contracts[1] = wasmtypes.ContractFixture(func(contract *wasmtypes.Contract) {
					contract.ContractAddress = genContractAddress(2, 2).String()
					err := contract.ContractInfo.SetExtension(&types.TgradeContractDetails{
						RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "begin_block"}},
					})
					require.NoError(t, err)
					contract.ContractInfo.CodeID = 2
				})
			}),
			wasmvm:         noopMock,
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.CallbackTypeBeginBlock, addr: genContractAddress(2, 2)}},
		},
		"callback set for gen msg contract": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(2, 1).String()}
				state.Wasm.Contracts = nil
				state.Wasm.Sequences = []wasmtypes.Sequence{{IDKey: wasmtypes.KeyLastCodeID, Value: 3}}
				state.Wasm.GenMsgs = []wasmtypes.GenesisState_GenMsgs{
					{Sum: &wasmtypes.GenesisState_GenMsgs_InstantiateContract{
						InstantiateContract: wasmtypes.MsgInstantiateContractFixture(
							func(msg *wasmtypes.MsgInstantiateContract) {
								msg.CodeID = 2
								msg.Funds = nil
							}),
					}},
				}
			}),
			wasmvm: NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
				// callback registers for end block on sudo call
				m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
				m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64) (*wasmvmtypes.Response, uint64, error) {
					tradeMsg := contract.TgradeMsg{Hooks: &contract.Hooks{RegisterEndBlock: &struct{}{}}}
					msgBz, err := json.Marshal(&tradeMsg)
					require.NoError(t, err)
					return &wasmvmtypes.Response{
						Messages: []wasmvmtypes.CosmosMsg{{Custom: msgBz}},
					}, 0, nil
				}
			}),
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.CallbackTypeEndBlock, addr: genContractAddress(2, 1)}},
		},
		"callbacks set from dump but not privileged anymore": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				err := state.Wasm.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredCallbacks: []*types.RegisteredCallback{{Position: 1, CallbackType: "begin_block"}},
				})
				require.NoError(t, err)
			}),
			wasmvm: noopMock,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(spec.wasmvm))
			k := keepers.TWasmKeeper

			// when
			msgHandler := wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(k))
			valset, gotErr := InitGenesis(ctx, k, spec.state, keepers.StakingKeeper, msgHandler)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Nil(t, valset)
			// and verify flags set
			for _, v := range spec.state.PrivilegedContractAddresses {
				addr, _ := sdk.AccAddressFromBech32(v)
				assert.True(t, k.IsPrivileged(ctx, addr))
				codeInfo := k.GetContractInfo(ctx, addr)
				assert.True(t, k.IsPinnedCode(ctx, codeInfo.CodeID))
			}
			var allCallbacks int
			for _, n := range types.AllCallbackTypeNames() {
				cb := *types.PrivilegedCallbackTypeFrom(n)
				k.IterateContractCallbacksByType(ctx, cb, func(prio uint8, contractAddr sdk.AccAddress) bool {
					allCallbacks++
					return false
				})
			}
			require.Equal(t, len(spec.expCallbackReg), allCallbacks)
			for _, x := range spec.expCallbackReg {
				gotAddr := k.getPrivilegedContractCallback(ctx, x.cbt, x.pos)
				assert.Equal(t, x.addr, gotAddr)
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

// genContractAddress generates a contract address as wasmd keeper does
func genContractAddress(codeID, instanceID uint64) sdk.AccAddress {
	contractID := codeID<<32 + instanceID
	addr := make([]byte, 20)
	addr[0] = 'C'
	binary.PutUvarint(addr[1:], contractID)
	return sdk.AccAddress(crypto.AddressHash(addr))
}
