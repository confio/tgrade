package simulation

// DONTCOVER

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
//nolint:gosec
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgUpdateValidator = "op_weight_msg_update_validator"
	OpWeightMsgDelegate        = "op_weight_msg_delegate"
	OpWeightMsgUndelegate      = "op_weight_msg_undelegate"
)

// BankKeeper extended bank keeper used by simulations
type BankKeeper interface {
	types.BankKeeper
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// subset used in simulations
type poeKeeper interface {
	ValsetContract(ctx sdk.Context) keeper.ValsetContract
	StakeContract(ctx sdk.Context) keeper.StakeContract
	GetBondDenom(ctx sdk.Context) string
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, bk BankKeeper, ak types.AccountKeeper, k poeKeeper) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator int
		weightMsgUpdateValidator int
		weightMsgDelegate        int
		weightMsgUndelegate      int
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

	appParams.GetOrGenerate(cdc, OpWeightMsgUndelegate, &weightMsgUndelegate, nil,
		func(_ *rand.Rand) {
			weightMsgUndelegate = params.DefaultWeightMsgUndelegate
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
		simulation.NewWeightedOperation(
			weightMsgUndelegate,
			SimulateMsgUndelegate(bk, ak, k),
		),
	}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(bk BankKeeper, ak types.AccountKeeper, k poeKeeper) simtypes.Operation {
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
		if !spendable.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "spendable coins amount is negative"), nil, nil
		}

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
func SimulateMsgUpdateValidator(bk BankKeeper, ak types.AccountKeeper, k poeKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, valAddr, err := getRandValidator(ctx, k)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateValidator, "cannot fetch random validator"), nil, err
		}

		simAccount, found := simtypes.FindAccount(accs, valAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateValidator, "unable to find account"), nil, fmt.Errorf("validator %s not found", valAddr.String())
		}

		description := stakingtypes.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			"https://"+simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgUpdateValidator(valAddr, description)

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(bk BankKeeper, ak types.AccountKeeper, k poeKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, valAddr, err := getRandValidator(ctx, k)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "cannot fetch random validator"), nil, err
		}

		simAccount, found := simtypes.FindAccount(accs, valAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to find account"), nil, fmt.Errorf("validator %s not found", valAddr.String())
		}

		denom := k.GetBondDenom(ctx)
		balance := bk.GetBalance(ctx, simAccount.Address, denom)
		if balance.IsNil() || !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "balance is negative"), nil, nil
		}

		spendable := bk.SpendableCoins(ctx, simAccount.Address).AmountOf(denom)
		if !spendable.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "spendable coins amount is negative"), nil, nil
		}

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

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgUndelegate(bk BankKeeper, ak types.AccountKeeper, k poeKeeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		_, valAddr, err := getRandValidator(ctx, k)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "cannot fetch random validator"), nil, err
		}

		delegated, err := k.StakeContract(ctx).QueryStakedAmount(ctx, valAddr)
		if err != nil || delegated == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "validator does not have any delegation entries"), nil, nil
		}

		unbondAmt, err := simtypes.RandPositiveInt(r, *delegated)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "invalid unbond amount"), nil, err
		}

		if unbondAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "unbond amount is zero"), nil, nil
		}

		msg := types.NewMsgUndelegate(
			valAddr, sdk.NewCoin(k.GetBondDenom(ctx), unbondAmt),
		)

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		simAccount, found := simtypes.FindAccount(accs, valAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "unable to find account"), nil, fmt.Errorf("delegator %s not found", valAddr.String())
		}
		if simAccount.PrivKey == nil {
			// delegation address does not exist in accs
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", valAddr.String())
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func getRandValidator(ctx sdk.Context, k poeKeeper) (stakingtypes.Validator, sdk.AccAddress, error) {
	validators, _, err := k.ValsetContract(ctx).ListValidators(ctx, nil)
	if len(validators) == 0 || err != nil {
		return stakingtypes.Validator{}, nil, fmt.Errorf("cannot fetch validator list: %s", err.Error())
	}

	val := validators[rand.Intn(len(validators))]
	valAddr, err := sdk.AccAddressFromBech32(val.OperatorAddress) // plain string to not run into bech32 prefix issues with valoper
	if err != nil {
		return stakingtypes.Validator{}, nil, err
	}
	return val, valAddr, nil
}
