package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// tgradeKeeper defines a subset of Keeper
type tgradeKeeper interface {
	IsPrivileged(ctx sdk.Context, contract sdk.AccAddress) bool
	appendToPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) (uint8, error)
	removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) bool
	setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

var _ wasmkeeper.Messenger = TgradeHandler{}

// TgradeHandler is a custom message handler plugin for wasmd.
type TgradeHandler struct {
	keeper    tgradeKeeper
	govRouter govtypes.Router
	cdc       codec.Marshaler
}

// NewTgradeHandler constructor
func NewTgradeHandler(cdc codec.Marshaler, keeper tgradeKeeper, govRouter govtypes.Router) *TgradeHandler {
	return &TgradeHandler{cdc: cdc, keeper: keeper, govRouter: govRouter}
}

// DispatchMsg handles wasmVM message for privileged contracts
func (h TgradeHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	if !h.keeper.IsPrivileged(ctx, contractAddr) {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	var tMsg contract.TgradeMsg
	if err := tMsg.UnmarshalWithAny(msg.Custom, h.cdc); err != nil {
		return nil, nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	switch {
	case tMsg.Hooks != nil:
		return nil, nil, h.handleHooks(ctx, contractAddr, tMsg.Hooks)
	case tMsg.ExecuteGovProposal != nil:
		return nil, nil, h.handleGovProposalExecution(ctx, contractAddr, tMsg.ExecuteGovProposal)
	}
	return nil, nil, wasmtypes.ErrUnknownMsg
}

// handle register/ unregister hook messages
func (h TgradeHandler) handleHooks(ctx sdk.Context, contractAddr sdk.AccAddress, hooks *contract.Hooks) error {
	contractInfo := h.keeper.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return err
	}

	register := func(c types.PrivilegedCallbackType) error {
		if details.HasRegisteredContractCallback(c) {
			return nil
		}
		pos, err := h.keeper.appendToPrivilegedContractCallbacks(ctx, c, contractAddr)
		if err != nil {
			return sdkerrors.Wrap(err, "callback registration")
		}
		details.AddRegisteredCallback(c, pos)
		return sdkerrors.Wrap(h.keeper.setContractDetails(ctx, contractAddr, &details), "store details")
	}
	unregister := func(tp types.PrivilegedCallbackType) error {
		if !details.HasRegisteredContractCallback(tp) {
			return nil
		}
		details.IterateRegisteredCallbacks(func(c types.PrivilegedCallbackType, pos uint8) bool {
			if c != tp {
				return false
			}
			h.keeper.removePrivilegedContractCallbacks(ctx, c, pos, contractAddr)
			details.RemoveRegisteredCallback(c, pos)
			return false
		})
		return sdkerrors.Wrap(h.keeper.setContractDetails(ctx, contractAddr, &details), "store details")
	}
	switch {
	case hooks.RegisterBeginBlock != nil:
		return register(types.CallbackTypeBeginBlock)
	case hooks.UnregisterBeginBlock != nil:
		return unregister(types.CallbackTypeBeginBlock)
	case hooks.RegisterEndBlock != nil:
		return register(types.CallbackTypeEndBlock)
	case hooks.UnregisterEndBlock != nil:
		return unregister(types.CallbackTypeEndBlock)
	case hooks.RegisterValidatorSetUpdate != nil:
		return register(types.CallbackTypeValidatorSetUpdate)
	case hooks.UnregisterValidatorSetUpdate != nil:
		return unregister(types.CallbackTypeValidatorSetUpdate)
	case hooks.RegisterGovProposalExecutor != nil:
		return register(types.CallbackTypeGovProposalExecutor)
	case hooks.UnregisterGovProposalExecutor != nil:
		return unregister(types.CallbackTypeGovProposalExecutor)
	default:
		return wasmtypes.ErrUnknownMsg
	}
}

// handle gov proposal execution
func (h TgradeHandler) handleGovProposalExecution(ctx sdk.Context, contractAddr sdk.AccAddress, exec *contract.ExecuteGovProposal) error {
	contractInfo := h.keeper.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return err
	}
	if !details.HasRegisteredContractCallback(types.CallbackTypeGovProposalExecutor) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "requires: %s", types.CallbackTypeGovProposalExecutor.String())
	}

	content := exec.GetProposalContent()
	if content == nil {
		return sdkerrors.Wrap(wasmtypes.ErrUnknownMsg, "unsupported content type")
	}
	if err := content.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "content")
	}
	if !h.govRouter.HasRoute(content.ProposalRoute()) {
		return sdkerrors.Wrap(govtypes.ErrNoProposalHandlerExists, content.ProposalRoute())
	}
	govHandler := h.govRouter.GetRoute(content.ProposalRoute())
	return govHandler(ctx, content)
}
