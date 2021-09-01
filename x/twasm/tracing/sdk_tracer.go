package tracing

import (
	"bytes"
	"fmt"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	tagModule      = "module"
	tagSDKMsgType  = "sdk_message_type"
	tagBlockHeight = "height"

	logRawStoreIO = "raw_store_io"
	logValsetDiff = "valset_diff"
)

var _ sdk.Router = &TraceRouter{}

// TraceRouter is a decorator to the sdk router that adds call tracing functionality
type TraceRouter struct {
	other sdk.Router
}

func NewTraceRouter(other sdk.Router) sdk.Router {
	if !tracerEnabled {
		return other
	}
	return &TraceRouter{other: other}
}

func (v *TraceRouter) AddRoute(r sdk.Route) sdk.Router {
	v.other = v.other.AddRoute(r)
	return v
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
			SetTag(tagSDKMsgType, fmt.Sprintf("%T", msg))

		ms := NewTracingMultiStore(ctx.MultiStore())
		result, err := realHandler(ctx.WithContext(goCtx).WithMultiStore(ms), msg)
		if err != nil {
			span.LogFields(log.Error(err))
		} else {
			addTagsFromEvents(span, result.GetEvents())
			span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
		}
		return result, err
	}
}

// BeginBlockTracer is a decorator to the begin block callback that adds tracing functionality
func BeginBlockTracer(other sdk.BeginBlocker) sdk.BeginBlocker {
	if !tracerEnabled {
		return other
	}
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "begin_block")
		span.SetTag(tagBlockHeight, req.Header.Height)
		defer span.Finish()
		ms := NewTracingMultiStore(ctx.MultiStore())
		result := other(ctx.WithContext(goCtx).WithMultiStore(ms), req)
		span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
		return result
	}
}

// EndBlockTracer is a decorator to the end block callback that adds tracing functionality
func EndBlockTracer(other sdk.EndBlocker) sdk.EndBlocker {
	if !tracerEnabled {
		return other
	}
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "end_block")
		span.SetTag(tagBlockHeight, req.Height)
		defer span.Finish()
		ms := NewTracingMultiStore(ctx.MultiStore())
		result := other(ctx.WithContext(goCtx).WithMultiStore(ms), req)
		span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
		span.LogFields(log.Object(logValsetDiff, result))
		return result
	}
}

var _ govtypes.Router = TraceGovRouter{}

// TraceGovRouter is a decorator to the sdk gov router that adds call tracing functionality
type TraceGovRouter struct {
	other govtypes.Router
}

func NewTraceGovRouter(other govtypes.Router) govtypes.Router {
	if !tracerEnabled {
		return other
	}
	return &TraceGovRouter{other: other}
}

func (t TraceGovRouter) AddRoute(r string, h govtypes.Handler) (rtr govtypes.Router) {
	return NewTraceGovRouter(t.other.AddRoute(r, h))
}

func (t TraceGovRouter) HasRoute(r string) bool {
	return t.other.HasRoute(r)
}

func (t TraceGovRouter) Seal() {
	t.other.Seal()
}

func (t TraceGovRouter) GetRoute(path string) (h govtypes.Handler) {
	realHandler := t.other.GetRoute(path)
	if realHandler == nil {
		return nil
	}
	return func(ctx sdk.Context, content govtypes.Content) error {
		if ctx.IsCheckTx() {
			return realHandler(ctx, content)
		}
		span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "gov_router")
		defer span.Finish()
		span.SetTag(tagModule, content.ProposalRoute()).
			SetTag(tagSDKMsgType, fmt.Sprintf("%T", content))

		ms := NewTracingMultiStore(ctx.MultiStore())
		err := realHandler(ctx.WithContext(goCtx).WithMultiStore(ms), content)
		if err != nil {
			span.LogFields(log.Error(err))
		} else {
			span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
		}
		return err
	}
}

var _ porttypes.IBCModule = TraceIBCHandler{}

// TraceIBCHandler is a decorator to the ibc module handler that adds call tracing functionality
type TraceIBCHandler struct {
	other porttypes.IBCModule
}

func NewTraceIBCHandler(other porttypes.IBCModule) porttypes.IBCModule {
	if !tracerEnabled {
		return other
	}
	return &TraceIBCHandler{other: other}
}

