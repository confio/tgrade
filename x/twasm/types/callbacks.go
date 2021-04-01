package types

import "fmt"

// PriviledgedCallbackType is a system callback to a contract
type PriviledgedCallbackType byte

var (
	// CallbackTypeBeginBlock called every block before the TX are processed
	CallbackTypeBeginBlock = registerCallbackType(0x1, "begin_block")
	// CallbackTypeEndBlock called every block after the TX are processed
	CallbackTypeEndBlock = registerCallbackType(0x2, "end_block")
	// CallbackTypeValidatorSetUpdate end-blocker that can modify the validator set
	CallbackTypeValidatorSetUpdate = registerCallbackType(0x3, "validator_Set_update")
)
var callbackTypeToString = make(map[PriviledgedCallbackType]string)

func registerCallbackType(i uint8, name string) PriviledgedCallbackType {
	r := PriviledgedCallbackType(i)
	if _, exists := callbackTypeToString[r]; exists {
		panic(fmt.Sprintf("type exists already: %d", i))
	}
	callbackTypeToString[r] = name
	return r
}

func (t PriviledgedCallbackType) String() string {
	return callbackTypeToString[t]
}
