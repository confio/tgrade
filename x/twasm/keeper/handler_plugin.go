package keeper

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
)

// TgradeWasmHandlerKeeper defines a subset of Keeper
type TgradeWasmHandlerKeeper interface {
	IsPrivileged(ctx sdk.Context, contract sdk.AccAddress) bool
	appendToPrivilegedContracts(ctx sdk.Context, privilegeType types.PrivilegeType, contractAddress sdk.AccAddress) (uint8, error)
	removePrivilegeRegistration(ctx sdk.Context, privilegeType types.PrivilegeType, pos uint8, contractAddr sdk.AccAddress) bool
	setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

// minter is a subset of bank keeper
type minter interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

var _ wasmkeeper.Messenger = TgradeHandler{}

// TgradeHandler is a custom message handler plugin for wasmd.
type TgradeHandler struct {
	keeper    TgradeWasmHandlerKeeper
	minter    minter
	govRouter govtypes.Router
	cdc       codec.Marshaler
}

// NewTgradeHandler constructor
func NewTgradeHandler(cdc codec.Marshaler, keeper TgradeWasmHandlerKeeper, bankKeeper minter, govRouter govtypes.Router) *TgradeHandler {
	return &TgradeHandler{cdc: cdc, keeper: keeper, govRouter: govRouter, minter: bankKeeper}
}

// DispatchMsg handles wasmVM message for privileged contracts
func (h TgradeHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, _ string, msg wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	if !h.keeper.IsPrivileged(ctx, contractAddr) {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	em := sdk.NewEventManager()
	ctx = ctx.WithEventManager(em)
	var tMsg contract.TgradeMsg
	if err := tMsg.UnmarshalWithAny(msg.Custom, h.cdc); err != nil {
		return nil, nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	switch {
	case tMsg.Privilege != nil:
		err := h.handlePrivilege(ctx, contractAddr, tMsg.Privilege)
		return em.Events(), nil, err
	case tMsg.ExecuteGovProposal != nil:
		err := h.handleGovProposalExecution(ctx, contractAddr, tMsg.ExecuteGovProposal)
		return em.Events(), nil, err
	case tMsg.MintTokens != nil:
		evts, err := h.handleMintToken(ctx, contractAddr, tMsg.MintTokens)
		return append(evts, em.Events()...), nil, err
	}

	ModuleLogger(ctx).Info("unhandled message", "msg", msg)
	return nil, nil, wasmtypes.ErrUnknownMsg
}

// handle register/ unregister privilege messages
func (h TgradeHandler) handlePrivilege(ctx sdk.Context, contractAddr sdk.AccAddress, msg *contract.PrivilegeMsg) error {
	contractInfo := h.keeper.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return err
	}

	register := func(c types.PrivilegeType) error {
		if details.HasRegisteredPrivilege(c) {
			return nil
		}
		pos, err := h.keeper.appendToPrivilegedContracts(ctx, c, contractAddr)
		if err != nil {
			return sdkerrors.Wrap(err, "privilege registration")
		}
		details.AddRegisteredPrivilege(c, pos)
		return sdkerrors.Wrap(h.keeper.setContractDetails(ctx, contractAddr, &details), "store details")
	}
	unregister := func(tp types.PrivilegeType) error {
		if !details.HasRegisteredPrivilege(tp) {
			return nil
		}
		details.IterateRegisteredPrivileges(func(c types.PrivilegeType, pos uint8) bool {
			if c != tp {
				return false
			}
			h.keeper.removePrivilegeRegistration(ctx, c, pos, contractAddr)
			details.RemoveRegisteredPrivilege(c, pos)
			return false
		})
		return sdkerrors.Wrap(h.keeper.setContractDetails(ctx, contractAddr, &details), "store details")
	}
	switch {
	case msg.Release != types.PrivilegeTypeEmpty:
		return unregister(msg.Release)
	case msg.Request != types.PrivilegeTypeEmpty:
		return register(msg.Request)
	default:
		return wasmtypes.ErrUnknownMsg
	}
}

// handle gov proposal execution
func (h TgradeHandler) handleGovProposalExecution(ctx sdk.Context, contractAddr sdk.AccAddress, exec *contract.ExecuteGovProposal) error {
	if err := h.assertHasPrivilege(ctx, contractAddr, types.PrivilegeTypeGovProposalExecutor); err != nil {
		return err
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

// handle mint token message
func (h TgradeHandler) handleMintToken(ctx sdk.Context, contractAddr sdk.AccAddress, mint *contract.MintTokens) ([]sdk.Event, error) {
	if err := h.assertHasPrivilege(ctx, contractAddr, types.PrivilegeTypeTokenMinter); err != nil {
		return nil, err
	}
	recipient, err := sdk.AccAddressFromBech32(mint.RecipientAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient")
	}
	amount, ok := sdk.NewIntFromString(mint.Amount)
	if !ok {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, mint.Amount+mint.Denom)
	}
	token := sdk.Coin{Denom: mint.Denom, Amount: amount}
	if err := token.Validate(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, err.Error()), "mint tokens handler")
	}
	if err := h.minter.MintCoins(ctx, types.ModuleName, sdk.NewCoins(token)); err != nil {
		return nil, sdkerrors.Wrap(err, "mint")
	}

	if err := h.minter.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(token)); err != nil {
		return nil, sdkerrors.Wrap(err, "send to recipient")
	}

	return sdk.Events{sdk.NewEvent(
		types.EventTypeMintTokens,
		sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, token.String()),
		sdk.NewAttribute(types.AttributeKeyRecipient, mint.RecipientAddr),
	)}, nil
}

// assertHasPrivilege helper to assert that the contract has the required privilege
func (h TgradeHandler) assertHasPrivilege(ctx sdk.Context, contractAddr sdk.AccAddress, requiredPrivilege types.PrivilegeType) error {
	contractInfo := h.keeper.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return err
	}
	if !details.HasRegisteredPrivilege(requiredPrivilege) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "requires: %s", requiredPrivilege.String())
	}
	return nil
}
