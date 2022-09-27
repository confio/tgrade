package poe

import (
	"encoding/json"
	"math"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
)

func TestInitGenesis(t *testing.T) {
	// scenario:
	// 			setup some genTX with random staking value
	// 			add the operators to the engagement group
	//			when init genesis is executed
	// 			then validators should be found in valset diff
	//			and contracts state as expected
	ctx, example := keeper.CreateDefaultTestInput(t)
	ctx = ctx.WithBlockHeight(0)
	deliverTXFn := unAuthorizedDeliverTXFn(t, ctx, example.PoEKeeper, example.TWasmKeeper.GetContractKeeper(), example.EncodingConfig.TxConfig.TxDecoder())
	app := NewAppModule(example.PoEKeeper, example.TWasmKeeper, example.BankKeeper, example.AccountKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	const numValidators = 15
	mutator, myValidators := withRandomValidators(t, ctx, example, numValidators)
	gs := types.GenesisStateFixture(mutator, func(m *types.GenesisState) {
		m.GetSeedContracts().BootstrapAccountAddress = wasmkeeper.DeterministicAccountAddress(t, 255).String()
	})
	bootstrapAccountAddr, _ := sdk.AccAddressFromBech32(gs.GetSeedContracts().BootstrapAccountAddress)
	example.Faucet.Fund(ctx, bootstrapAccountAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(100_000_000_000)))

	fundMembers := func(members []string, coins sdk.Int) {
		for _, member := range members {
			addr, err := sdk.AccAddressFromBech32(member)
			require.NoError(t, err)
			example.Faucet.Fund(ctx, addr, sdk.NewCoin(types.DefaultBondDenom, coins))
		}
	}

	fundMembers(gs.GetSeedContracts().OversightCommunityMembers, sdk.NewInt(1_000_000))
	fundMembers(gs.GetSeedContracts().ArbiterPoolMembers, sdk.NewInt(1_000_000))

	// when
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(gs)
	gotValset := app.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	// then valset diff matches
	assert.Equal(t, len(myValidators.expValidatorSet()), len(gotValset))
	gotValsetMap := valsetAsMap(gotValset)
	for _, v := range myValidators.expValidatorSet() {
		assert.InEpsilon(t, v.Power, gotValsetMap[v.PubKey.String()], 1) // compare in epsilon (differences due to floating vs. fixed point math)
	}

	// and engagement group is setup as expected
	addr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeEngagement)
	require.NoError(t, err)
	gotMembers := queryAllMembers(t, ctx, example.TWasmKeeper, addr)
	assert.Equal(t, myValidators.expEngagementGroup(), gotMembers)

	// and staking group setup as expected
	addr, err = example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	require.NoError(t, err)

	gotMembers = queryAllMembers(t, ctx, example.TWasmKeeper, addr)
	assert.Equal(t, myValidators.expStakingGroup(), gotMembers)

	// and valset config
	gotValsetConfig, err := example.PoEKeeper.ValsetContract(ctx).QueryConfig(ctx)
	require.NoError(t, err)

	mixerAddr, err := example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeMixer)
	require.NoError(t, err)

	communityPoolAddr := wasmkeeper.BuildContractAddressClassic(5, 5)
	engagementAddr := wasmkeeper.BuildContractAddressClassic(1, 1)
	expConfig := &contract.ValsetConfigResponse{
		Membership:    mixerAddr.String(),
		MinPoints:     1,
		MaxValidators: 100,
		Scaling:       1,
		FeePercentage: sdk.MustNewDecFromStr("0.50"),
		DistributionContracts: []contract.DistributionContract{
			{Address: engagementAddr.String(), Ratio: sdk.MustNewDecFromStr("0.475")},
			{Address: communityPoolAddr.String(), Ratio: sdk.MustNewDecFromStr("0.05")},
		},
		EpochReward:    sdk.NewInt64Coin("utgd", 100000),
		ValidatorGroup: wasmkeeper.BuildContractAddressClassic(1, 7).String(),
		AutoUnjail:     false,
	}
	assert.Equal(t, expConfig, gotValsetConfig)

	// and all poe contract addresses unique
	allAddr := make([]sdk.Address, 0, len(types.PoEContractType_name))
	for k := range types.PoEContractType_name {
		if types.PoEContractType(k) == types.PoEContractTypeUndefined {
			continue
		}
		addr, err = example.PoEKeeper.GetPoEContractAddress(ctx, types.PoEContractType(k))
		require.NoError(t, err)
		require.NotContains(t, allAddr, addr)
		allAddr = append(allAddr, addr)
	}
}

type validators []validator

type validator struct {
	operatorAddr sdk.AccAddress
	pubKey       cryptotypes.PubKey
	stakedAmount uint64
	engagement   uint64
}

func (v validator) geometricMean() int64 {
	// calculates sqrt(state * engagement)
	return int64(math.Trunc(math.Sqrt(float64(v.stakedAmount) * float64(v.engagement))))
}

func (v validator) sigmoidSqrt() int64 {
	// tldr: simpler but good enough version of what is impl in the contradt
	//
	// calculates sigmoid sqrt(state, engagement)
	// sigmoidSqrt returns a sigmoid-like value of the geometric mean of staked amount and
	// engagement points.
	// It is equal to `sigmoid` with `p = 0.5`, but implemented using integer sqrt instead of
	// fixed-point fractional power.
	maxRewards := 1000
	sSqrt := 0.0003

	reward := float64(maxRewards) * (2./(1.+math.Exp(-sSqrt*math.Sqrt(float64(v.stakedAmount/sdk.DefaultPowerReduction.Uint64())*float64(v.engagement)))) - 1.)

	return int64(math.Trunc(reward))
}

