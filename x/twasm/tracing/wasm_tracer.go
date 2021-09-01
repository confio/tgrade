package tracing

import (
	"encoding/json"
	"fmt"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"strings"
)

const (
	tagContract          = "contract"
	tagSenderContract    = "sender_contract"
	tagCodeID            = "code_id"
	tagWasmMsgCategory   = "wasm_message_category"
	tagWasmMsgType       = "wasm_message_type"
	tagWasmQueryCategory = "wasm_query_category"
	tagWasmQueryType     = "wasm_query_type"

	logRawWasmMsg         = "raw_wasm_message"
	logWasmMsgResult      = "raw_wasm_message_result"
	logRawWasmQuery       = "raw_wasm_query"
	logRawWasmQueryResult = "raw_wasm_query_result"
)

var _ wasmkeeper.Messenger = TraceMessageHandler{}

// TraceMessageHandler is a decorator to another message handler adds call tracing functionality
type TraceMessageHandler struct {
	other wasmkeeper.Messenger
	cdc   codec.Marshaler
}

// TraceMessageHandlerDecorator wasm keeper option to decorate the messenger for tracing
func TraceMessageHandlerDecorator(cdc codec.Marshaler) func(other wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(other wasmkeeper.Messenger) wasmkeeper.Messenger {
		if !tracerEnabled {
			return other
		}
		return &TraceMessageHandler{other: other, cdc: cdc}
	}
}

func (h TraceMessageHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if ctx.IsCheckTx() {
		return h.other.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
	}

	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "messenger")

	defer span.Finish()
	span.SetTag(tagSenderContract, contractAddr.String())
	addTagsFromWasmContractMsg(span, msg, h.cdc)
	ms := NewTracingMultiStore(ctx.MultiStore())
	events, data, err := h.other.DispatchMsg(ctx.WithContext(goCtx).WithMultiStore(ms), contractAddr, contractIBCPortID, msg)
	if err != nil {
		span.LogFields(log.Error(err))
	} else {
		addTagsFromEvents(span, events)
		span.LogFields(log.Object(logWasmMsgResult, data), log.String(logRawStoreIO, ms.buf.String()))
	}
	return events, data, err
}

var _ wasmkeeper.WasmVMQueryHandler = TraceQueryPlugin{}

// TraceQueryPlugin is a decorator to a WASMVMQueryHandler that adds tracing functionality
type TraceQueryPlugin struct {
	other wasmkeeper.WasmVMQueryHandler
}

// TraceQueryDecorator wasm keeper option to decorate the query handler for tracing
func TraceQueryDecorator(other wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
	if !tracerEnabled {
		return other
	}
	return &TraceQueryPlugin{other: other}
}

func (t TraceQueryPlugin) HandleQuery(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
	if ctx.IsCheckTx() { // track only internal queries
		return t.other.HandleQuery(ctx, caller, request)
	}
	span, goCtx := opentracing.StartSpanFromContext(ctx.Context(), "wasm-query")
	defer span.Finish()
	span.SetTag(tagSenderContract, caller.String())
	addTagsFromWasmQuery(span, request)
	ms := NewTracingMultiStore(ctx.MultiStore())
	result, err := t.other.HandleQuery(ctx.WithContext(goCtx).WithMultiStore(ms), caller, request)
	if err != nil {
		span.LogFields(log.Error(err))
	} else {
		span.LogFields(log.String(logRawWasmQueryResult, string(result)))
		span.LogFields(log.String(logRawStoreIO, ms.buf.String()))
	}
	return result, err
}

func addTagsFromWasmQuery(span opentracing.Span, req wasmvmtypes.QueryRequest) {
	bz, err := json.Marshal(&req)
	if err != nil {
		bz = []byte("failed to marshal original query")
	}
	span.LogFields(log.String(logRawWasmQuery, string(bz)))

	switch {
	case req.Bank != nil:
		span.SetTag(tagWasmQueryCategory, fmt.Sprintf("%T", req.Bank))
		switch {
		case req.Bank.Balance != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Bank.Balance))
		case req.Bank.AllBalances != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Bank.AllBalances))
		default:
			span.SetTag(tagWasmQueryType, "unknown")
		}
	case req.Custom != nil:
		span.SetTag(tagWasmQueryCategory, "custom")
	case req.IBC != nil:
		span.SetTag(tagWasmQueryCategory, fmt.Sprintf("%T", req.IBC))
		switch {
		case req.IBC.PortID != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.IBC.PortID))
		case req.IBC.ListChannels != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.IBC.ListChannels))
		case req.IBC.Channel != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.IBC.Channel))
		default:
			span.SetTag(tagWasmQueryType, "unknown")
		}
	case req.Staking != nil:
		span.SetTag(tagWasmQueryCategory, fmt.Sprintf("%T", req.Staking))
		switch {
		case req.Staking.Validator != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Staking.Validator))
		case req.Staking.AllValidators != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Staking.AllValidators))
		case req.Staking.AllDelegations != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Staking.AllDelegations))
		case req.Staking.BondedDenom != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Staking.BondedDenom))
		case req.Staking.Delegation != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Staking.Delegation))
		default:
			span.SetTag(tagWasmQueryType, "unknown")
		}
	case req.Stargate != nil:
		span.SetTag(tagWasmQueryCategory, fmt.Sprintf("%T", req.Stargate))
		span.SetTag(tagWasmQueryType, req.Stargate.Path)
	case req.Wasm != nil:
		span.SetTag(tagWasmQueryCategory, fmt.Sprintf("%T", req.Wasm))
		switch {
		case req.Wasm.Smart != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Wasm.Smart))
		case req.Wasm.Raw != nil:
			span.SetTag(tagWasmQueryType, fmt.Sprintf("%T", req.Wasm.Raw))
		case req.Staking.AllValidators != nil:
		default:
			span.SetTag(tagWasmQueryType, "unknown")
		}
	}
}

