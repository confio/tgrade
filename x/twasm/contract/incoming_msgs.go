package contract

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TgradeMsg messages coming from a contract
type TgradeMsg struct { // todo (Alex): replace with proper msg
	Hooks              *Hooks       `json:"hooks"`
	MintTokens         *MintTokens  `json:"mint_tokens"`
	WasmSudo           *RunSudo     `json:"wasm_sudo"`
	ExecuteGovProposal *GovProposal `json:"execute_gov_proposal"`
}

type Hooks struct {
	RegisterBeginBlock   *struct{} `json:"register_begin_block"`
	UnregisterBeginBlock *struct{} `json:"unregister_begin_block"`
	// these are called the end of every block
	RegisterEndBlock   *struct{} `json:"register_end_block"`
	UnregisterEndBlock *struct{} `json:"unregister_end_block"`
	// only max 1 contract can be registered here, this is called in EndBlock (after everything else) and can change the validator set.
	RegisterValidatorSetUpdate   *struct{} `json:"register_validator_set_update"`
	UnregisterValidatorSetUpdate *struct{} `json:"unregister_validator_set_update"`
}
type MintTokens struct {
	Denom     string `json:"denom"`
	Amount    uint64 `json:"amount"` // todo (Alex): revisit types
	Recipient string `json:"recipient"`
}

type RunSudo struct {
	ContractAddr string `json:"contract_addr"`
	/// msg is the json-encoded SudoMsg struct (as raw Binary)
	Msg  []byte     `json:"msg"`
	Send []sdk.Coin `json:"send"`
}

type GovProposal struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	// mapable to gov.Content types
	Text            *struct{}
	SoftwareUpgrade *struct{}
	// all wasm lifecycle proposals, including adding permissions
	InstantiateContract *struct { /* TODO */
	}
	MigrateContract *struct { /* TODO */
	}
	SetContractAdmin *struct { /* TODO */
	}
	ClearContractAdmin *struct { /* TODO */
	}
	MakeContractPermissioned *struct { /* TODO */
	}
	RemoveContractPermissions *struct { /* TODO */
	}
	PinContract *struct { /* TODO */
	}
	UnpinContract *struct { /* TODO */
	}
	// x/upgrade callback
	SetUpgrade *struct { /* TODO */
	}
	// set parameters of any module
	SetParams *struct { /* TODO */
	}
	// participate in ibc governance
	UpdateIBCClient *struct { /* TODO */
	}
	// Allows raw bytes (if client and wasmd are aware of something the contract is not)
	// Like CosmosMsg::Stargate but for the governance router, not normal router
	RawProtoProposal *struct { /*todo: data: Binary }*/
	}
	// others???
}
