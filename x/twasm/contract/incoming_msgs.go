package contract

import (
	"encoding/json"
	"sort"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/confio/tgrade/x/twasm/types"
)

// TgradeMsg messages coming from a contract
type TgradeMsg struct {
	Privilege          *PrivilegeMsg          `json:"privilege,omitempty"`
	ExecuteGovProposal *ExecuteGovProposal    `json:"execute_gov_proposal,omitempty"`
	MintTokens         *MintTokens            `json:"mint_tokens,omitempty"`
	ConsensusParams    *ConsensusParamsUpdate `json:"consensus_params,omitempty"`
	Delegate           *Delegate              `json:"delegate,omitempty"`
	Undelegate         *Undelegate            `json:"undelegate,omitempty"`
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

type PrivilegeMsg struct {
	Request types.PrivilegeType `json:"request,omitempty"`
	Release types.PrivilegeType `json:"release,omitempty"`
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
func (p ExecuteGovProposal) GetProposalContent(sender sdk.AccAddress) govtypes.Content {
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
		p.Proposal.InstantiateContract.RunAs = sender.String()
		return p.Proposal.InstantiateContract
	case p.Proposal.StoreCode != nil:
		p.Proposal.StoreCode.Title = p.Title
		p.Proposal.StoreCode.Description = p.Description
		p.Proposal.StoreCode.RunAs = sender.String()
		return p.Proposal.StoreCode
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
	switch { //nolint:gocritic
	case p.Proposal.RegisterUpgrade != nil:
		// revisit with https://github.com/confio/tgrade/issues/364
		if p.Proposal.RegisterUpgrade.UpgradedClientState != nil { //nolint:staticcheck
			return sdkerrors.ErrInvalidRequest.Wrap("upgrade logic for IBC has been moved to the IBC module")
		}
	}
	return err
}

// ProtoAny data type to map from json to cosmos-sdk Any type.
type ProtoAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

// Encode converts to a cosmos-sdk Any type.
func (a ProtoAny) Encode() *codectypes.Any {
	return &codectypes.Any{
		TypeUrl: a.TypeURL,
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
	customUnmarshalers := map[string]func(b []byte) error{
		"ibc_client_update": func(b []byte) error {
			proxy := struct {
				ClientID string    `json:"client_id"`
				Header   *ProtoAny `json:"header"`
			}{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.IBCClientUpdate = &ibcclienttypes.ClientUpdateProposal{
				SubjectClientId: proxy.ClientID,
			}
			return nil
		},
		"register_upgrade": func(b []byte) error {
			proxy := struct {
				Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
				Height int64  `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
				Info   string `protobuf:"bytes,4,opt,name=info,proto3" json:"info,omitempty"`
			}{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.RegisterUpgrade = &upgradetypes.Plan{
				Name:   proxy.Name,
				Height: proxy.Height,
				Info:   proxy.Info,
			}
			return nil
		},
		"migrate_contract": func(b []byte) error {
			proxy := struct { // custom type as not name compatible with wasmvmtypes.MigrateMsg
				Contract string `json:"contract"`
				CodeID   uint64 `json:"code_id"`
				Msg      []byte `json:"migrate_msg"`
			}{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.MigrateContract = &wasmtypes.MigrateContractProposal{
				Contract: proxy.Contract,
				CodeID:   proxy.CodeID,
				Msg:      proxy.Msg,
			}
			return nil
		}, "instantiate_contract": func(b []byte) error {
			proxy := wasmvmtypes.InstantiateMsg{}
			if err := json.Unmarshal(b, &proxy); err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			funds, err := convertWasmCoinsToSdkCoins(proxy.Funds)
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			result.InstantiateContract = &wasmtypes.InstantiateContractProposal{
				// RunAs:       "",
				Admin:  proxy.Admin,
				CodeID: proxy.CodeID,
				Label:  proxy.Label,
				Msg:    proxy.Msg,
				Funds:  funds,
			}
			return nil
		},
	}
	// make deterministic
	fieldNames := make([]string, 0, len(customUnmarshalers))
	for k := range customUnmarshalers {
		fieldNames = append(fieldNames, k)
	}
	sort.Strings(fieldNames)
	for _, field := range fieldNames {
		unmarshaler := customUnmarshalers[field]
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

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L32-L54
	InstantiateContract *wasmtypes.InstantiateContractProposal `json:"instantiate_contract"`

	// see https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L14-L27
	StoreCode *wasmtypes.StoreCodeProposal `json:"store_code"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L56-L70
	MigrateContract *wasmtypes.MigrateContractProposal `json:"migrate_contract"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L72-L82
	SetContractAdmin *wasmtypes.UpdateAdminProposal `json:"set_contract_admin"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L84-L93
	ClearContractAdmin *wasmtypes.ClearAdminProposal `json:"clear_contract_admin"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L95-L107
	PinCodes *wasmtypes.PinCodesProposal `json:"pin_codes"`

	// See https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/proposal.proto#L109-L121
	UnpinCodes *wasmtypes.UnpinCodesProposal `json:"unpin_codes"`
}

// MintTokens custom message to mint native tokens on the chain.
// See https://github.com/confio/tgrade-contracts/blob/main/packages/bindings/schema/tgrade_msg.json
type MintTokens struct {
	Denom         string `json:"denom"`
	Amount        string `json:"amount"`
	RecipientAddr string `json:"recipient"`
}

// ConsensusParamsUpdate subset of tendermint params.
// See https://github.com/tendermint/tendermint/blob/v0.34.8/proto/tendermint/abci/types.proto#L282-L289
type ConsensusParamsUpdate struct {
	Block    *BlockParams    `json:"block,omitempty"`
	Evidence *EvidenceParams `json:"evidence,omitempty"`
}

// Delegate funds. Used for vesting accounts.
type Delegate struct {
	Funds      wasmvmtypes.Coin `json:"funds"`
	StakerAddr string           `json:"staker"`
}

// Undelegate funds. Used with vesting accounts.
type Undelegate struct {
	Funds         wasmvmtypes.Coin `json:"funds"`
	RecipientAddr string           `json:"recipient"`
}

// ValidateBasic check basics
func (c ConsensusParamsUpdate) ValidateBasic() error {
	if c.Block == nil && c.Evidence == nil {
		return wasmtypes.ErrEmpty
	}
	if err := c.Block.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "block")
	}
	if err := c.Evidence.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "evidence")
	}
	return nil
}

type BlockParams struct {
	// MaxBytes Maximum number of bytes (over all tx) to be included in a block
	MaxBytes *int64 `json:"max_bytes,omitempty"`
	// MaxGas Maximum gas (over all tx) to be executed in one block.
	MaxGas *int64 `json:"max_gas,omitempty"`
}

// ValidateBasic check basics
func (p *BlockParams) ValidateBasic() error {
	if p == nil {
		return nil
	}
	if p.MaxBytes == nil && p.MaxGas == nil {
		return wasmtypes.ErrEmpty
	}
	return nil
}

type EvidenceParams struct {
	// MaxAgeNumBlocks Max age of evidence, in blocks.
	MaxAgeNumBlocks *int64 `json:"max_age_num_blocks,omitempty"`
	// MaxAgeDuration Max age of evidence, in seconds.
	// It should correspond with an app's "unbonding period"
	MaxAgeDuration *int64 `json:"max_age_duration,omitempty"`
	// MaxBytes Maximum number of bytes of evidence to be included in a block
	MaxBytes *int64 `json:"max_bytes,omitempty"`
}

// ValidateBasic check basics
func (p *EvidenceParams) ValidateBasic() error {
	if p == nil {
		return nil
	}
	if p.MaxAgeNumBlocks == nil && p.MaxAgeDuration == nil && p.MaxBytes == nil {
		return wasmtypes.ErrEmpty
	}
	return nil
}

// copied from wasmd. should be public soon
func convertWasmCoinsToSdkCoins(coins []wasmvmtypes.Coin) (sdk.Coins, error) {
	var toSend sdk.Coins
	for _, coin := range coins {
		c, err := convertWasmCoinToSdkCoin(coin)
		if err != nil {
			return nil, err
		}
		toSend = append(toSend, c)
	}
	return toSend, nil
}

// copied from wasmd. should be public soon
func convertWasmCoinToSdkCoin(coin wasmvmtypes.Coin) (sdk.Coin, error) {
	amount, ok := sdk.NewIntFromString(coin.Amount)
	if !ok {
		return sdk.Coin{}, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coin.Amount+coin.Denom)
	}
	r := sdk.Coin{
		Denom:  coin.Denom,
		Amount: amount,
	}
	return r, r.Validate()
}
