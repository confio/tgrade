package poe

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/confio/tgrade/x/twasm"

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
	app := NewAppModule(example.PoEKeeper, example.TWasmKeeper, deliverTXFn, example.EncodingConfig.TxConfig, example.TWasmKeeper.GetContractKeeper())

	const numValidators = 15
	mutator, myValidators := withRandomValidators(t, ctx, example, numValidators)
	gs := types.GenesisStateFixture(mutator)
	adminAddr, _ := sdk.AccAddressFromBech32(gs.SystemAdminAddress)
	example.Faucet.Fund(ctx, adminAddr, sdk.NewCoin(types.DefaultBondDenom, sdk.NewInt(100_000_000_000)))

	fundMembers := func(members []string, coins sdk.Int) {
		for _, member := range members {
			addr, err := sdk.AccAddressFromBech32(member)
			require.NoError(t, err)
			example.Faucet.Fund(ctx, addr, sdk.NewCoin(types.DefaultBondDenom, coins))
		}
	}

	fundMembers(gs.OversightCommunityMembers, sdk.NewInt(1_000_000))
	fundMembers(gs.ArbiterPoolMembers, sdk.NewInt(1_000_000))

	// when
	genesisBz := example.EncodingConfig.Marshaler.MustMarshalJSON(&gs)
	gotValset := app.InitGenesis(ctx, example.EncodingConfig.Marshaler, genesisBz)

	// then valset diff matches
	assert.Equal(t, valsetAsMap(myValidators.expValidatorSet()), valsetAsMap(gotValset)) // compare unordered

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

	communityPoolAddr := twasm.ContractAddress(5, 5)
	engagementAddr := twasm.ContractAddress(1, 1)
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
		ValidatorGroup: twasm.ContractAddress(1, 7).String(),
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

func (v validator) power() int64 {
	// mixer contract calculates sqrt(state*engagement)
	return int64(math.Trunc(math.Sqrt(float64(v.stakedAmount) * float64(v.engagement))))
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

// staking group members, sorted by staked amount desc
func (v validators) expStakingGroup() []contract.TG4Member {
	r := make([]contract.TG4Member, len(v))
	for i, x := range v {
		r[i] = contract.TG4Member{
			Addr:   x.operatorAddr.String(),
			Points: x.stakedAmount,
		}
	}
	return contract.SortByWeightDesc(r)
}

// return genesis mutator that adds the given mumber of validators to the genesis
func withRandomValidators(t *testing.T, ctx sdk.Context, example keeper.TestKeepers, numValidators int) (func(m *types.GenesisState), validators) {
	collectValidators := make(validators, numValidators)
	return func(m *types.GenesisState) {
		f := fuzz.New()
		m.GenTxs = make([]json.RawMessage, numValidators)
		m.Engagement = make([]types.TG4Member, numValidators)
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
			m.GenTxs[i] = genTx
			m.Engagement[i] = types.TG4Member{Address: opAddr.String(), Points: uint64(engagement)}
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
func unAuthorizedDeliverTXFn(t *testing.T, ctx sdk.Context, k keeper.Keeper, contractKeeper wasmtypes.ContractOpsKeeper, txDecoder sdk.TxDecoder) func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
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
