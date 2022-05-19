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

type BankKeeper interface {
	simulation.BankKeeper
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simstate *module.SimulationState,
	ak wasmtypes.AccountKeeper,
	bk BankKeeper,
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
			return simtypes.NoOpMsg(twasmtypes.ModuleName, wasmtypes.MsgStoreCode{}.Type(), "no chain permission"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)

		permission := wasmKeeper.GetParams(ctx).InstantiateDefaultPermission
		config := permission.With(simAccount.Address)

		msg := wasmtypes.MsgStoreCode{
			Sender:                simAccount.Address.String(),
			WASMByteCode:          wasmBz,
			InstantiatePermission: &config,
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

// CodeIDSelector returns code id to be used in simulations
type CodeIDSelector = func(ctx sdk.Context, wasmKeeper wasmsimulation.WasmKeeper) uint64

// SimulateMsgInstantiateContract generates a MsgInstantiateContract with random values
func SimulateMsgInstantiateContract(ak wasmtypes.AccountKeeper, bk BankKeeper, wasmKeeper wasmsimulation.WasmKeeper, codeSelector CodeIDSelector) simtypes.Operation {
	return func(
		r *rand.Rand,
		app *baseapp.BaseApp,
		ctx sdk.Context,
		accs []simtypes.Account,
		chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		codeID := codeSelector(ctx, wasmKeeper)
		if codeID == 0 {
			return simtypes.NoOpMsg(twasmtypes.ModuleName, wasmtypes.MsgInstantiateContract{}.Type(), "no codes with permission available"), nil, nil
		}
		deposit := sdk.Coins{}
		spendableCoins := bk.SpendableCoins(ctx, simAccount.Address)
		for _, v := range spendableCoins {
			if bk.IsSendEnabledCoin(ctx, v) {
				deposit = deposit.Add(simtypes.RandSubsetCoins(r, sdk.NewCoins(v))...)
			}
		}

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
			CoinsSpentInMsg: deposit,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func TgradeSimulationCodeIdSelector(ctx sdk.Context, wasmKeeper wasmsimulation.WasmKeeper) uint64 {
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