func (t TraceIBCHandler) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_open_init")
	defer span.Finish()
	return t.other.OnChanOpenInit(ctx.WithContext(goCtx), order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

func (t TraceIBCHandler) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version, counterpartyVersion string,
) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version, counterpartyVersion)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_open_try")
	defer span.Finish()
	span.SetTag(tagModule, types.ModuleName)
	return t.other.OnChanOpenTry(ctx.WithContext(goCtx), order, connectionHops, portID, channelID, channelCap, counterparty, version, counterpartyVersion)
}

func (t TraceIBCHandler) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyVersion string,
) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanOpenAck(ctx, portID, channelID, counterpartyVersion)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_open_ack")
	defer span.Finish()
	span.SetTag(tagModule, types.ModuleName)
	return t.other.OnChanOpenAck(ctx.WithContext(goCtx), portID, channelID, counterpartyVersion)
}

func (t TraceIBCHandler) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanOpenConfirm(ctx, portID, channelID)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_open_confirm")
	defer span.Finish()
	span.SetTag(tagModule, types.ModuleName)
	return t.other.OnChanOpenConfirm(ctx.WithContext(goCtx), portID, channelID)
}

func (t TraceIBCHandler) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanCloseInit(ctx, portID, channelID)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_close_init")
	defer span.Finish()
	span.SetTag(tagModule, types.ModuleName)
	return t.other.OnChanCloseInit(ctx.WithContext(goCtx), portID, channelID)
}

func (t TraceIBCHandler) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	if ctx.IsCheckTx() {
		return t.other.OnChanCloseConfirm(ctx, portID, channelID)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_close_confirm")
	defer span.Finish()
	span.SetTag(tagModule, types.ModuleName)
	return t.other.OnChanCloseConfirm(ctx.WithContext(goCtx), portID, channelID)
}

func (t TraceIBCHandler) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, []byte, error) {
	if ctx.IsCheckTx() {
		return t.other.OnRecvPacket(ctx, packet)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_recv_packet")
	defer span.Finish()
	return t.other.OnRecvPacket(ctx.WithContext(goCtx), packet)
}

func (t TraceIBCHandler) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte) (*sdk.Result, error) {
	if ctx.IsCheckTx() {
		return t.other.OnAcknowledgementPacket(ctx, packet, acknowledgement)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_ack_packet")
	defer span.Finish()
	return t.other.OnAcknowledgementPacket(ctx.WithContext(goCtx), packet, acknowledgement)
}

func (t TraceIBCHandler) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) (*sdk.Result, error) {
	if ctx.IsCheckTx() {
		return t.other.OnTimeoutPacket(ctx, packet)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ibc_chan_timeout_packet")
	defer span.Finish()
	return t.other.OnTimeoutPacket(ctx.WithContext(goCtx), packet)
}

// NewTraceAnteHandler decorates the ante handler with tracing functionality
func NewTraceAnteHandler(other sdk.AnteHandler) sdk.AnteHandler {
	if !tracerEnabled {
		return other
	}
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if simulate || ctx.IsCheckTx() {
			return other(ctx, tx, simulate)
		}
		span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "ante_handler")
		defer span.Finish()

		for _, msg := range tx.GetMsgs() {
			span.SetTag(tagSDKMsgType, fmt.Sprintf("%T", msg))
		}
		ms := NewTracingMultiStore(ctx.MultiStore())
		newCtx, err := other(ctx.WithContext(goCtx).WithMultiStore(ms), tx, simulate)
		if err != nil {
			span.LogFields(log.Error(err))
		} else {
			span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
		}
		return newCtx, err
	}
}

// tracingMultiStore Multistore that traces all operations
type tracingMultiStore struct {
	sdk.MultiStore
	buf bytes.Buffer
}

// NewTracingMultiStore constructor
func NewTracingMultiStore(store sdk.MultiStore) *tracingMultiStore {
	return &tracingMultiStore{MultiStore: store}
}

func (t *tracingMultiStore) GetStore(k sdk.StoreKey) sdk.Store {
	return tracekv.NewStore(t.MultiStore.GetKVStore(k), &t.buf, nil)
}

func (t *tracingMultiStore) GetKVStore(k sdk.StoreKey) sdk.KVStore {
	return tracekv.NewStore(t.MultiStore.GetKVStore(k), &t.buf, nil)
}
