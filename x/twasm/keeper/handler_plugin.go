package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	abci "github.com/tendermint/tendermint/abci/types"

	poetypes "github.com/confio/tgrade/x/poe/types"
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

// bankKeeper is a subset of the SDK bank keeper
type bankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// ConsensusParamsUpdater is a subset of baseapp to store the consensus params
type ConsensusParamsUpdater interface {
	GetConsensusParams(ctx sdk.Context) *abci.ConsensusParams
	StoreConsensusParams(ctx sdk.Context, cp *abci.ConsensusParams)
}

var _ wasmkeeper.Messenger = TgradeHandler{}

// TgradeHandler is a custom message handler plugin for wasmd.
type TgradeHandler struct {
	cdc                    codec.Codec
	keeper                 TgradeWasmHandlerKeeper
	bankKeeper             bankKeeper
	govRouter              govtypes.Router
	consensusParamsUpdater ConsensusParamsUpdater
}

// NewTgradeHandler constructor
func NewTgradeHandler(
	cdc codec.Codec,
	keeper TgradeWasmHandlerKeeper,
	bankKeeper bankKeeper,
	consensusParamsUpdater ConsensusParamsUpdater,
	govRouter govtypes.Router,
) *TgradeHandler {
	return &TgradeHandler{
		cdc:                    cdc,
		keeper:                 keeper,
		govRouter:              restrictParamsDecorator(govRouter),
		bankKeeper:             bankKeeper,
		consensusParamsUpdater: consensusParamsUpdater,
	}
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
	// main message dispatcher
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
	case tMsg.ConsensusParams != nil:
		evts, err := h.handleConsensusParamsUpdate(ctx, contractAddr, tMsg.ConsensusParams)
		return append(evts, em.Events()...), nil, err
	case tMsg.Delegate != nil:
		evts, err := h.handleDelegate(ctx, contractAddr, tMsg.Delegate)
		return append(evts, em.Events()...), nil, err
	case tMsg.Undelegate != nil:
		evts, err := h.handleUndelegate(ctx, contractAddr, tMsg.Undelegate)
		return append(evts, em.Events()...), nil, err
	}

	return nil, nil, sdkerrors.Wrapf(wasmtypes.ErrUnknownMsg, "unknown type: %T", msg)
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

	content := exec.GetProposalContent(contractAddr)
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
	if err := h.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(token)); err != nil {
		return nil, sdkerrors.Wrap(err, "mint")
	}

	if err := h.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(token)); err != nil {
		return nil, sdkerrors.Wrap(err, "send to recipient")
	}

	return sdk.Events{sdk.NewEvent(
		types.EventTypeMintTokens,
		sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, token.String()),
		sdk.NewAttribute(types.AttributeKeyRecipient, mint.RecipientAddr),
	)}, nil
}

// handle the consensus parameters update message
func (h TgradeHandler) handleConsensusParamsUpdate(ctx sdk.Context, contractAddr sdk.AccAddress, pUpdate *contract.ConsensusParamsUpdate) ([]sdk.Event, error) {
	if err := h.assertHasPrivilege(ctx, contractAddr, types.PrivilegeConsensusParamChanger); err != nil {
		return nil, err
	}
	if err := pUpdate.ValidateBasic(); err != nil {
		return nil, err
	}
	params := h.consensusParamsUpdater.GetConsensusParams(ctx)
	h.consensusParamsUpdater.StoreConsensusParams(ctx, mergeConsensusParamsUpdate(params, pUpdate))
	return nil, nil
}

func mergeConsensusParamsUpdate(src *abci.ConsensusParams, delta *contract.ConsensusParamsUpdate) *abci.ConsensusParams {
	if delta.Block != nil {
		if delta.Block.MaxBytes != nil {
			src.Block.MaxBytes = *delta.Block.MaxBytes
		}
		if delta.Block.MaxGas != nil {
			src.Block.MaxGas = *delta.Block.MaxGas
		}
	}
	if delta.Evidence != nil {
		if delta.Evidence.MaxAgeNumBlocks != nil {
			src.Evidence.MaxAgeNumBlocks = *delta.Evidence.MaxAgeNumBlocks
		}
		if delta.Evidence.MaxAgeDuration != nil {
			src.Evidence.MaxAgeDuration = time.Duration(*delta.Evidence.MaxAgeDuration) * time.Second
		}
		if delta.Evidence.MaxBytes != nil {
			src.Evidence.MaxBytes = *delta.Evidence.MaxBytes
		}
	}
	return src
}

