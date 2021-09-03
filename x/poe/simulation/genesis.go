package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdkhelpers "github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"math/rand"
	"time"
)

// Simulation parameter constants
const (
	unbondingTime     = "unbonding_time"
	maxValidators     = "max_validators"
	historicalEntries = "historical_entries"
)

// GenUnbondingTime randomized UnbondingTime
func GenUnbondingTime(r *rand.Rand) (ubdTime time.Duration) {
	return time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
}

// GenMaxValidators randomized MaxValidators
func GenMaxValidators(r *rand.Rand) (maxValidators uint32) {
	return uint32(r.Intn(250) + 1)
}

// RandomizedGenState generates a random GenesisState
func RandomizedGenState(simState *module.SimulationState) {
	// prepare genTX
	var genTxs []json.RawMessage
	var engagements []types.TG4Member
	for i := 1; i < int(simState.NumBonded); i++ {
		acc := simState.Accounts[i]
		engagements = append(engagements, types.TG4Member{
			Address: acc.Address.String(),
			Weight:  10,
		})

		privkeySeed := make([]byte, 15)
		if _, err := simState.Rand.Read(privkeySeed); err != nil {
			panic(err)
		}

		createValMsg, err := types.NewMsgCreateValidator(
			acc.Address,
			ed25519.GenPrivKeyFromSecret(privkeySeed).PubKey(),
			sdk.NewCoin("stake", sdk.NewInt(simState.InitialStake)),
			stakingtypes.NewDescription("testing", "", "", "", ""),
		)
		if err != nil {
			panic(err)
		}
		txConfig := types.MakeEncodingConfig(nil).TxConfig
		txBuilder := txConfig.NewTxBuilder()
		err = txBuilder.SetMsgs(createValMsg)
		if err != nil {
			panic(err)
		}

		fmt.Printf("--> len: %d\n", len(acc.PubKey.Bytes()))
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
		var seq uint64 = 0
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

	// setup PoE genesis data
	stakingGenesis := types.GenesisState{
		Params:             types.DefaultParams(),
		SeedContracts:      true,
		GenTxs:             genTxs,
		SystemAdminAddress: simState.Accounts[0].Address.String(),
		Engagement:         engagements,
		BondDenom:          "stake",
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&stakingGenesis)

	// adjust bank module state for PoE

	// now add back the tokens to the bonded supply
	var bankGenesis banktypes.GenesisState
	rawBankState, ok := simState.GenState[banktypes.ModuleName]
	if !ok {
		panic("no bank genesis state")
	}
	simState.Cdc.MustUnmarshalJSON(rawBankState, &bankGenesis)
	bound := sdk.ZeroInt()
	for i := 0; i < int(simState.NumBonded); i++ {
		bound = bound.Add(sdk.NewInt(simState.InitialStake))
	}
	// adjust supply or bank invariants will fail. Staking module did add the amount to the module account
	bankGenesis.Supply = bankGenesis.Supply.Sub(sdk.NewCoins(sdk.NewCoin("stake", bound)))
	// always have bank transfers enabled or we fail in PoE
	bankGenesis.Params = bankGenesis.Params.SetSendEnabledParam("stake", true)
	simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGenesis)
}
