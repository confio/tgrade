package tracing

import (
	"encoding/base64"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/opentracing/opentracing-go"
	opentractinglog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
	"time"
)

func TestDispatchWithTracingStore(t *testing.T) {
	tracerEnabled = true
	var (
		myKey                   = []byte(`foo`)
		myVal                   = []byte(`bar`)
		randAddr sdk.AccAddress = rand.Bytes(sdk.AddrLen)
	)
	t.Cleanup(func() { tracerEnabled = false })

	ctx, enc, storeKey := createMinTestInput(t)
	xctx, commit := ctx.CacheContext()
	mock := wasmtesting.MockMessageHandler{DispatchMsgFn: func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
		ctx.KVStore(storeKey).Set(myKey, myVal)
		return nil, nil, nil
	}}
	m := TraceMessageHandlerDecorator(enc)(&mock)

	tracer := &MockCaptureLogsTracer{}
	opentracing.SetGlobalTracer(tracer)

	// when
	_, _, err := m.DispatchMsg(xctx, randAddr, "noIBC", wasmvmtypes.CosmosMsg{Bank: &wasmvmtypes.BankMsg{Send: &wasmvmtypes.SendMsg{
		ToAddress: randAddr.String(),
		Amount:    wasmvmtypes.Coins{wasmvmtypes.NewCoin(123, "ALX")},
	}}})
	commit()
	require.NoError(t, err)
	require.Len(t, tracer.captured, 3)
	require.Equal(t, logRawStoreIO, tracer.captured[2].Key())
	require.NotEmpty(t, tracer.captured[2].Value())
	encoding := base64.StdEncoding
	exp := fmt.Sprintf(`{"operation":"write","key":"%s","value":"%s","metadata":null}`, encoding.EncodeToString(myKey), encoding.EncodeToString(myVal))
	line := tracer.captured[2].Value().(string)
	assert.Equal(t, exp, line[:len(line)-1])
}

func createMinTestInput(t *testing.T) (sdk.Context, *codec.ProtoCodec, *sdk.KVStoreKey) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(legacyAmino)

	return ctx, marshaler, storeKey
}

var _ opentracing.Tracer = &MockCaptureLogsTracer{}

type MockCaptureLogsTracer struct {
	opentracing.Tracer

	captured []opentractinglog.Field
}

type Spanner struct {
	opentracing.Span
	tracer *MockCaptureLogsTracer
}

func (s Spanner) LogFields(fields ...opentractinglog.Field) {
	s.tracer.captured = append(s.tracer.captured, fields...)
}

func (m *MockCaptureLogsTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return &Spanner{tracer: m}
}
func (s Spanner) SetTag(key string, value interface{}) opentracing.Span {
	return &s
}
func (s Spanner) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s Spanner) Finish() {
}
