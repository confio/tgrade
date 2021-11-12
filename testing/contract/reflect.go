package contract

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type ReflectHandleMsg struct {
	ReflectSubMsg *ReflectSubPayload `json:"reflect_sub_msg,omitempty"`
}

type ReflectSubPayload struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}
