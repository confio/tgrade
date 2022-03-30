package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, bk types.XBankKeeper, ak types.AccountKeeper, k keeper.Keeper) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) {
			weightMsgCreateValidator = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(bk, ak, k),
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
			return simtypes.NoOpMsg(stakingtypes.ModuleName, stakingtypes.TypeMsgCreateValidator, err.Error()), nil, err
		}
		if found != nil {
			return simtypes.NoOpMsg(stakingtypes.ModuleName, stakingtypes.TypeMsgCreateValidator, "validator exists"), nil, nil
		}

		denom := k.GetBondDenom(ctx)
		balance := bk.GetBalance(ctx, simAccount.Address, denom)
		if balance.IsNil() || !balance.IsPositive() {
			return simtypes.NoOpMsg(stakingtypes.ModuleName, stakingtypes.TypeMsgCreateValidator, "balance is negative"), nil, nil
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address).AmountOf(denom)
		amount, err := simtypes.RandPositiveInt(r, spendable)
		if err != nil {
			return simtypes.NoOpMsg(stakingtypes.ModuleName, stakingtypes.TypeMsgCreateValidator, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		var fees sdk.Coins
		coins, hasNeg := sdk.NewCoins(sdk.NewCoin(denom, spendable)).SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(stakingtypes.ModuleName, stakingtypes.TypeMsgCreateValidator, "unable to generate fees"), nil, err
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
			return simtypes.NoOpMsg(stakingtypes.ModuleName, msg.Type(), "unable to create CreateValidator message"), nil, err
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
