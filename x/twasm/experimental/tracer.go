package experimental

import (
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

const (
	tagContract  = "contract" // caller or acting contract
	tagCodeID    = "code_id"
	tagModule    = "module"
	tagMsgType   = "message_type"
	tagQueryType = "query_type"

	logResult = "result"
	logEvents = "events"
)

var _ sdk.Router = TraceRouter{}

type TraceRouter struct {
	other sdk.Router
}

func NewTraceRouter(other sdk.Router) *TraceRouter {
	return &TraceRouter{other: other}
}

func (v TraceRouter) AddRoute(r sdk.Route) sdk.Router {
	return TraceRouter{v.other.AddRoute(r)}
}

func (v TraceRouter) Route(ctx sdk.Context, path string) sdk.Handler {
	realHandler := v.other.Route(ctx, path)
	if realHandler == nil {
		return nil
	}
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		if ctx.IsCheckTx() {
			return realHandler(ctx, msg)
		}
		span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "router")
		defer span.Finish()
		span.SetTag(tagModule, msg.Route()).
			SetTag(tagMsgType, fmt.Sprintf("%T", msg))

		ctx = ctx.WithContext(goCtx)
		result, err := realHandler(ctx, msg)
		if err == nil {
			addTagsFromEvents(span, result.GetEvents())
		}
		return result, err
	}
}

var _ wasmkeeper.Messenger = TraceMessageHandler{}

// TraceMessageHandler is a decorator to another message handler that prints verbose log messages
type TraceMessageHandler struct {
	other wasmkeeper.Messenger
}

func NewTraceMessageHandler(other wasmkeeper.Messenger) *TraceMessageHandler {
	return &TraceMessageHandler{other: other}
}

// DispatchMsg encodes the wasmVM message and dispatches it.
func (h TraceMessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if ctx.IsCheckTx() {
		return h.other.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	}

	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "messenger")
	defer span.Finish()
	span.SetTag(tagContract, contractAddr.String())
	addTagsFromWasmContractMsg(span, msg)

	events, data, err := h.other.DispatchMsg(ctx.WithContext(goCtx), contractAddr, contractIBCPortID, msg)
	addTagsFromEvents(span, events)
	span.LogFields(
		log.Object(logResult, data),
		log.Error(err),
	)
	return events, data, err
}

func addTagsFromWasmContractMsg(span opentracing.Span, msg wasmvmtypes.CosmosMsg) {
	// todo: more detailed message type
	switch {
	case msg.Bank != nil:
		span.SetTag(tagMsgType, fmt.Sprintf("%T", msg.Bank))
	case msg.Custom != nil:
		span.SetTag(tagMsgType, "custom")
	case msg.IBC != nil:
		span.SetTag(tagMsgType, fmt.Sprintf("%T", msg.IBC))
	case msg.Staking != nil:
		span.SetTag(tagMsgType, fmt.Sprintf("%T", msg.Staking))
	case msg.Stargate != nil:
		span.SetTag(tagMsgType, fmt.Sprintf("%T", msg.Stargate))
	case msg.Wasm != nil:
		span.SetTag(tagMsgType, fmt.Sprintf("%T", msg.Wasm))
	}
}

func addTagsFromEvents(span opentracing.Span, events sdk.Events) {
	for _, e := range events {
		if e.Type == wasmtypes.CustomEventType || e.Type == sdk.EventTypeMessage {
			for _, a := range e.Attributes {
				if string(a.Key) == wasmtypes.AttributeKeyContract {
					span.SetTag(tagContract, string(a.Value))
				}
				if string(a.Key) == wasmtypes.AttributeKeyCodeID {
					span.SetTag(tagCodeID, string(a.Value))
				}
			}
		}
	}
}

var _ wasmkeeper.WASMVMQueryHandler = TraceQueryPlugin{}

type TraceQueryPlugin struct {
	other wasmkeeper.WASMVMQueryHandler
}

func NewTraceQueryPlugin(other wasmkeeper.WASMVMQueryHandler) *TraceQueryPlugin {
	return &TraceQueryPlugin{other: other}
}

func (t TraceQueryPlugin) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	if ctx.IsCheckTx() { // track only internal queries
		return t.other.HandleQuery(ctx, caller, request)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "query")
	defer span.Finish()
	span.SetTag(tagContract, caller.String())
	addTagsFromWasmQuery(span, request)
	return t.other.HandleQuery(ctx.WithContext(goCtx), caller, request)
}

func addTagsFromWasmQuery(span opentracing.Span, req wasmvmtypes.QueryRequest) {
	// todo: more detailed message type
	switch {
	case req.Bank != nil:
		span.SetTag(tagQueryType, fmt.Sprintf("%T", req.Bank))
	case req.Custom != nil:
		span.SetTag(tagQueryType, "custom")
	case req.IBC != nil:
		span.SetTag(tagQueryType, fmt.Sprintf("%T", req.IBC))
	case req.Staking != nil:
		span.SetTag(tagQueryType, fmt.Sprintf("%T", req.Staking))
	case req.Stargate != nil:
		span.SetTag(tagQueryType, fmt.Sprintf("%T", req.Stargate))
	case req.Wasm != nil:
		span.SetTag(tagQueryType, fmt.Sprintf("%T", req.Wasm))
	}
}
