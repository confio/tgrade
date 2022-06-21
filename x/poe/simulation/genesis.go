package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/math"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdkhelpers "github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/confio/tgrade/x/poe/types"
)

const defaultDenom = "utgd"

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {
	// tgrade-contracts do not support "stake" token
	// quick and dirty hack to replace the default staking denom everywhere

	for k, v := range simState.GenState {
		if len(v) == 0 {
			continue
		}
		simState.GenState[k] = []byte(strings.ReplaceAll(string(v), "\"stake\"", fmt.Sprintf("%q", defaultDenom)))
	}

	// ensure bank module state for PoE
	var bankGenesis banktypes.GenesisState
	rawBankState, ok := simState.GenState[banktypes.ModuleName]
	if !ok {
		panic("no bank genesis state")
	}
	simState.Cdc.MustUnmarshalJSON(rawBankState, &bankGenesis)
	totalBound := sdk.ZeroInt()
	for i := 0; i < int(simState.NumBonded); i++ {
		availableBalance := bankGenesis.Balances[i].GetCoins().AmountOf(defaultDenom)
		stakedAmount := sdk.MinInt(sdk.NewInt(simState.InitialStake), availableBalance)
		if stakedAmount.IsZero() || !stakedAmount.Equal(sdk.NewInt(simState.InitialStake)) {
			panic("not enough to stake on balance")
		}
		totalBound = totalBound.Add(sdk.NewInt(simState.InitialStake))
	}

	if len(simState.Accounts)-int(simState.NumBonded) == 0 {
		// fail fast as no account has enough tokens to pay deposit for OC or AP
		// panic("all accounts are bonded")
		// todo: find a proper solution to this hack
		simState.NumBonded = 1 + simState.Rand.Int63n(simState.NumBonded-1)
	}

	// add some random oversight community members
	opMembersCount := math.MinInt(simState.Rand.Int()+1, len(simState.Accounts)-int(simState.NumBonded))
	ocMembers := make([]string, opMembersCount)
	for i := 0; i < opMembersCount; i++ {
		ocMembers[i] = simState.Accounts[int(simState.NumBonded)+i].Address.String()
	}

	// add some random arbiter pool members
	apMembersCount := math.MinInt(simState.Rand.Int()+1, len(simState.Accounts)-int(simState.NumBonded))
	apMembers := make([]string, apMembersCount)
	for i := 0; i < apMembersCount; i++ {
		apMembers[i] = simState.Accounts[int(simState.NumBonded)+i].Address.String()
	}

	txConfig := types.MakeEncodingConfig(nil).TxConfig

	// prepare genTX
	genTxs := make([]json.RawMessage, 0, int(simState.NumBonded))
	engagements := make([]types.TG4Member, 0, int(simState.NumBonded))
	for i := 1; i < int(simState.NumBonded); i++ {
		acc := simState.Accounts[i]
		if acc.Address.String() != bankGenesis.Balances[i].Address {
			panic("all is broken when accounts do not match balances")
		}
		engagements = append(engagements, types.TG4Member{
			Address: acc.Address.String(),
			// this is what genesis validators get
			Points: 2000,
		})

		privkeySeed := make([]byte, 15)
		if _, err := simState.Rand.Read(privkeySeed); err != nil {
			panic(err)
		}

		createValMsg, err := types.NewMsgCreateValidator(
			acc.Address,
			ed25519.GenPrivKeyFromSecret(privkeySeed).PubKey(),
			sdk.NewCoin(defaultDenom, sdk.NewInt(simState.InitialStake)),
			sdk.NewCoin(defaultDenom, sdk.ZeroInt()),
			stakingtypes.NewDescription("testing", "", "", "", ""),
		)
		if err != nil {
			panic(err)
		}
		txBuilder := txConfig.NewTxBuilder()
		err = txBuilder.SetMsgs(createValMsg)
		if err != nil {
			panic(err)
		}

		// First round: we gather all the signer infos. We use the "set empty
		// signature" hack to do that.
		sigv2 := signing.SignatureV2{
			PubKey: acc.PubKey,
			Data: &signing.SingleSignatureData{
				SignMode:  txConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: 0,
		}
		err = txBuilder.SetSignatures(sigv2)
		if err != nil {
			panic(err)
		}

		// Second round: all signer infos are set, so each signer can sign.
		var seq uint64
		data := authsigning.SignerData{
			ChainID:       sdkhelpers.SimAppChainID,
			AccountNumber: 0, // in genesis
			Sequence:      seq,
		}
		sigv2, err = tx.SignWithPrivKey(txConfig.SignModeHandler().DefaultMode(), data, txBuilder, acc.PrivKey, txConfig, seq)
		if err != nil {
			panic(err)
		}
		if err := txBuilder.SetSignatures(sigv2); err != nil {
			panic(err)
		}
		txBz, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			panic(err)
		}
		genTxs = append(genTxs, txBz)
	}

	poeGenesis := types.DefaultGenesisState()
	// ensure they have reasonable engagement for the simulations
	poeGenesis.Params.InitialValEngagementPoints = 100
	// we use 100.000 instead of 1.000.000 to make the simulated staked tokens a higher value
	// (I couldn't figure out how to adjust how much each account had, so this effectively multiplies by 10)
	// poeGenesis.GetSeedContracts().StakeContractConfig.TokensPerPoint = 100000
	poeGenesis.GetSeedContracts().GenTxs = genTxs
	poeGenesis.GetSeedContracts().BootstrapAccountAddress = simState.Accounts[len(simState.Accounts)-1].Address.String() // use a non validator account
	poeGenesis.GetSeedContracts().Engagement = engagements
	poeGenesis.GetSeedContracts().BondDenom = defaultDenom
	poeGenesis.GetSeedContracts().OversightCommunityMembers = ocMembers
	poeGenesis.GetSeedContracts().ArbiterPoolMembers = apMembers
	poeGenesis.GetSeedContracts().ArbiterPoolContractConfig.EscrowAmount.Denom = poeGenesis.GetSeedContracts().BondDenom
	poeGenesis.GetSeedContracts().ArbiterPoolContractConfig.DisputeCost.Denom = poeGenesis.GetSeedContracts().BondDenom
	poeGenesis.GetSeedContracts().ValsetContractConfig.EpochReward.Denom = poeGenesis.GetSeedContracts().BondDenom
	poeGenesis.GetSeedContracts().ValsetContractConfig.EpochLength = time.Second
	poeGenesis.GetSeedContracts().OversightCommitteeContractConfig.EscrowAmount.Denom = poeGenesis.GetSeedContracts().BondDenom
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(poeGenesis)
	if err := types.ValidateGenesis(*poeGenesis, txConfig.TxJSONDecoder()); err != nil {
		panic(err)
	}

	// Bank module
	// adjust supply or bank invariants will fail. Staking module did add the amount to the module account
	fmt.Printf("total supply: %s bond tokens: %s", bankGenesis.Supply.String(), sdk.NewCoin(defaultDenom, totalBound).String())
	bankGenesis.Supply = bankGenesis.Supply.Sub(sdk.NewCoins(sdk.NewCoin(defaultDenom, totalBound)))

	// always have bank transfers enabled or we fail in PoE
	bankGenesis.Params = bankGenesis.Params.SetSendEnabledParam(defaultDenom, true)
	simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGenesis)
}
