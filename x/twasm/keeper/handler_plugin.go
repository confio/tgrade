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
	removePrivilegedContractCallbacks(ctx sdk.Context, callbackType types.PrivilegedCallbackType, pos uint8, contractAddr sdk.AccAddress)
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
	details, err := types.ContractDetails(*contractInfo)
	if err != nil {
		return nil, nil, err
	}

	register := func(tp types.PrivilegedCallbackType) {
		if details.HasRegisteredContractCallback(tp) {
			return
		}
		pos := h.keeper.appendToPrivilegedContractCallbacks(ctx, tp, contractAddr)
		details.AddRegisteredCallback(tp, pos)
		h.keeper.setContractDetails(ctx, contractAddr, details)
	}
	unregister := func(tp types.PrivilegedCallbackType) {
		if !details.HasRegisteredContractCallback(tp) {
			return
		}
		details.IterateRegisteredCallbacks(func(t types.PrivilegedCallbackType, pos uint8) bool {
			if t != tp {
				return false
			}
			h.keeper.removePrivilegedContractCallbacks(ctx, t, pos, contractAddr)
			details.RemoveRegisteredCallback(t, pos)
			return false
		})
		h.keeper.setContractDetails(ctx, contractAddr, details)
	}
	switch {
	case hooks.RegisterBeginBlock != nil:
		register(types.CallbackTypeBeginBlock)
	case hooks.UnregisterBeginBlock != nil:
		unregister(types.CallbackTypeBeginBlock)
	case hooks.RegisterEndBlock != nil:
		register(types.CallbackTypeEndBlock)
	case hooks.UnregisterEndBlock != nil:
		unregister(types.CallbackTypeEndBlock)
	case hooks.RegisterValidatorSetUpdate != nil:
		register(types.CallbackTypeValidatorSetUpdate)
	case hooks.UnregisterValidatorSetUpdate != nil:
		unregister(types.CallbackTypeValidatorSetUpdate)
	default:
		return nil, nil, wasmtypes.ErrUnknownMsg
	}
	return nil, nil, nil
}
