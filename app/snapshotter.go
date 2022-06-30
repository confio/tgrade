package app

import (
	"fmt"

	snapshot "github.com/cosmos/cosmos-sdk/snapshots/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protoio "github.com/gogo/protobuf/io"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
)

var _ snapshot.ExtensionSnapshotter = &TgradeSnapshotter{}

type TgradeSnapshotter struct {
	*keeper.WasmSnapshotter
	app *TgradeApp
}

func NewTgradeSnapshotter(app *TgradeApp, cms sdk.MultiStore, wasm *keeper.Keeper) *TgradeSnapshotter {
	return &TgradeSnapshotter{
		WasmSnapshotter: keeper.NewWasmSnapshotter(cms, wasm),
		app:             app,
	}
}

func (t *TgradeSnapshotter) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshot.SnapshotItem, error) {
	item, err := t.WasmSnapshotter.Restore(height, format, protoReader)
	if err != nil {
		ctx := t.app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := t.app.twasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
		t.app.poeKeeper.InitContractAddressCache(ctx)
	}
	return item, err
}
