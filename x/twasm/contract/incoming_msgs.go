package contract

import (
	"encoding/json"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// TgradeMsg messages coming from a contract
type TgradeMsg struct {
	Hooks              *Hooks              `json:"hooks"`
	ExecuteGovProposal *ExecuteGovProposal `json:"execute_gov_proposal"`
}

// UnmarshalWithAny from json to Go objects with cosmos-sdk Any types that have their objects/ interfaces unpacked and
// set in the `cachedValue` attribute.
func (p *TgradeMsg) UnmarshalWithAny(bz []byte, unpacker codectypes.AnyUnpacker) error {
	if err := json.Unmarshal(bz, p); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	// unpack interfaces in protobuf Any types
	if p.ExecuteGovProposal != nil {
		return sdkerrors.Wrap(p.ExecuteGovProposal.unpackInterfaces(unpacker), "execute_gov_proposal")
	}
	return nil
}

// Hooks contains method to interact with system callbacks
type Hooks struct {
	RegisterBeginBlock   *struct{} `json:"register_begin_block"`
	UnregisterBeginBlock *struct{} `json:"unregister_begin_block"`
	// these are called the end of every block
	RegisterEndBlock   *struct{} `json:"register_end_block"`
	UnregisterEndBlock *struct{} `json:"unregister_end_block"`
	// only max 1 contract can be registered here, this is called in EndBlock (after everything else) and can change the validator set.
	RegisterValidatorSetUpdate   *struct{} `json:"register_validator_set_update"`
	UnregisterValidatorSetUpdate *struct{} `json:"unregister_validator_set_update"`

	RegisterGovProposalExecutor   *struct{} `json:"register_gov_proposal_executor"`
	UnregisterGovProposalExecutor *struct{} `json:"unregister_gov_proposal_executor"`
}

// ExecuteGovProposal will execute an approved proposal in the Cosmos SDK "Gov Router".
// That allows access to many of the system internals, like sdk params or x/upgrade,
// as well as privileged access to the wasm module (eg. mark module privileged)
type ExecuteGovProposal struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Proposal    GovProposal `json:"proposal"`
}

// GetProposalContent converts message payload to gov content type. returns `nil` when unknown.
// The response is not guaranteed to be valid content.
func (p ExecuteGovProposal) GetProposalContent() govtypes.Content {
	switch {
	case p.Proposal.Text != nil:
		p.Proposal.Text.Title = p.Title
		p.Proposal.Text.Description = p.Description
		return p.Proposal.Text
	case p.Proposal.RegisterUpgrade != nil:
		return &upgradetypes.SoftwareUpgradeProposal{
			Title:       p.Title,
			Description: p.Description,
			Plan:        *p.Proposal.RegisterUpgrade,
		}
	case p.Proposal.CancelUpgrade != nil:
		p.Proposal.CancelUpgrade.Title = p.Title
		p.Proposal.CancelUpgrade.Description = p.Description
		return p.Proposal.CancelUpgrade
	case p.Proposal.ChangeParams != nil:
		return &proposaltypes.ParameterChangeProposal{
			Title:       p.Title,
			Description: p.Description,
			Changes:     *p.Proposal.ChangeParams,
		}
	case p.Proposal.IBCClientUpdate != nil:
		p.Proposal.IBCClientUpdate.Title = p.Title
		p.Proposal.IBCClientUpdate.Description = p.Description
		return p.Proposal.IBCClientUpdate
	case p.Proposal.PromoteToPrivilegedContract != nil:
		p.Proposal.PromoteToPrivilegedContract.Title = p.Title
		p.Proposal.PromoteToPrivilegedContract.Description = p.Description
		return p.Proposal.PromoteToPrivilegedContract
	case p.Proposal.DemotePrivilegedContract != nil:
		p.Proposal.DemotePrivilegedContract.Title = p.Title
		p.Proposal.DemotePrivilegedContract.Description = p.Description
		return p.Proposal.DemotePrivilegedContract
	case p.Proposal.InstantiateContract != nil:
		p.Proposal.InstantiateContract.Title = p.Title
		p.Proposal.InstantiateContract.Description = p.Description
		return p.Proposal.InstantiateContract
	case p.Proposal.MigrateContract != nil:
		p.Proposal.MigrateContract.Title = p.Title
		p.Proposal.MigrateContract.Description = p.Description
		return p.Proposal.MigrateContract
	case p.Proposal.SetContractAdmin != nil:
		p.Proposal.SetContractAdmin.Title = p.Title
		p.Proposal.SetContractAdmin.Description = p.Description
		return p.Proposal.SetContractAdmin
	case p.Proposal.ClearContractAdmin != nil:
		p.Proposal.ClearContractAdmin.Title = p.Title
		p.Proposal.ClearContractAdmin.Description = p.Description
		return p.Proposal.ClearContractAdmin
	case p.Proposal.PinCodes != nil:
		p.Proposal.PinCodes.Title = p.Title
		p.Proposal.PinCodes.Description = p.Description
		return p.Proposal.PinCodes
	case p.Proposal.UnpinCodes != nil:
		p.Proposal.UnpinCodes.Title = p.Title
		p.Proposal.UnpinCodes.Description = p.Description
		return p.Proposal.UnpinCodes
	default:
		return nil
	}
}

