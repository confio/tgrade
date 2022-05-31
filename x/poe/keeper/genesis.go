package keeper

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/poe/types"
)

type DeliverTxFn func(abci.RequestDeliverTx) abci.ResponseDeliverTx

// initer is subset of keeper to set initial state
type initer interface {
	SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress)
	setParams(ctx sdk.Context, params types.Params)
}

// InitGenesis - initialize accounts and deliver genesis transactions
func InitGenesis(
	ctx sdk.Context,
	keeper initer,
	deliverTx DeliverTxFn,
	genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) error {
	keeper.setParams(ctx, genesisState.Params)
	if genesisState.GetImportDump() != nil {
		for _, v := range genesisState.GetImportDump().Contracts {
			addr, err := sdk.AccAddressFromBech32(v.Address)
			if err != nil {
				return sdkerrors.Wrapf(err, "decode address: %s", v.Address)
			}
			keeper.SetPoEContractAddress(ctx, v.ContractType, addr)
		}
	} else if genesisState.GetSeedContracts() != nil {
		// seed mode
		if err := DeliverGenTxs(genesisState.GetSeedContracts().GenTxs, deliverTx, txEncodingConfig); err != nil {
			return sdkerrors.Wrap(err, "deliver gentx")
		}
	}
	return nil
}

// DeliverGenTxs iterates over all genesis txs, decodes each into a Tx and
// invokes the provided DeliverTxFn with the decoded Tx.
func DeliverGenTxs(genTxs []json.RawMessage, deliverTx DeliverTxFn, txEncodingConfig client.TxEncodingConfig) error {
	for i, genTx := range genTxs {
		tx, err := txEncodingConfig.TxJSONDecoder()(genTx)
		if err != nil {
			return sdkerrors.Wrap(err, "json decode gentx")
		}
		if err := tx.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "gentx %d", i)
		}
		bz, err := txEncodingConfig.TxEncoder()(tx)
		if err != nil {
			return sdkerrors.Wrap(err, "encode tx")
		}

		res := deliverTx(abci.RequestDeliverTx{Tx: bz})
		if !res.IsOK() {
			return sdkerrors.Wrapf(types.ErrDeliverGenTXFailed, "gentx at position %d: %s", i, res.Log)
		}
	}
	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper *Keeper) *types.GenesisState {
	genState := types.GenesisState{
		Params:    keeper.GetParams(ctx),
		SetupMode: &types.GenesisState_ImportDump{ImportDump: &types.ImportDump{}},
	}
	keeper.IteratePoEContracts(ctx, func(Ctype types.PoEContractType, addr sdk.AccAddress) bool {
		genState.GetImportDump().Contracts = append(genState.GetImportDump().Contracts, types.PoEContract{
			ContractType: Ctype,
			Address:      addr.String(),
		})
		return false
	})
	return &genState
}
