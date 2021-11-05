package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SmartQuerier interface {
	QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
}

// Sudoer with access to sudo method
type Sudoer interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

// Executor with access to execute method
type Executor interface {
	Execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error)
}

// TWasmKeeper is a subset of x/twasm keeper
type TWasmKeeper interface {
	SmartQuerier
	Sudoer
	GetContractKeeper() wasmtypes.ContractOpsKeeper
}
