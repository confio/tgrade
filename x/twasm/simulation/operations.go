package simulation

import (
	"io/ioutil"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	poetypes "github.com/confio/tgrade/x/poe/types"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"

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
			SimulateMsgStoreCode(ak, bk, wasmKeeper, wasmBz, 5_000_000),
		),
		simulation.NewWeightedOperation(
			weightMsgInstantiateContract,
			SimulateMsgInstantiateContract(ak, bk, wasmKeeper, TgradeSimulationCodeIdSelector),
		),
	}
}

// Alex: this should go into wasmd again
// SimulateMsgStoreCode generates a MsgStoreCode with random values
func SimulateMsgStoreCode(ak wasmtypes.AccountKeeper, bk simulation.BankKeeper, wasmKeeper wasmsimulation.WasmKeeper, wasmBz []byte, gas uint64) simtypes.Operation {
	return func(
		r *rand.Rand,
		app *baseapp.BaseApp,
		ctx sdk.Context,
		accs []simtypes.Account,
		chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if wasmKeeper.GetParams(ctx).CodeUploadAccess.Permission != wasmtypes.AccessTypeEverybody {
			return simtypes.NoOpMsg(wasmtypes.ModuleName, wasmtypes.MsgStoreCode{}.Type(), "no chain permission"), nil, nil
		}

		config := &wasmtypes.AccessConfig{
			Permission: wasmtypes.AccessTypeEverybody,
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := wasmtypes.MsgStoreCode{
			Sender:                simAccount.Address.String(),
			WASMByteCode:          wasmBz,
			InstantiatePermission: config,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           &msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    twasmtypes.ModuleName,
		}

		return GenAndDeliverTxWithRandFees(txCtx, gas)
	}
}

// Alex: this would be the extension point in wasdm
// SimulationCodeIdSelector returns code id to be used in simulations
type SimulationCodeIdSelector = func(wasmKeeper wasmsimulation.WasmKeeper, ctx sdk.Context) uint64

// Alex: add implementation to wasmd as existing selector that picks the first code id
func DefaultSimulationCodeIdSelector(wasmKeeper wasmsimulation.WasmKeeper, ctx sdk.Context) uint64 {
	var codeID uint64
	wasmKeeper.IterateCodeInfos(ctx, func(u uint64, info wasmtypes.CodeInfo) bool {
		if info.InstantiateConfig.Permission != wasmtypes.AccessTypeEverybody {
			return false
		}
		codeID = u
		return true
	})
	return codeID
}

// Alex: this should stay in wasmd
// SimulateMsgInstantiateContract generates a MsgInstantiateContract with random values
func SimulateMsgInstantiateContract(ak wasmtypes.AccountKeeper, bk simulation.BankKeeper, wasmKeeper wasmsimulation.WasmKeeper, codeSelector SimulationCodeIdSelector) simtypes.Operation {
	return func(
		r *rand.Rand,
		app *baseapp.BaseApp,
		ctx sdk.Context,
		accs []simtypes.Account,
		chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		codeID := codeSelector(wasmKeeper, ctx)
		if codeID == 0 {
			return simtypes.NoOpMsg(wasmtypes.ModuleName, wasmtypes.MsgInstantiateContract{}.Type(), "no codes with permission available"), nil, nil
		}

		deposit := simtypes.RandSubsetCoins(r, bk.SpendableCoins(ctx, simAccount.Address))
		msg := wasmtypes.MsgInstantiateContract{
			Sender: simAccount.Address.String(),
			Admin:  simtypes.RandomAccounts(r, 1)[0].Address.String(),
			CodeID: codeID,
			Label:  simtypes.RandStringOfLength(r, 10),
			Msg:    []byte(`{}`),
			Funds:  deposit,
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             &msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      twasmtypes.ModuleName,
			CoinsSpentInMsg: deposit, // Alex: this is important or the account will run out of funds
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// Alex: this one we need in this file
func TgradeSimulationCodeIdSelector(wasmKeeper wasmsimulation.WasmKeeper, ctx sdk.Context) uint64 {
	var codeID uint64
	wasmKeeper.IterateCodeInfos(ctx, func(u uint64, info wasmtypes.CodeInfo) bool {
		if info.InstantiateConfig.Permission != wasmtypes.AccessTypeEverybody ||
			u < uint64(len(poetypes.PoEContractType_name)-1) { // skip all PoE contracts
			return false
		}
		codeID = u
		return true
	})
	return codeID
}
