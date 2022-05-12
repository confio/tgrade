package simulation

import (
	"io/ioutil"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	wasmparams "github.com/CosmWasm/wasmd/app/params"
	wasmsimulation "github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simstate *module.SimulationState,
	ak wasmtypes.AccountKeeper,
	bk simulation.BankKeeper,
	wasmKeeper wasmsimulation.WasmKeeper,
) simulation.WeightedOperations {
	var (
		weightMsgStoreCode           int
		weightMsgInstantiateContract int
		wasmContractPath             string
	)

	simstate.AppParams.GetOrGenerate(simstate.Cdc, wasmsimulation.OpWeightMsgStoreCode, &weightMsgStoreCode, nil,
		func(_ *rand.Rand) {
			weightMsgStoreCode = wasmparams.DefaultWeightMsgStoreCode
		},
	)

	simstate.AppParams.GetOrGenerate(simstate.Cdc, wasmsimulation.OpWeightMsgInstantiateContract, &weightMsgInstantiateContract, nil,
		func(_ *rand.Rand) {
			weightMsgInstantiateContract = wasmparams.DefaultWeightMsgInstantiateContract
		},
	)
	simstate.AppParams.GetOrGenerate(simstate.Cdc, wasmsimulation.OpReflectContractPath, &wasmContractPath, nil,
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
			wasmsimulation.SimulateMsgStoreCode(ak, bk, wasmKeeper, wasmBz),
		),
		simulation.NewWeightedOperation(
			weightMsgInstantiateContract,
			wasmsimulation.SimulateMsgInstantiateContract(ak, bk, wasmKeeper),
		),
	}
}
