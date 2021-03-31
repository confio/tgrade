package types

// PriviledgedCallbackType is a system callback to a contract
type PriviledgedCallbackType byte

const CallbackTypeBeginBlock PriviledgedCallbackType = 0x1
const CallbackTypeEndBlock PriviledgedCallbackType = 0x2

// CallbackTypeValidatorSetUpdate is as last section in end-blocker
const CallbackTypeValidatorSetUpdate PriviledgedCallbackType = 0x3
