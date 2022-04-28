package app

import (
	"encoding/json"
	"time"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/confio/tgrade/x/poe/contract"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *TgradeApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs []string,
) (servertypes.ExportedApp, error) {
	if forZeroHeight {
		panic("zero height export not supported")
	}
	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight()}).
		WithBlockTime(time.Now().UTC()) // todo (Alex): check if there is any way to get the last block time

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1
	genState := app.mm.ExportGenesis(ctx, app.appCodec)
	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := activeValidatorSet(app, ctx, err)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}
	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, err
}

func activeValidatorSet(app *TgradeApp, ctx sdk.Context, err error) ([]tmtypes.GenesisValidator, error) {
	var result []tmtypes.GenesisValidator
	valset := app.poeKeeper.ValsetContract(ctx)
	valset.IterateActiveValidators(ctx, func(c contract.ValidatorInfo) bool {
		var opAddr sdk.AccAddress
		opAddr, err = sdk.AccAddressFromBech32(c.Operator)
		if err != nil {
			return true
		}
		var pk pc.PublicKey
		pk, err = contract.ConvertToTendermintPubKey(c.ValidatorPubkey)
		if err != nil {
			return true
		}
		var tmPk tmcrypto.PubKey
		tmPk, err = encoding.PubKeyFromProto(pk)
		if err != nil {
			return true
		}
		var meta *stakingtypes.Validator
		meta, err = valset.QueryValidator(ctx, opAddr)
		if err != nil {
			return true
		}
		moniker := ""
		if meta != nil {
			moniker = meta.GetMoniker()
		}
		result = append(result, tmtypes.GenesisValidator{
			Address: tmPk.Address(),
			PubKey:  tmPk,
			Power:   int64(c.Power),
			Name:    moniker,
		})
		return false
	}, nil)
	return result, err
}
