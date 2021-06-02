package poe

import (
	"encoding/json"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

type deliverTxfn func(abci.RequestDeliverTx) abci.ResponseDeliverTx

// DeliverGenTxs iterates over all genesis txs, decodes each into a Tx and
// invokes the provided deliverTxfn with the decoded Tx. It returns the result
// of the staking module's ApplyAndReturnValidatorSetUpdates.
func DeliverGenTxs(
	ctx sdk.Context, genTxs []json.RawMessage,
	stakingKeeper types.StakingKeeper, deliverTx deliverTxfn,
	txEncodingConfig client.TxEncodingConfig,
) ([]abci.ValidatorUpdate, error) {

	for _, genTx := range genTxs {
		tx, err := txEncodingConfig.TxJSONDecoder()(genTx)
		if err != nil {
			panic(err)
		}

		bz, err := txEncodingConfig.TxEncoder()(tx)
		if err != nil {
			panic(err)
		}

		res := deliverTx(abci.RequestDeliverTx{Tx: bz})
		if !res.IsOK() {
			panic(res.Log)
		}
	}

	return stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
}