// handle delegate token message
func (h TgradeHandler) handleDelegate(ctx sdk.Context, contractAddr sdk.AccAddress, delegate *contract.Delegate) ([]sdk.Event, error) {
	if err := h.assertHasPrivilege(ctx, contractAddr, types.PrivilegeDelegator); err != nil {
		return nil, err
	}
	fromAddr, err := sdk.AccAddressFromBech32(delegate.StakerAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "fromAddr")
	}
	amt, err := wasmkeeper.ConvertWasmCoinsToSdkCoins(wasmvmtypes.Coins{delegate.Funds})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	if err := h.bankKeeper.DelegateCoinsFromAccountToModule(ctx, fromAddr, poetypes.BondedPoolName, amt); err != nil {
		return nil, sdkerrors.Wrap(err, "delegate")
	}
	if err := h.bankKeeper.SendCoinsFromModuleToAccount(ctx, poetypes.BondedPoolName, contractAddr, amt); err != nil {
		return nil, sdkerrors.Wrap(err, "module to contract")
	}

	return sdk.Events{sdk.NewEvent(
		types.EventTypeDelegateTokens,
		sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		sdk.NewAttribute(types.AttributeKeySender, delegate.StakerAddr),
	)}, nil
}

// handle undelegate token message
func (h TgradeHandler) handleUndelegate(ctx sdk.Context, contractAddr sdk.AccAddress, undelegate *contract.Undelegate) ([]sdk.Event, error) {
	if err := h.assertHasPrivilege(ctx, contractAddr, types.PrivilegeDelegator); err != nil {
		return nil, err
	}
	recipient, err := sdk.AccAddressFromBech32(undelegate.RecipientAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient")
	}
	amt, err := wasmkeeper.ConvertWasmCoinsToSdkCoins(wasmvmtypes.Coins{undelegate.Funds})
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}
	if err := h.bankKeeper.SendCoinsFromAccountToModule(ctx, contractAddr, poetypes.BondedPoolName, amt); err != nil {
		return nil, sdkerrors.Wrap(err, "contract to module")
	}
	if err := h.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, poetypes.BondedPoolName, recipient, amt); err != nil {
		return nil, sdkerrors.Wrap(err, "undelegate")
	}

	return sdk.Events{sdk.NewEvent(
		types.EventTypeUndelegateTokens,
		sdk.NewAttribute(wasmtypes.AttributeKeyContractAddr, contractAddr.String()),
		sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		sdk.NewAttribute(types.AttributeKeyRecipient, undelegate.RecipientAddr),
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

var _ govtypes.Router = restrictedParamsRouter{}

// decorator that prevents updates in baseapp subspace
type restrictedParamsRouter struct {
	nested govtypes.Router
}

// decorate router to prevent consensus updates via param proposal
func restrictParamsDecorator(router govtypes.Router) restrictedParamsRouter {
	return restrictedParamsRouter{nested: router}
}

func (d restrictedParamsRouter) HasRoute(r string) bool {
	return d.nested.HasRoute(r)
}

func (d restrictedParamsRouter) GetRoute(path string) (h govtypes.Handler) {
	r := d.nested.GetRoute(path)
	if path == paramproposal.RouterKey {
		return func(ctx sdk.Context, content govtypes.Content) error {
			if p, ok := content.(*paramproposal.ParameterChangeProposal); ok {
				// prevent updates in baseapp subspace
				for _, c := range p.Changes {
					if c.Subspace == baseapp.Paramspace {
						return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "base params can not be modified via params proposal")
					}
				}
			}
			return r(ctx, content)
		}
	}
	return r
}

func (d restrictedParamsRouter) AddRoute(r string, h govtypes.Handler) (rtr govtypes.Router) {
	panic("not supported")
}

func (d restrictedParamsRouter) Seal() {
	panic("not supported")
}
