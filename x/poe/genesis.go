package poe

import (
	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// DefaultGenesisState default values
func DefaultGenesisState() types.GenesisState {
	return types.GenesisState{
		SystemAdminAddr:        sdk.AccAddress(make([]byte, sdk.AddrLen)).String(),
		EngagementContractAddr: twasm.ContractAddress(1, 1).String(),
		StakingContractAddr:    twasm.ContractAddress(2, 2).String(),
		MixerContractAddr:      twasm.ContractAddress(3, 3).String(),
		ValsetContractAddr:     twasm.ContractAddress(4, 4).String(),
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
