package testing

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ wasmtypes.ContractOpsKeeper = &ContractOpsKeeperMock{}

// ContractOpsKeeperMock implements wasmtypes.ContractOpsKeeper for testing purpose
type ContractOpsKeeperMock struct {
	CreateFn                   func(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string, instantiateAccess *wasmtypes.AccessConfig) (codeID uint64, err error)
	InstantiateFn              func(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error)
	ExecuteFn                  func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) (*sdk.Result, error)
	MigrateFn                  func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newCodeID uint64, msg []byte) (*sdk.Result, error)
	UpdateContractAdminFn      func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newAdmin sdk.AccAddress) error
	ClearContractAdminFn       func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress) error
	PinCodeFn                  func(ctx sdk.Context, codeID uint64) error
	UnpinCodeFn                func(ctx sdk.Context, codeID uint64) error
	SetContractInfoExtensionFn func(ctx sdk.Context, contract sdk.AccAddress, extra wasmtypes.ContractInfoExtension) error
}

func (m ContractOpsKeeperMock) Create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string, instantiateAccess *wasmtypes.AccessConfig) (codeID uint64, err error) {
	if m.CreateFn == nil {
		panic("not expected to be called")
	}
	return m.CreateFn(ctx, creator, wasmCode, source, builder, instantiateAccess)
}

func (m ContractOpsKeeperMock) Instantiate(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error) {
	if m.InstantiateFn == nil {
		panic("not expected to be called")
	}
	return m.InstantiateFn(ctx, codeID, creator, admin, initMsg, label, deposit)
}

func (m ContractOpsKeeperMock) Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) (*sdk.Result, error) {
	if m.ExecuteFn == nil {
		panic("not expected to be called")
	}
	return m.ExecuteFn(ctx, contractAddress, caller, msg, coins)
}

func (m ContractOpsKeeperMock) Migrate(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newCodeID uint64, msg []byte) (*sdk.Result, error) {
	if m.MigrateFn == nil {
		panic("not expected to be called")
	}
	return m.MigrateFn(ctx, contractAddress, caller, newCodeID, msg)
}

func (m ContractOpsKeeperMock) UpdateContractAdmin(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, newAdmin sdk.AccAddress) error {
	if m.UpdateContractAdminFn == nil {
		panic("not expected to be called")
	}
	return m.UpdateContractAdminFn(ctx, contractAddress, caller, newAdmin)
}

func (m ContractOpsKeeperMock) ClearContractAdmin(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress) error {
	if m.ClearContractAdminFn == nil {
		panic("not expected to be called")
	}
	return m.ClearContractAdminFn(ctx, contractAddress, caller)
}

func (m ContractOpsKeeperMock) PinCode(ctx sdk.Context, codeID uint64) error {
	if m.PinCodeFn == nil {
		panic("not expected to be called")
	}
	return m.PinCodeFn(ctx, codeID)
}

func (m ContractOpsKeeperMock) UnpinCode(ctx sdk.Context, codeID uint64) error {
	if m.UnpinCodeFn == nil {
		panic("not expected to be called")
	}
	return m.UnpinCodeFn(ctx, codeID)
}

func (m ContractOpsKeeperMock) SetContractInfoExtension(ctx sdk.Context, contract sdk.AccAddress, extra wasmtypes.ContractInfoExtension) error {
	if m.SetContractInfoExtensionFn == nil {
		panic("not expected to be called")
	}
	return m.SetContractInfoExtensionFn(ctx, contract, extra)
}

type CapturedCreateCalls struct {
	Creator           sdk.AccAddress
	WasmCode          []byte
	Source            string
	Builder           string
	InstantiateAccess *wasmtypes.AccessConfig
}

// CaptureCreateFn records all calls in the returned slice
func CaptureCreateFn() (func(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string, instantiateAccess *wasmtypes.AccessConfig) (codeID uint64, err error), *[]CapturedCreateCalls) {
	var nextCodeID uint64
	var captured []CapturedCreateCalls
	return func(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string, instantiateAccess *wasmtypes.AccessConfig) (codeID uint64, err error) {
		captured = append(captured, CapturedCreateCalls{Creator: creator, WasmCode: wasmCode, Source: source, Builder: builder, InstantiateAccess: instantiateAccess})
		nextCodeID++
		return nextCodeID, nil
	}, &captured
}

type CapturedInstantiateCalls struct {
	CodeID         uint64
	Creator, Admin sdk.AccAddress
	InitMsg        []byte
	Label          string
	Deposit        sdk.Coins
}

// DefaultCaptureInstantiateFnCodeID value used for building the contract address
const DefaultCaptureInstantiateFnCodeID uint64 = 0

// CaptureInstantiateFn records all calls in the returned slice
func CaptureInstantiateFn() (func(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error), *[]CapturedInstantiateCalls) {
	var nextInstanceID uint64
	var captured []CapturedInstantiateCalls
	return func(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error) {
		captured = append(captured, CapturedInstantiateCalls{CodeID: codeID, Creator: creator, Admin: admin, InitMsg: initMsg, Label: label, Deposit: deposit})
		nextInstanceID++
		return twasm.ContractAddress(DefaultCaptureInstantiateFnCodeID, nextInstanceID), nil, nil
	}, &captured
}

type CapturedExecuteCalls struct {
	ContractAddress sdk.AccAddress
	Caller          sdk.AccAddress
	Msg             []byte
	Coins           sdk.Coins
}

// CaptureExecuteFn records all calls in the returned slice
func CaptureExecuteFn() (func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) (*sdk.Result, error), *[]CapturedExecuteCalls) {
	var captured []CapturedExecuteCalls
	return func(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) (*sdk.Result, error) {
		captured = append(captured, CapturedExecuteCalls{ContractAddress: contractAddress, Caller: caller, Msg: msg, Coins: coins})
		return nil, nil
	}, &captured
}

func CapturePinCodeFn() (func(ctx sdk.Context, codeID uint64) error, *[]uint64) {
	var captured []uint64
	return func(ctx sdk.Context, codeID uint64) error {
		captured = append(captured, codeID)
		return nil
	}, &captured
}
