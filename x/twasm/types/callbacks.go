package types

import "fmt"

// PrivilegedCallbackType is a system callback to a contract
type PrivilegedCallbackType byte

var (
	// CallbackTypeBeginBlock called every block before the TX are processed
	CallbackTypeBeginBlock = registerCallbackType(0x1, "begin_block")
	// CallbackTypeEndBlock called every block after the TX are processed
	CallbackTypeEndBlock = registerCallbackType(0x2, "end_block")
	// CallbackTypeValidatorSetUpdate end-blocker that can modify the validator set
	CallbackTypeValidatorSetUpdate = registerCallbackType(0x3, "validator_set_update")
)

var callbackTypeToString = make(map[PrivilegedCallbackType]string)

func registerCallbackType(i uint8, name string) PrivilegedCallbackType {
	r := PrivilegedCallbackType(i)
	if _, exists := callbackTypeToString[r]; exists {
		panic(fmt.Sprintf("type exists already: %d", i))
	}
	if PrivilegedCallbackTypeFrom(name) != nil {
		panic(fmt.Sprintf("name exists already: %q", name))
	}
	callbackTypeToString[r] = name
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
