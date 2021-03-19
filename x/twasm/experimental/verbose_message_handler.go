package experimental

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO (Alex) : make messenger interface public in wasmd
type messenger interface {
	// DispatchMsg encodes the wasmVM message and dispatches it.
	DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error)
}

// VerboseMessageHandler is a decorator to another message handler that prints verbose log messages
type VerboseMessageHandler struct {
	other messenger
}

func NewVerboseMessageHandler(other messenger) *VerboseMessageHandler {
	return &VerboseMessageHandler{other: other}
}

// DispatchMsg encodes the wasmVM message and dispatches it.
func (h VerboseMessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	events, data, err := h.other.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	ctx.Logger().Info("Handled incoming message", "source", msg, "events", events, "data", "err", err)
	return events, data, err
}
