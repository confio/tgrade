package keeper

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cosmwasm "github.com/CosmWasm/wasmvm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
)

func TestInitGenesis(t *testing.T) {
	type vmCalls struct {
		pinCalled  bool
		sudoCalled bool
	}
	noopMock := NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
		m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
		m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{}, 0, nil
		}
	})

	type registeredCallback struct {
		addr sdk.AccAddress
		pos  uint8
		cbt  types.PrivilegeType
	}

	specs := map[string]struct {
		state          types.GenesisState
		wasmvm         *wasmtesting.MockWasmer
		expCallbackReg []registeredCallback
		expErr         bool
		expVmCalls     vmCalls
	}{
		"pin WASM code": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.Codes = append(state.Codes,
					wasmtypes.CodeFixture(func(c *wasmtypes.Code) { c.CodeID = 5 }),
					wasmtypes.CodeFixture(func(c *wasmtypes.Code) { c.CodeID = 7 }),
				)
				state.PinnedCodeIDs = []uint64{5, 7}
			}),
			wasmvm:     noopMock,
			expVmCalls: vmCalls{true, true},
		},
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
		"privilege set for dumped contract": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(2, 2).String()}
				state.Contracts[1] = types.ContractFixture(t, func(contract *types.Contract) {
					contract.ContractAddress = genContractAddress(2, 2).String()
					err := contract.ContractInfo.SetExtension(&types.TgradeContractDetails{
						RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
					})
					require.NoError(t, err)
					contract.ContractInfo.CodeID = 2
				})
			}),
			wasmvm:         noopMock,
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.PrivilegeTypeBeginBlock, addr: genContractAddress(2, 2)}},
		},
		"privilege set for gen msg contract": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(2, 1).String()}
				state.Contracts = nil
				state.Sequences = []wasmtypes.Sequence{{IDKey: wasmtypes.KeyLastCodeID, Value: 3}}
				state.GenMsgs = []wasmtypes.GenesisState_GenMsgs{
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
				m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
					tradeMsg := contract.TgradeMsg{Privilege: &contract.PrivilegeMsg{Request: types.PrivilegeTypeEndBlock}}
					msgBz, err := json.Marshal(&tradeMsg)
					require.NoError(t, err)
					return &wasmvmtypes.Response{
						Messages: []wasmvmtypes.SubMsg{{ReplyOn: wasmvmtypes.ReplyNever, Msg: wasmvmtypes.CosmosMsg{Custom: msgBz}}},
					}, 0, nil
				}
			}),
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.PrivilegeTypeEndBlock, addr: genContractAddress(2, 1)}},
		},
		"privileges set from dump": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[1].ContractAddress = genContractAddress(2, 2).String()
				err := state.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
				})
				require.NoError(t, err)
			}),
			wasmvm:         noopMock,
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.PrivilegeTypeBeginBlock, addr: genContractAddress(2, 2)}},
		},
		"invalid contract details from dump": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				var err error
				state.Contracts[1].ContractInfo.Extension, err = codectypes.NewAnyWithValue(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "non-existing-privilege"}},
				})
				require.NoError(t, err)
			}),
			wasmvm: noopMock,
			expErr: true,
		},
		"no contract details in dump": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				require.NoError(t, state.Contracts[1].ContractInfo.SetExtension(nil))
			}),
			wasmvm: noopMock,
		},
		"privileged state importer contract imports from dump": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[1].ContractAddress = genContractAddress(2, 2).String()
				state.Contracts[1].ContractState = &types.Contract_CustomModel{CustomModel: &types.CustomModel{Msg: wasmtypes.RawContractMessage(`{"my":"state"}`)}}
				err := state.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "state_exporter_importer"}},
				})
				require.NoError(t, err)
			}),
			wasmvm: NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
				m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
				m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
					return &wasmvmtypes.Response{}, 0, nil
				}
			}),
			expCallbackReg: []registeredCallback{{pos: 1, cbt: types.PrivilegeStateExporterImporter, addr: genContractAddress(2, 2)}},
		},
		"privileged state importer contract imports from dump with custom model removed": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[1].ContractAddress = genContractAddress(2, 2).String()
				state.Contracts[1].ContractState = &types.Contract_CustomModel{}
				err := state.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "state_exporter_importer"}},
				})
				require.NoError(t, err)
			}),
			wasmvm: noopMock,
			expErr: true,
		},
		"privileged state importer contract imports fails on sudo call with custom msg": {
			state: types.GenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[0].ContractAddress = genContractAddress(2, 2).String()
				state.Contracts[0].ContractState = &types.Contract_CustomModel{CustomModel: &types.CustomModel{Msg: wasmtypes.RawContractMessage(`{"my":"state"}`)}}
				err := state.Contracts[0].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "state_exporter_importer"}},
				})
				require.NoError(t, err)
				state.Contracts = []types.Contract{state.Contracts[0]}
			}),
			wasmvm: NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
				m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
				m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
					return &wasmvmtypes.Response{}, 0, errors.New("testing error")
				}
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(spec.wasmvm))
			k := keepers.TWasmKeeper

			// when
			msgHandler := wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(k))
			valset, gotErr := InitGenesis(ctx, k, spec.state, msgHandler)

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
			for _, id := range spec.state.PinnedCodeIDs {
				assert.True(t, k.IsPinnedCode(ctx, id))
			}

			var allRegisteredCallbacksCount int
			for _, n := range types.AllPrivilegeTypeNames() {
				cb := *types.PrivilegeTypeFrom(n)
				k.IteratePrivilegedContractsByType(ctx, cb, func(prio uint8, contractAddr sdk.AccAddress) bool {
					allRegisteredCallbacksCount++
					return false
				})
			}
			require.Equal(t, len(spec.expCallbackReg), allRegisteredCallbacksCount)
			for _, x := range spec.expCallbackReg {
				gotAddr := k.getPrivilegedContract(ctx, x.cbt, x.pos)
				assert.Equal(t, x.addr, gotAddr)
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	wasmCodes := make(map[string]cosmwasm.WasmCode)
	noopVMMock := NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
		m.CreateFn = func(code cosmwasm.WasmCode) (cosmwasm.Checksum, error) {
			hash := sha256.Sum256(code)
			wasmCodes[string(hash[:])] = code
			return hash[:], nil
		}
		m.PinFn = func(checksum cosmwasm.Checksum) error { return nil }
		m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{}, 0, nil
		}
		m.GetCodeFn = func(checksum cosmwasm.Checksum) (cosmwasm.WasmCode, error) {
			r, ok := wasmCodes[string(checksum)]
			require.True(t, ok)
			return r, nil
		}
	})

	specs := map[string]struct {
		srcState   types.GenesisState
		expState   types.GenesisState
		alterState func(ctx sdk.Context, keepers TestKeepers)
		expErr     bool
		mockVM     *wasmtesting.MockWasmer
	}{
		"export with privileged contract": {
			srcState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(1, 1).String()}
			}),
			alterState: func(ctx sdk.Context, keepers TestKeepers) {
				priv := types.PrivilegeTypeBeginBlock
				setContractPrivilege(t, ctx, keepers, genContractAddress(1, 1), priv)
			},
			expState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[1].ContractAddress = genContractAddress(1, 1).String()
				err := state.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "begin_blocker"}},
				})
				require.NoError(t, err)
			}),
			mockVM: noopVMMock,
		},
		"privileged state exporter contract": {
			srcState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = []string{genContractAddress(1, 1).String()}
			}),
			expState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
				state.Contracts[1].ContractState = &types.Contract_CustomModel{CustomModel: &types.CustomModel{Msg: wasmtypes.RawContractMessage(`{"my":"state"}`)}}
				err := state.Contracts[1].ContractInfo.SetExtension(&types.TgradeContractDetails{
					RegisteredPrivileges: []types.RegisteredPrivilege{{Position: 1, PrivilegeType: "state_exporter_importer"}},
				})
				require.NoError(t, err)
			}),
			alterState: func(ctx sdk.Context, keepers TestKeepers) {
				priv := types.PrivilegeStateExporterImporter
				setContractPrivilege(t, ctx, keepers, genContractAddress(1, 1), priv)
			},
			mockVM: NewWasmVMMock(func(m *wasmtesting.MockWasmer) {
				m.CreateFn = noopVMMock.CreateFn
				m.PinFn = noopVMMock.PinFn
				m.GetCodeFn = noopVMMock.GetCodeFn
				m.SudoFn = func(codeID cosmwasm.Checksum, env wasmvmtypes.Env, sudoMsg []byte, store cosmwasm.KVStore, goapi cosmwasm.GoAPI, querier cosmwasm.Querier, gasMeter cosmwasm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
					// return tgrade message with exported state
					return &wasmvmtypes.Response{
						Data: []byte(`{"my":"state"}`),
					}, 0, nil
				}
			}),
		},
		"export without privileged contracts": {
			srcState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
			}),
			expState: types.DeterministicGenesisStateFixture(t, func(state *types.GenesisState) {
				state.PrivilegedContractAddresses = nil
			}),
			mockVM: noopVMMock,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, keepers := CreateDefaultTestInput(t, wasmkeeper.WithWasmEngine(spec.mockVM))
			k := keepers.TWasmKeeper

			msgHandler := wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(k))
			_, err := InitGenesis(ctx, k, spec.srcState, msgHandler)
			require.NoError(t, err)

			if spec.alterState != nil {
				spec.alterState(ctx, keepers)
			}
			// when & then
			newState := ExportGenesis(ctx, k)
			assert.Equal(t, spec.expState, *newState)
		})
	}
}

func setContractPrivilege(t *testing.T, ctx sdk.Context, keepers TestKeepers, contractAddr sdk.AccAddress, priv types.PrivilegeType) {
	t.Helper()
	var details types.TgradeContractDetails
	contractInfo := keepers.TWasmKeeper.GetContractInfo(ctx, contractAddr)
	require.NoError(t, contractInfo.ReadExtension(&details))
	pos, err := keepers.TWasmKeeper.appendToPrivilegedContracts(ctx, priv, contractAddr)
	require.NoError(t, err)
	details.AddRegisteredPrivilege(priv, pos)
	require.NoError(t, keepers.TWasmKeeper.setContractDetails(ctx, contractAddr, &details))
}

// genContractAddress generates a contract address as wasmd keeper does
func genContractAddress(codeID, instanceID uint64) sdk.AccAddress {
	return wasmkeeper.BuildContractAddress(codeID, instanceID)
}