func addTagsFromWasmContractMsg(span opentracing.Span, msg wasmvmtypes.CosmosMsg, cdc codec.Marshaler) {
	bz, err := json.Marshal(&msg)
	if err != nil {
		bz = []byte("failed to marshal original message")
	}
	span.LogFields(log.String(logRawWasmMsg, string(bz)))
	switch {
	case msg.Bank != nil:
		span.SetTag(tagWasmMsgCategory, fmt.Sprintf("%T", msg.Bank))
		switch {
		case msg.Bank.Burn != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Bank.Burn))
		case msg.Bank.Send != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Bank.Send))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	case msg.Custom != nil:
		span.SetTag(tagWasmMsgCategory, "custom")
		var tMsg contract.TgradeMsg
		if err := tMsg.UnmarshalWithAny(msg.Custom, cdc); err != nil {
			span.SetTag(tagWasmMsgType, "can not unmarshal as tgrade message")
			return
		}
		switch {
		case tMsg.Privilege != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", tMsg.Privilege))
		case tMsg.ExecuteGovProposal != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", tMsg.ExecuteGovProposal))
		case tMsg.MintTokens != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", tMsg.MintTokens))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	case msg.Distribution != nil:
		span.SetTag(tagWasmMsgCategory, "distribution")
		switch {
		case msg.Distribution.SetWithdrawAddress != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Distribution.SetWithdrawAddress))
		case msg.Distribution.WithdrawDelegatorReward != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Distribution.WithdrawDelegatorReward))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	case msg.IBC != nil:
		span.SetTag(tagWasmMsgCategory, fmt.Sprintf("%T", msg.IBC))
		switch {
		case msg.IBC.SendPacket != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.IBC.SendPacket))
		case msg.IBC.Transfer != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.IBC.Transfer))
		case msg.IBC.CloseChannel != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.IBC.CloseChannel))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	case msg.Staking != nil:
		span.SetTag(tagWasmMsgCategory, fmt.Sprintf("%T", msg.Staking))
		switch {
		case msg.Staking.Delegate != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Staking.Delegate))
		case msg.Staking.Undelegate != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Staking.Undelegate))
		case msg.Staking.Redelegate != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Staking.Redelegate))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	case msg.Stargate != nil:
		span.SetTag(tagWasmMsgCategory, fmt.Sprintf("%T", msg.Stargate))
		span.SetTag(tagWasmMsgType, msg.Stargate.TypeURL)
	case msg.Wasm != nil:
		span.SetTag(tagWasmMsgCategory, fmt.Sprintf("%T", msg.Wasm))
		switch {
		case msg.Wasm.Migrate != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Wasm.Migrate))
		case msg.Wasm.Execute != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Wasm.Execute))
		case msg.Wasm.Instantiate != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Wasm.Instantiate))
		case msg.Wasm.UpdateAdmin != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Wasm.UpdateAdmin))
		case msg.Wasm.ClearAdmin != nil:
			span.SetTag(tagWasmMsgType, fmt.Sprintf("%T", msg.Wasm.ClearAdmin))
		default:
			span.SetTag(tagWasmMsgType, "unknown")
		}
	}
}

func addTagsFromEvents(span opentracing.Span, events sdk.Events) {
	for _, e := range events {
		if e.Type == wasmtypes.WasmModuleEventType ||
			strings.HasPrefix(e.Type, wasmtypes.CustomContractEventPrefix) ||
			e.Type == sdk.EventTypeMessage {
			for _, a := range e.Attributes {
				if string(a.Key) == wasmtypes.AttributeKeyContractAddr {
					span.SetTag(tagContract, string(a.Value))
				}
				if string(a.Key) == wasmtypes.AttributeKeyCodeID {
					span.SetTag(tagCodeID, string(a.Value))
				}
			}
		}
	}
}
