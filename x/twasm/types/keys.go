package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of this module.
	ModuleName = wasmtypes.ModuleName

	// RouterKey is used to route governance proposals
	RouterKey = wasmtypes.RouterKey

	// StoreKey is the prefix under which we store this module's data
	StoreKey = wasmtypes.StoreKey
)

// nolint
var (
	// leave enough space to not conflict with wasm prefixes

	PrivilegedContractsSecondaryIndexPrefix = []byte{0xa0}
	ContractCallbacksSecondaryIndexPrefix   = []byte{0xa1}
)

func GetPrivilegedContractsSecondaryIndexKey(contractAddr sdk.AccAddress) []byte {
	return append(PrivilegedContractsSecondaryIndexPrefix, contractAddr...)
}

// GetContractCallbacksSecondaryIndexKey returns the key for privileged contract callbacks
// `<prefix><callbackType><position><contractAddr>`
func GetContractCallbacksSecondaryIndexKey(callbackType PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) []byte {
	prefix := GetContractCallbacksSecondaryIndexPrefix(callbackType)
	prefixLen := len(prefix)
	const posLen = 1 // 1 byte for position
	r := make([]byte, prefixLen+posLen+sdk.AddrLen)
	copy(r[0:], prefix)
	copy(r[prefixLen:], []byte{pos})
	copy(r[prefixLen+posLen:], contractAddr)
	return r
}

// GetContractCallbacksSecondaryIndexPrefix return `<prefix><callbackType>`
func GetContractCallbacksSecondaryIndexPrefix(callbackType PrivilegedCallbackType) []byte {
	return append(ContractCallbacksSecondaryIndexPrefix, byte(callbackType))
}
