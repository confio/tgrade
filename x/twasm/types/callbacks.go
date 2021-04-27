package types

import (
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// PrivilegedCallbackType is a system callback to a contract
type PrivilegedCallbackType byte

var (
	// CallbackTypeBeginBlock called every block before the TX are processed
	// Multiple contracts can register for this callback
	CallbackTypeBeginBlock = registerCallbackType(0x1, "begin_block", false)
	// CallbackTypeEndBlock called every block after the TX are processed
	// Multiple contracts can register for this callback
	CallbackTypeEndBlock = registerCallbackType(0x2, "end_block", false)
	// CallbackTypeValidatorSetUpdate end-blocker that can modify the validator set
	// This callback is exclusive to one contract instance, only.
	CallbackTypeValidatorSetUpdate = registerCallbackType(0x3, "validator_set_update", true)
)

var (
	// callbackTypeToString stores the string representation for every type
	callbackTypeToString = make(map[PrivilegedCallbackType]string)
	// singleInstanceCallbackTypes stores a flag for singleton instances only
	singleInstanceCallbackTypes = make(map[PrivilegedCallbackType]struct{})
)

// registerCallbackType internal method to register callback types with meta data.
func registerCallbackType(i uint8, name string, singleton bool) PrivilegedCallbackType {
	r := PrivilegedCallbackType(i)
	if _, exists := callbackTypeToString[r]; exists {
		panic(fmt.Sprintf("type exists already: %d", i))
	}
	if PrivilegedCallbackTypeFrom(name) != nil {
		panic(fmt.Sprintf("name exists already: %q", name))
	}
	callbackTypeToString[r] = name
	if singleton {
		singleInstanceCallbackTypes[r] = struct{}{}
	}
	return r
}

// PrivilegedCallbackTypeFrom convert name to type. Returns nil when none matches
func PrivilegedCallbackTypeFrom(name string) *PrivilegedCallbackType {
	for k, v := range callbackTypeToString {
		if v == name {
			return &k
		}
	}
	return nil
}

// AllCallbackTypeNames returns a list of all callback type names
func AllCallbackTypeNames() []string {
	result := make([]string, 0, len(callbackTypeToString))
	for _, v := range callbackTypeToString {
		result = append(result, v)
	}
	return result
}

func (t PrivilegedCallbackType) String() string {
	return callbackTypeToString[t]
}

// IsSingleton returns if only a single contract instance for this type can register (true) or multiple (false)
func (t PrivilegedCallbackType) IsSingleton() bool {
	_, ok := singleInstanceCallbackTypes[t]
	return ok
}

// ValidateBasic checks if the callback type was registered
func (t PrivilegedCallbackType) ValidateBasic() error {
	if _, ok := callbackTypeToString[t]; !ok {
		return wasmtypes.ErrInvalid
	}
	return nil
}
