package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/app/params"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgUpdateValidator = "op_weight_msg_update_validator"
	OpWeightMsgDelegate        = "op_weight_msg_delegate"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, bk types.XBankKeeper, ak types.AccountKeeper, k keeper.Keeper) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator int
		weightMsgUpdateValidator int
		weightMsgDelegate        int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) {
			weightMsgCreateValidator = params.DefaultWeightMsgCreateValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateValidator, &weightMsgUpdateValidator, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateValidator = params.DefaultWeightMsgUpdateValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDelegate, &weightMsgDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgDelegate = params.DefaultWeightMsgDelegate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(bk, ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateValidator,
			SimulateMsgUpdateValidator(bk, ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegate,
			SimulateMsgDelegate(bk, ak, k),
		),
	}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(bk types.XBankKeeper, ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := simAccount.Address

		// ensure the validator doesn't exist already
		found, err := k.ValsetContract(ctx).QueryValidator(ctx, address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, err.Error()), nil, err
		}
		if found != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "validator exists"), nil, nil
		}

		denom := k.GetBondDenom(ctx)
		balance := bk.GetBalance(ctx, simAccount.Address, denom)
		if balance.IsNil() || !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "balance is negative"), nil, nil
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address).AmountOf(denom)
		amount, err := simtypes.RandPositiveInt(r, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		var fees sdk.Coins
		coins, hasNeg := sdk.NewCoins(sdk.NewCoin(denom, spendable)).SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate fees"), nil, err
			}
		}

		description := stakingtypes.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			"https://"+simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		msg, err := types.NewMsgCreateValidator(address, simAccount.ConsKey.PubKey(), selfDelegation, sdk.NewCoin(denom, sdk.ZeroInt()), description)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to create CreateValidator message"), nil, err
		}

		txCtx := simulation.OperationInput{
			AccountKeeper: ak,
			Bankkeeper:    bk,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgUpdateValidator generates a MsgUpdateValidator with random values
func SimulateMsgUpdateValidator(bk types.XBankKeeper, ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		validators, _, err := k.ValsetContract(ctx).ListValidators(ctx, nil)
		if len(validators) == 0 || err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateValidator, "numbers of validator equal zero"), nil, nil
		}

		val := validators[rand.Intn(len(validators))]
		address := val.GetOperator()

		simAccount, found := simtypes.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateValidator, "unable to find account"), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address)

		description := stakingtypes.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			"https://"+simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		accAddr, err := sdk.AccAddressFromBech32(address.String())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateValidator, "validator address is not valid"), nil, err
		}
		msg := types.NewMsgUpdateValidator(accAddr, description)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(bk types.XBankKeeper, ak types.AccountKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		validators, _, err := k.ValsetContract(ctx).ListValidators(ctx, nil)
		if len(validators) == 0 || err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "numbers of validator equal zero"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)

		denom := k.GetBondDenom(ctx)
		balance := bk.GetBalance(ctx, simAccount.Address, denom)
		if balance.IsNil() || !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "balance is negative"), nil, nil
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address).AmountOf(denom)
		amount, err := simtypes.RandPositiveInt(r, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		var fees sdk.Coins
		coins, hasNeg := sdk.NewCoins(sdk.NewCoin(denom, spendable)).SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate fees"), nil, err
			}
		}

		msg := types.NewMsgDelegate(simAccount.Address, selfDelegation, sdk.NewCoin(denom, sdk.ZeroInt()))

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}
