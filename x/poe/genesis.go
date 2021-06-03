package poe

import (
	"github.com/confio/tgrade/x/poe/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// DefaultGenesisState default values
func DefaultGenesisState() types.GenesisState {
	return types.GenesisState{
		SeedContracts:      true,
		SystemAdminAddress: sdk.AccAddress(make([]byte, sdk.AddrLen)).String(),
	}
}

// InitGenesis - initialize accounts and deliver genesis transactions
func InitGenesis(
	ctx sdk.Context, stakingKeeper genutiltypes.StakingKeeper,
	deliverTx deliverTxfn, genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) (validators []abci.ValidatorUpdate, err error) {
	if len(genesisState.GenTxs) > 0 {
		validators, err = DeliverGenTxs(ctx, genesisState.GenTxs, stakingKeeper, deliverTx, txEncodingConfig)
	}
	return
}
