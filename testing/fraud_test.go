//go:build system_test
// +build system_test

package testing

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"

	testingcontract "github.com/confio/tgrade/testing/contract"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/stretchr/testify/require"
)

func TestRecursiveMsgsExternalTrigger(t *testing.T) {
	sut.ResetDirtyChain(t)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	codeID := cli.StoreWasm("testing/contract/hackatom.wasm")
	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	contractAddr := cli.InstantiateWasm(codeID, initMsg)

	specs := map[string]struct {
		gas           string
		expErrMatcher func(t require.TestingT, err error, msgAndArgs ...interface{})
	}{
		"simulation": {
			gas:           "auto",
			expErrMatcher: ErrOutOfGasMatcher,
		},
		"tx": { // tx will be rejected by Tendermint in post abci checkTX operation
			gas:           strconv.Itoa(math.MaxInt64),
			expErrMatcher: ErrTimeoutMatcher,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			execMsg := `{"message_loop":{}}`
			for _, n := range sut.AllNodes(t) {
				cli.WithRunErrorMatcher(spec.expErrMatcher).WithNodeAddress(n.RPCAddr()).
					Execute(contractAddr, execMsg, defaultSrcAddr, "--gas="+spec.gas, "--broadcast-mode=async")
			}
			sut.AwaitNextBlock(t)
		})
	}
}

func TestRecursiveSmartQuery(t *testing.T) {
	sut.ResetDirtyChain(t)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	maliciousContractAddr := cli.InstantiateWasm(cli.StoreWasm("testing/contract/hackatom.wasm"), initMsg)

	msg := fmt.Sprintf(`{"recurse":{"depth":%d, "work":0}}`, math.MaxUint32)

	// when
	for _, n := range sut.AllNodes(t) {
		cli.WithRunErrorMatcher(ErrOutOfGasMatcher).WithNodeAddress(n.RPCAddr()).
			QuerySmart(maliciousContractAddr, msg)
	}
	sut.AwaitNextBlock(t)
}

func TestRecursiveMsgsEmittedByContractInSimulation(t *testing.T) {
	sut.ResetDirtyChain(t)
	sut.StartChain(t)
	cli := NewTgradeCli(t, sut, verbose)

	reflectContractAddr := cli.InstantiateWasm(cli.StoreWasm("testing/contract/reflect.wasm"), "{}")

	initMsg := fmt.Sprintf(`{"verifier":%q, "beneficiary":%q}`, randomBech32Addr(), randomBech32Addr())
	maliciousContractAddr := cli.InstantiateWasm(cli.StoreWasm("testing/contract/hackatom.wasm"), initMsg)

	payloadMsg := wasmvmtypes.CosmosMsg{
		Wasm: &wasmvmtypes.WasmMsg{
			Execute: &wasmvmtypes.ExecuteMsg{
				ContractAddr: maliciousContractAddr,
				Msg:          []byte(`{"message_loop":{}}`),
			},
		},
	}
	reflectMsg := testingcontract.ReflectHandleMsg{
		ReflectSubMsg: &testingcontract.ReflectSubPayload{
			Msgs: []wasmvmtypes.SubMsg{{
				ID:      1,
				Msg:     payloadMsg,
				ReplyOn: wasmvmtypes.ReplyAlways,
			}},
		},
	}

	execMsgBz, err := json.Marshal(reflectMsg)
	require.NoError(t, err)

	// when
	for _, n := range sut.AllNodes(t) {
		cli.WithRunErrorMatcher(ErrOutOfGasMatcher).WithNodeAddress(n.RPCAddr()).
			Execute(reflectContractAddr, string(execMsgBz), defaultSrcAddr, "--gas=auto")
	}
	sut.AwaitNextBlock(t)
}
