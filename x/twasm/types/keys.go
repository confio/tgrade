package types

import (
	"fmt"
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

// todo (Alex): maybe define types in protobuf enums instead?

// PriviledgedCallbackType is a system callback to a contract
type PriviledgedCallbackType byte

const CallbackTypeBeginBlock PriviledgedCallbackType = 0x1
const CallbackTypeEndBlock PriviledgedCallbackType = 0x2

// CallbackTypeValidatorSetUpdate is as last section in endblocker
const CallbackTypeValidatorSetUpdate PriviledgedCallbackType = 0x3

func GetPrivilegedContractsSecondaryIndexKey(contractAddr sdk.AccAddress) []byte {
	return append(PrivilegedContractsSecondaryIndexPrefix, contractAddr...)
}

// GetContractCallbacksSecondaryIndexKey returns the key for priviledged contract callbacks
// `<prefix><callbackType><position><contractAddr>`
func GetContractCallbacksSecondaryIndexKey(callbackType PriviledgedCallbackType, pos uint64, contractAddr sdk.AccAddress) []byte {
	prefix := GetContractCallbacksSecondaryIndexPrefix(callbackType)
	prefixLen := len(prefix)
	const posLen = 8 // todo (reviewer): can be smaller than uint64. what would be a good max?
	r := make([]byte, prefixLen+posLen+sdk.AddrLen)
	copy(r[0:], prefix)
	copy(r[prefixLen:], sdk.Uint64ToBigEndian(pos))
	copy(r[prefixLen+posLen:], contractAddr)
	return r
}

func GetContractCallbacksSecondaryIndexPrefix(callbackType PriviledgedCallbackType) []byte {
	return append(ContractCallbacksSecondaryIndexPrefix, byte(callbackType))
}

func SplitUnprefixedContractCallbacksSecondaryIndexKey(s []byte) (PriviledgedCallbackType, uint64, sdk.AccAddress) {
	if len(s) != 1+8+sdk.AddrLen {
		panic(fmt.Sprintf("unexpected key lenght %d", len(s)))
	}
	return PriviledgedCallbackType(s[0]), sdk.BigEndianToUint64(s[1:9]), s[9:]
}
