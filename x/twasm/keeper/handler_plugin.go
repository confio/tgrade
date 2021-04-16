package keeper

import (
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/contract"
	"github.com/confio/tgrade/x/twasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type tgradeKeeper interface {
	IsPrivileged(ctx sdk.Context, contract sdk.AccAddress) bool
	appendToPrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, contractAddress sdk.AccAddress) uint8
	removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress) bool
	setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
}

var _ wasmkeeper.Messenger = TgradeHandler{}

type TgradeHandler struct {
	keeper tgradeKeeper
}

func NewTgradeHandler(keeper tgradeKeeper) *TgradeHandler {
	return &TgradeHandler{keeper: keeper}
}

func (h TgradeHandler) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
	if msg.Custom == nil {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	if !h.keeper.IsPrivileged(ctx, contractAddr) {
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	var tMsg contract.TgradeMsg
	if err := json.Unmarshal(msg.Custom, &tMsg); err != nil {
		return nil, nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	switch {
	case tMsg.Hooks != nil:
		return h.handleHooks(ctx, contractAddr, tMsg.Hooks)

	}
	return nil, nil, wasmtypes.ErrUnknownMsg
}

func (h TgradeHandler) handleHooks(ctx sdk.Context, contractAddr sdk.AccAddress, hooks *contract.Hooks) ([]sdk.Event, [][]byte, error) {
	contractInfo := h.keeper.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return nil, nil, sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return nil, nil, err
	}

	register := func(c types.PrivilegedCallbackType) error {
		if details.HasRegisteredContractCallback(c) {
			return nil
		}
		pos := h.keeper.appendToPrivilegedContractCallbacks(ctx, c, contractAddr)
		details.AddRegisteredCallback(c, pos)
		return h.keeper.setContractDetails(ctx, contractAddr, &details)
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
		return h.keeper.setContractDetails(ctx, contractAddr, &details)
	}
	var err error
	switch {
	case hooks.RegisterBeginBlock != nil:
		err = register(types.CallbackTypeBeginBlock)
	case hooks.UnregisterBeginBlock != nil:
		err = unregister(types.CallbackTypeBeginBlock)
	case hooks.RegisterEndBlock != nil:
		err = register(types.CallbackTypeEndBlock)
	case hooks.UnregisterEndBlock != nil:
		err = unregister(types.CallbackTypeEndBlock)
	case hooks.RegisterValidatorSetUpdate != nil:
		err = register(types.CallbackTypeValidatorSetUpdate)
	case hooks.UnregisterValidatorSetUpdate != nil:
		err = unregister(types.CallbackTypeValidatorSetUpdate)
	default:
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	return nil, nil, err
}
