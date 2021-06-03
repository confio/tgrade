package keeper

import (
	"encoding/json"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

type DeliverTxFn func(abci.RequestDeliverTx) abci.ResponseDeliverTx

// InitGenesis - initialize accounts and deliver genesis transactions
func InitGenesis(
	ctx sdk.Context,
	keeper Keeper,
	deliverTx DeliverTxFn,
	genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) error {
	for _, v := range genesisState.Contracts {
		addr, _ := sdk.AccAddressFromBech32(v.Address)
		keeper.SetPoEContractAddress(ctx, v.ContractType, addr)
	}
	admin, err := sdk.AccAddressFromBech32(genesisState.SystemAdminAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	keeper.setPoESystemAdminAddress(ctx, admin)

	if len(genesisState.GenTxs) > 0 {
		if err := DeliverGenTxs(genesisState.GenTxs, deliverTx, txEncodingConfig); err != nil {
			return sdkerrors.Wrap(err, "deliver gentx")
		}
	}
	return nil
}

// DeliverGenTxs iterates over all genesis txs, decodes each into a Tx and
// invokes the provided DeliverTxFn with the decoded Tx.
func DeliverGenTxs(genTxs []json.RawMessage, deliverTx DeliverTxFn, txEncodingConfig client.TxEncodingConfig) error {
	for _, genTx := range genTxs {
		tx, err := txEncodingConfig.TxJSONDecoder()(genTx)
		if err != nil {
			return sdkerrors.Wrap(err, "json decode gentx")
		}

		bz, err := txEncodingConfig.TxEncoder()(tx)
		if err != nil {
			return sdkerrors.Wrap(err, "encode tx")
		}

		res := deliverTx(abci.RequestDeliverTx{Tx: bz})
		if !res.IsOK() {
			return sdkerrors.Wrap(types.ErrDeliverGenTXFailed, res.Log)
		}
	}
	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) *types.GenesisState {
	genState := types.GenesisState{
		SeedContracts:      false,
		SystemAdminAddress: keeper.GetPoESystemAdminAddress(ctx).String(),
		Contracts:          make([]types.PoEContract, 0),
	}
	keeper.IteratePoEContracts(ctx, func(ctype types.PoEContractTypes, addr sdk.AccAddress) bool {
		genState.Contracts = append(genState.Contracts, types.PoEContract{
			ContractType: ctype,
			Address:      addr.String(),
		})
		return false
	})
	return &genState
}
