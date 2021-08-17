package contract_test

import (
	"encoding/json"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/rand"
	"sort"
	"testing"
)

func TestListValidators(t *testing.T) {
	// setup contracts and seed some data
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, expValidators := withRandomValidators(t, ctx, example, 3)
	gs := types.GenesisStateFixture(mutator)
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	// when
	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	require.NoError(t, err)
	vals, err := contract.ListValidators(ctx, example.TWasmKeeper, contractAddr)

	// then
	require.NoError(t, err)
	sort.Slice(expValidators, func(i, j int) bool {
		return expValidators[i].OperatorAddress < expValidators[j].OperatorAddress
	})
	gotValidators := make([]stakingtypes.Validator, len(vals))
	for i, v := range vals {
		gotValidators[i], err = v.ToValidator()
		require.NoError(t, err)
	}
	assert.Equal(t, expValidators, gotValidators)
}

func TestGetValidator(t *testing.T) {
	// setup contracts and seed some data
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, expValidators := withRandomValidators(t, ctx, example, 2)
	gs := types.GenesisStateFixture(mutator)
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	specs := map[string]struct {
		operatorAddr string
		expVal       stakingtypes.Validator
		expEmpty     bool
	}{
		"query one validator": {
			operatorAddr: expValidators[0].OperatorAddress,
			expVal:       expValidators[0],
		},
		"query other validator": {
			operatorAddr: expValidators[1].OperatorAddress,
			expVal:       expValidators[1],
		},
		"query with unknown address": {
			operatorAddr: sdk.AccAddress(rand.Bytes(sdk.AddrLen)).String(),
			expEmpty:     true,
		},
		"query with invalid address": {
			operatorAddr: "not an address",
			expEmpty:     true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
			require.NoError(t, err)
			opAddr, _ := sdk.AccAddressFromBech32(spec.operatorAddr)

			// when
			val, err := contract.QueryValidator(ctx, example.TWasmKeeper, contractAddr, opAddr)

			// then
			if spec.expEmpty {
				assert.Nil(t, val)
				return
			}
			gotVal, err := val.ToValidator()
			require.NoError(t, err)
			assert.Equal(t, spec.expVal, gotVal)
		})
	}

}
func TestQueryUnbondingPeriod(t *testing.T) {
	// setup contracts and seed some data
	ctx, example := keeper.CreateDefaultTestInput(t)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	module := poe.NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	mutator, _ := withRandomValidators(t, ctx, example, 1)
	gs := types.GenesisStateFixture(mutator)
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	module.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	contractAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)
	// when
	res, err := contract.QueryStakingUnbondingPeriod(ctx, example.TWasmKeeper, contractAddr)

	// then

	configuredTime := 21 * 24 * 60 * 60 // in bootstrap
	assert.Equal(t, contract.Duration{Time: configuredTime}, res)
}

// unAuthorizedDeliverTXFn applies the TX without ante handler checks for testing purpose
func unAuthorizedDeliverTXFn(t *testing.T, ctx sdk.Context, k keeper.Keeper, contractKeeper wasmtypes.ContractOpsKeeper, txDecoder sdk.TxDecoder) func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
	t.Helper()
	h := poe.NewHandler(k, contractKeeper, nil)
	return func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
		genTx, err := txDecoder(tx.GetTx())
		require.NoError(t, err)
		msgs := genTx.GetMsgs()
		require.Len(t, msgs, 1)
		msg := msgs[0].(*types.MsgCreateValidator)
		_, err = h(ctx, msg)
		require.NoError(t, err)
		t.Logf("+++ create validator: %s\n", msg.DelegatorAddress)
		return abci.ResponseDeliverTx{}
	}
}

// return genesis mutator that adds the given mumber of validators to the genesis
func withRandomValidators(t *testing.T, ctx sdk.Context, example keeper.TestKeepers, numValidators int) (func(m *types.GenesisState), []stakingtypes.Validator) {
	collectValidators := make([]stakingtypes.Validator, numValidators)
	return func(m *types.GenesisState) {
		f := fuzz.New()
		m.GenTxs = make([]json.RawMessage, numValidators)
		m.Engagement = make([]types.TG4Member, numValidators)
		for i := 0; i < numValidators; i++ {
			var ( // power * engagement must be less than 10^18 (constraint is in the contract)
				power      uint16
				engagement uint16
				desc       stakingtypes.Description
			)
			f.NilChance(0).Fuzz(&power) // must be > 0 so that staked amount is > 0
			f.Fuzz(&engagement)
			for len(desc.Moniker) < 3 { // ensure min length is met
				f.Fuzz(&desc)
			}

			genTx, opAddr, pubKey := types.RandomGenTX(t, uint32(power), func(m *types.MsgCreateValidator) {
				m.Description = desc
			})
			any, err := codectypes.NewAnyWithValue(pubKey)
			require.NoError(t, err)
			stakedAmount := sdk.TokensFromConsensusPower(int64(power)).Uint64()
			collectValidators[i] = stakingtypes.Validator{
				OperatorAddress: opAddr.String(),
				ConsensusPubkey: any,
				Description:     desc,
			}
			m.GenTxs[i] = genTx
			m.Engagement[i] = types.TG4Member{Address: opAddr.String(), Weight: uint64(engagement)}
			example.AccountKeeper.NewAccountWithAddress(ctx, opAddr)
			example.BankKeeper.SetBalances(ctx, opAddr, sdk.NewCoins(
				sdk.NewCoin(types.DefaultBondDenom, sdk.NewIntFromUint64(stakedAmount)),
			))
		}
	}, collectValidators
}
