package simulation

// DONTCOVER

import (
	"io/ioutil"
	"math/rand"

	wasmsimulation "github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/confio/tgrade/app/params"
	poetypes "github.com/confio/tgrade/x/poe/types"
)

// Simulation operation weights constants
//nolint:gosec
const (
	OpWeightMsgStoreCode           = "op_weight_msg_store_code"
	OpWeightMsgInstantiateContract = "op_weight_msg_instantiate_contract"
	OpWeightMsgExecuteContract     = "op_weight_msg_execute_contract"
	OpReflectContractPath          = "op_reflect_contract_path"
)

// WasmKeeper is a subset of the wasm keeper used by simulations
type WasmKeeper interface {
	wasmsimulation.WasmKeeper
	IterateContractInfo(ctx sdk.Context, cb func(sdk.AccAddress, wasmtypes.ContractInfo) bool)
	QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	PeekAutoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64
}
type BankKeeper interface {
	simulation.BankKeeper
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simstate *module.SimulationState,
	ak wasmtypes.AccountKeeper,
	bk BankKeeper,
	wasmKeeper WasmKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgStoreCode           int
		weightMsgInstantiateContract int
		weightMsgExecuteContract     int
		wasmContractPath             string
	)

	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgStoreCode, &weightMsgStoreCode, nil,
		func(_ *rand.Rand) {
			weightMsgStoreCode = params.DefaultWeightMsgStoreCode
		},
	)

	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgInstantiateContract, &weightMsgInstantiateContract, nil,
		func(_ *rand.Rand) {
			weightMsgInstantiateContract = params.DefaultWeightMsgInstantiateContract
		},
	)
	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpWeightMsgExecuteContract, &weightMsgInstantiateContract, nil,
		func(_ *rand.Rand) {
			weightMsgExecuteContract = params.DefaultWeightMsgExecuteContract
		},
	)
	simstate.AppParams.GetOrGenerate(simstate.Cdc, OpReflectContractPath, &wasmContractPath, nil,
		func(_ *rand.Rand) {
			// simulations are run from the `app` folder
			wasmContractPath = "../testing/contract/reflect.wasm"
		},
	)
	wasmBz, err := ioutil.ReadFile(wasmContractPath)
	if err != nil {
		panic(err)
	}

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgStoreCode,
			wasmsimulation.SimulateMsgStoreCode(ak, bk, wasmKeeper, wasmBz, 5_000_000),
		),
		simulation.NewWeightedOperation(
			weightMsgInstantiateContract,
			wasmsimulation.SimulateMsgInstantiateContract(ak, bk, wasmKeeper, TgradeSimulationCodeIDSelector),
		),
		simulation.NewWeightedOperation(
			weightMsgExecuteContract,
			wasmsimulation.SimulateMsgExecuteContract(
				ak,
				bk,
				wasmKeeper,
				TgradeSimulationExecuteContractSelector,
				wasmsimulation.DefaultSimulationExecuteSenderSelector,
				wasmsimulation.DefaultSimulationExecutePayloader,
			),
		),
	}
}

// TgradeSimulationCodeIDSelector picks the first code id with unrestricted permission
func TgradeSimulationCodeIDSelector(ctx sdk.Context, wasmKeeper wasmsimulation.WasmKeeper) uint64 {
	var codeID uint64
	wasmKeeper.IterateCodeInfos(ctx, func(u uint64, info wasmtypes.CodeInfo) bool {
		if info.InstantiateConfig.Permission != wasmtypes.AccessTypeEverybody ||
			u < uint64(len(poetypes.PoEContractType_name)) { // skip all PoE contracts
			return false
		}
		codeID = u
		return true
	})
	return codeID
}

// TgradeSimulationExecuteContractSelector picks the first non PoE contract address
func TgradeSimulationExecuteContractSelector(ctx sdk.Context, wasmKeeper wasmsimulation.WasmKeeper) sdk.AccAddress {
	var r sdk.AccAddress
	wasmKeeper.IterateContractInfo(ctx, func(address sdk.AccAddress, info wasmtypes.ContractInfo) bool {
		if info.CodeID < uint64(len(poetypes.PoEContractType_name)-1) { // skip all PoE contracts
			return false
		}
		r = address
		return true
	})
	return r
}
