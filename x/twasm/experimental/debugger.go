package experimental

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VerboseRouter struct {
	other sdk.Router
}

func NewVerboseRouter(other sdk.Router) *VerboseRouter {
	return &VerboseRouter{other: other}
}

func (v VerboseRouter) AddRoute(r sdk.Route) sdk.Router {
	return VerboseRouter{v.other.AddRoute(r)}
}

func (v VerboseRouter) Route(ctx sdk.Context, path string) sdk.Handler {
	realHandler := v.other.Route(ctx, path)
	if realHandler == nil {
		return nil
	}
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		result, err := realHandler(ctx, msg)
		if err != nil {
			ctx.Logger().Info("Route message", "msg", msg, "path", path, "events", "err", err)
		} else {
			ctx.Logger().Info("Route message", "msg", msg, "path", path, "events", result.Events, "data", result.Data)
		}
		return result, err
	}
}

var _ wasmkeeper.Messenger = VerboseMessageHandler{}

// VerboseMessageHandler is a decorator to another message handler that prints verbose log messages
type VerboseMessageHandler struct {
	other wasmkeeper.Messenger
}

func NewVerboseMessageHandler(other wasmkeeper.Messenger) *VerboseMessageHandler {
	return &VerboseMessageHandler{other: other}
}

// DispatchMsg encodes the wasmVM message and dispatches it.
func (h VerboseMessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	events, data, err := h.other.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	ctx.Logger().Info("Dispatch message", "msg", msg, "events", events, "data", data, "err", err)
	return events, data, err
}
