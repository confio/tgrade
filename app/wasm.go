package app

import (
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	poewasm "github.com/confio/tgrade/x/poe/wasm"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
)

func SetupWasmHandlers(cdc codec.Marshaler,
	bankKeeper twasmtypes.BankKeeper,
	govRouter govtypes.Router,
	result twasmkeeper.TgradeWasmHandlerKeeper,
	poeKeeper poewasm.ViewKeeper,
) []wasmkeeper.Option {
	extMessageHandlerOpt := wasmkeeper.WithMessageHandlerDecorator(func(nested wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			// disable staking messages
			wasmkeeper.MessageHandlerFunc(func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
				if msg.Staking != nil {
					return nil, nil, sdkerrors.Wrap(wasmtypes.ErrExecuteFailed, "not supported, yet")
				}
				return nil, nil, wasmtypes.ErrUnknownMsg
			}),
			nested,
			// append our custom message handler
			twasmkeeper.NewTgradeHandler(cdc, result, bankKeeper, govRouter),
		)
	})
	extQueryHandlerOpt := wasmkeeper.WithQueryHandlerDecorator(func(nested wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
		return wasmkeeper.WasmVMQueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
			if request.Staking != nil {
				return poewasm.StakingQuerier(poeKeeper)(ctx, request.Staking)
			}
			return nested.HandleQuery(ctx, caller, request)
		})
	})
	return []wasm.Option{
		extMessageHandlerOpt,
		extQueryHandlerOpt,
	}
}