// unpackInterfaces unpacks the Any type into the interface type in `Any.cachedValue`
func (p *ExecuteGovProposal) unpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var err error
	switch {
	case p.Proposal.RegisterUpgrade != nil:
		return p.Proposal.RegisterUpgrade.UnpackInterfaces(unpacker)
	case p.Proposal.IBCClientUpdate != nil:
		return p.Proposal.IBCClientUpdate.UnpackInterfaces(unpacker)
	}
	return err
}

// ProtoAny data type to map from json to cosmos-sdk Any type.
type ProtoAny struct {
	TypeUrl string `json:"type_url"`
	Value   []byte `json:"value"`
}

// Encode convertes to a cosmos-sdk Any type.
func (a ProtoAny) Encode() *codectypes.Any {
	return &codectypes.Any{
		TypeUrl: a.TypeUrl,
		Value:   a.Value,
	}
}

// GovProposal bridge to unmarshal json to proposal content types
type GovProposal struct {
	proposalContent
}

// UnmarshalJSON is a custom unmarshaler that supports the cosmos-sdk Any types.
func (p *GovProposal) UnmarshalJSON(b []byte) error {
	var raws map[string]json.RawMessage
	if err := json.Unmarshal(b, &raws); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// sdk protobuf Any types don't map back nicely to Go structs. So we do this manually
	var result GovProposal
	var customUnmarshalers = map[string]func(b []byte) error{
		"ibc_client_update": func(b []byte) error {
			var proxy = struct {
				ClientId string    `json:"client_id"`
				Header   *ProtoAny `json:"header"`
			}{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.IBCClientUpdate = &ibcclienttypes.ClientUpdateProposal{
				ClientId: proxy.ClientId,
				Header:   proxy.Header.Encode(),
			}
			return nil
		},
		"register_upgrade": func(b []byte) error {
			var proxy = struct {
				Name                string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
				Height              int64    `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
				Info                string   `protobuf:"bytes,4,opt,name=info,proto3" json:"info,omitempty"`
				UpgradedClientState ProtoAny `protobuf:"bytes,5,opt,name=upgraded_client_state,json=upgradedClientState,proto3" json:"upgraded_client_state,omitempty" yaml:"upgraded_client_state"`
			}{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.RegisterUpgrade = &upgradetypes.Plan{
				Name:                proxy.Name,
				Height:              proxy.Height,
				Info:                proxy.Info,
				UpgradedClientState: proxy.UpgradedClientState.Encode(),
			}
			return nil
		},
	}
	for field, unmarshaler := range customUnmarshalers {
		if bz, ok := raws[field]; ok {
			if err := unmarshaler(bz); err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "proposal: %q: %s", field, err.Error())
			}
			*p = result
			return nil
		}
	}
	// default: use vanilla json unmarshaler when no custom one exists
	if err := json.Unmarshal(b, &result.proposalContent); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	*p = result
	return nil
}

// proposalContent contains the concrete cosmos-sdk/ tgrade gov proposal types
type proposalContent struct {
	// Signaling proposal, the text and description field will be recorded
	Text *govtypes.TextProposal `json:"text"`

	// Register an "live upgrade" on the x/upgrade module
	// See https://github.com/cosmos/cosmos-sdk/blob/v0.42.3/proto/cosmos/upgrade/v1beta1/upgrade.proto#L12-L53
	RegisterUpgrade *upgradetypes.Plan `json:"register_upgrade"`

	// There can only be one pending upgrade at a given time. This cancels the pending upgrade, if any.
	// See https://github.com/cosmos/cosmos-sdk/blob/v0.42.3/proto/cosmos/upgrade/v1beta1/upgrade.proto#L57-L62
	CancelUpgrade *upgradetypes.CancelSoftwareUpgradeProposal `json:"cancel_upgrade"`

	// Defines a proposal to change one or more parameters.
	// See https://github.com/cosmos/cosmos-sdk/blob/v0.42.3/proto/cosmos/params/v1beta1/params.proto#L9-L27
	ChangeParams *[]proposaltypes.ParamChange `json:"change_params"`

	// Updates the matching client to set a new trusted header.
	// This can be used by governance to restore a client that has timed out or forked or otherwise broken.
	// See https://github.com/cosmos/cosmos-sdk/blob/v0.42.3/proto/ibc/core/client/v1/client.proto#L36-L49
	IBCClientUpdate *ibcclienttypes.ClientUpdateProposal `json:"ibc_client_update"`

	// See https://github.com/confio/tgrade/blob/privileged_contracts_5/proto/confio/twasm/v1beta1/proposal.proto
	PromoteToPrivilegedContract *types.PromoteToPrivilegedContractProposal `json:"promote_to_privileged_contract"`

	// See https://github.com/confio/tgrade/blob/privileged_contracts_5/proto/confio/twasm/v1beta1/proposal.proto
	DemotePrivilegedContract *types.DemotePrivilegedContractProposal `json:"demote_privileged_contract"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L32-L54
	InstantiateContract *wasmtypes.InstantiateContractProposal `json:"instantiate_contract"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L56-L70
	MigrateContract *wasmtypes.MigrateContractProposal `json:"migrate_contract"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L72-L82
	SetContractAdmin *wasmtypes.UpdateAdminProposal `json:"set_contract_admin"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L84-L93
	ClearContractAdmin *wasmtypes.ClearAdminProposal `json:"clear_contract_admin"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L95-L107
	PinCodes *wasmtypes.PinCodesProposal `json:"pin_codes"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1beta1/proposal.proto#L109-L121
	UnpinCodes *wasmtypes.UnpinCodesProposal `json:"unpin_codes"`
}
