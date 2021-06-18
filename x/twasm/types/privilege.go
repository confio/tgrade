package types

import (
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// PrivilegeType is a system callback to a contract
type PrivilegeType byte

var (
	// PrivilegeTypeEmpty is empty value
	PrivilegeTypeEmpty PrivilegeType = 0
	// PrivilegeTypeBeginBlock called every block before the TX are processed
	// Multiple contracts can register for this callback privilege
	PrivilegeTypeBeginBlock = registerCallbackType(0x1, "begin_blocker", false)

	// PrivilegeTypeEndBlock called every block after the TX are processed
	// Multiple contracts can register for this callback privilege
	PrivilegeTypeEndBlock = registerCallbackType(0x2, "end_blocker", false)

	// PrivilegeTypeValidatorSetUpdate end-blocker that can modify the validator set
	// This callback privilege is exclusive to one contract instance, only.
	PrivilegeTypeValidatorSetUpdate = registerCallbackType(0x3, "validator_set_updater", true)

	// PrivilegeTypeGovProposalExecutor
	// This is a permission privilege to execute governance proposals.
	PrivilegeTypeGovProposalExecutor = registerCallbackType(0x4, "gov_proposal_executor", false)

	// PrivilegeTypeTokenMinter
	// This is a permission to mint native tokens on the chain.
	PrivilegeTypeTokenMinter = registerCallbackType(0x5, "token_minter", false)
)

var (
	// callbackTypeToString stores the string representation for every type
	callbackTypeToString = make(map[PrivilegeType]string)
	// singleInstanceCallbackTypes stores a flag for singleton instances only
	singleInstanceCallbackTypes = make(map[PrivilegeType]struct{})
)

// registerCallbackType internal method to register callback types with meta data.
func registerCallbackType(i uint8, name string, singleton bool) PrivilegeType {
	if i == 0 {
		panic("unique number must not be empty")
	}
	r := PrivilegeType(i)
	if _, exists := callbackTypeToString[r]; exists {
		panic(fmt.Sprintf("type exists already: %d", i))
	}
	if PrivilegeTypeFrom(name) != nil {
		panic(fmt.Sprintf("name exists already: %q", name))
	}
	callbackTypeToString[r] = name
	if singleton {
		singleInstanceCallbackTypes[r] = struct{}{}
	}
	return r
}

// PrivilegeTypeFrom convert name to type. Returns nil when none matches
func PrivilegeTypeFrom(name string) *PrivilegeType {
	for k, v := range callbackTypeToString {
		if v == name {
			return &k
		}
	}
	return nil
}

// AllPrivilegeTypeNames returns a list of all callback type names
func AllPrivilegeTypeNames() []string {
	result := make([]string, 0, len(callbackTypeToString))
	for _, v := range callbackTypeToString {
		result = append(result, v)
	}
	return result
}

func (t PrivilegeType) String() string {
	return callbackTypeToString[t]
}

// IsSingleton returns if only a single contract instance for this type can register (true) or multiple (false)
func (t PrivilegeType) IsSingleton() bool {
	_, ok := singleInstanceCallbackTypes[t]
	return ok
}

// ValidateBasic checks if the callback type was registered
func (t PrivilegeType) ValidateBasic() error {
	if _, ok := callbackTypeToString[t]; !ok {
		return wasmtypes.ErrInvalid
	}
	return nil
}

var _ json.Unmarshaler = &PrivilegeTypeBeginBlock
var _ json.Marshaler = &PrivilegeTypeBeginBlock

func (t *PrivilegeType) UnmarshalJSON(raw []byte) error {
	var src string
	if err := json.Unmarshal(raw, &src); err != nil {
		return err
	}
	if len(src) == 0 {
		return wasmtypes.ErrInvalid
	}
	if v := PrivilegeTypeFrom(src); v != nil {
		*t = *v
	}
	return nil
}

func (t PrivilegeType) MarshalJSON() ([]byte, error) {
	if t == 0 {
		return nil, nil
	}
	if err := t.ValidateBasic(); err != nil {
		return nil, err
	}
	return json.Marshal(t.String())
}