func (v validator) sigmoid() int64 {
	// reward = r_max * (2 / (1 + e^(-s * (stake * engagement)^p) ) - 1)
	maxRewards := 1000000
	s := 0.00001
	p := 0.62

	reward := float64(maxRewards) * (2./(1.+
		math.Exp(-s*math.Pow(float64(v.stakedAmount/sdk.DefaultPowerReduction.Uint64())*float64(v.engagement), p))) - 1.)

	return int64(math.Trunc(reward))
}

func (v validator) power() int64 {
	// FIXME: Select according to bootstrap / mixer setup params
	// return v.geometricMean()
	// return v.sigmoidSqrt()
	return v.sigmoid()
}

// validator diff sorted by power desc
func (v validators) expValidatorSet() []abci.ValidatorUpdate {
	r := make([]abci.ValidatorUpdate, 0, len(v))
	for _, x := range v {
		if power := x.power(); power > 0 {
			r = append(r, abci.ValidatorUpdate{
				PubKey: crypto.PublicKey{
					Sum: &crypto.PublicKey_Ed25519{Ed25519: x.pubKey.Bytes()},
				},
				Power: power,
			})
		}
	}
	return r
}

// engagement group members, sorted by engagement points desc
func (v validators) expEngagementGroup() []contract.TG4Member {
	r := make([]contract.TG4Member, len(v))
	for i, x := range v {
		r[i] = contract.TG4Member{
			Addr:   x.operatorAddr.String(),
			Points: x.engagement,
		}
	}
	return contract.SortByWeightDesc(r)
}

// staking group members, sorted by points desc
func (v validators) expStakingGroup() []contract.TG4Member {
	r := make([]contract.TG4Member, len(v))
	for i, x := range v {
		r[i] = contract.TG4Member{
			Addr:   x.operatorAddr.String(),
			Points: x.stakedAmount / sdk.DefaultPowerReduction.Uint64(),
		}
	}
	return contract.SortByWeightDesc(r)
}

// return genesis mutator that adds the given mumber of validators to the genesis
func withRandomValidators(t *testing.T, ctx sdk.Context, example keeper.TestKeepers, numValidators int) (func(m *types.GenesisState), validators) {
	collectValidators := make(validators, numValidators)
	return func(m *types.GenesisState) {
		f := fuzz.New()
		m.GetSeedContracts().GenTxs = make([]json.RawMessage, numValidators)
		m.GetSeedContracts().Engagement = make([]types.TG4Member, numValidators)
		for i := 0; i < numValidators; i++ {
			var ( // power * engagement must be less than 10^18 (constraint is in the contract)
				power      uint16
				engagement uint16
			)
			f.NilChance(0).Fuzz(&power) // must be > 0 so that staked amount is > 0
			f.Fuzz(&engagement)

			genTx, opAddr, pubKey := types.RandomGenTX(t, uint32(power))
			stakedAmount := sdk.TokensFromConsensusPower(int64(power), sdk.DefaultPowerReduction).Uint64()
			collectValidators[i] = validator{
				operatorAddr: opAddr,
				pubKey:       pubKey,
				stakedAmount: stakedAmount,
				engagement:   uint64(engagement),
			}
			m.GetSeedContracts().GenTxs[i] = genTx
			m.GetSeedContracts().Engagement[i] = types.TG4Member{Address: opAddr.String(), Points: uint64(engagement)}
			example.AccountKeeper.NewAccountWithAddress(ctx, opAddr)
			example.Faucet.Fund(ctx, opAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.NewIntFromUint64(stakedAmount)))
		}
	}, collectValidators
}

func queryAllMembers(t *testing.T, ctx sdk.Context, k *twasmkeeper.Keeper, addr sdk.AccAddress) []contract.TG4Member {
	pagination := contract.Paginator{
		Limit: 30,
	}
	members, err := contract.QueryTG4MembersByWeight(ctx, k, addr, &pagination)
	require.NoError(t, err)
	return members
}

// unAuthorizedDeliverTXFn applies the TX without ante handler checks for testing purpose
func unAuthorizedDeliverTXFn(t *testing.T, ctx sdk.Context, k *keeper.Keeper, contractKeeper wasmtypes.ContractOpsKeeper, txDecoder sdk.TxDecoder) func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
	t.Helper()
	h := NewHandler(k, contractKeeper, nil)
	return func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
		genTx, err := txDecoder(tx.GetTx())
		require.NoError(t, err)
		msgs := genTx.GetMsgs()
		require.Len(t, msgs, 1)
		msg := msgs[0].(*types.MsgCreateValidator)
		_, err = h(ctx, msg)
		require.NoError(t, err)
		t.Logf("+++ create validator: %s\n", msg.OperatorAddress)
		return abci.ResponseDeliverTx{}
	}
}

func valsetAsMap(s []abci.ValidatorUpdate) map[string]int64 {
	r := make(map[string]int64, len(s))
	for _, v := range s {
		r[v.PubKey.String()] = v.Power
	}
	return r
}
