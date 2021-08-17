package keeper

import (
	"context"
	"encoding/hex"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
)

// ContractSource is a subset of the keeper
type ContractSource interface {
	GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
}
type msgServer struct {
	contractSource ContractSource
	contractKeeper wasmtypes.ContractOpsKeeper
	smartQuerier   types.SmartQuerier
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(poeKeeper ContractSource, contractKeeper wasmtypes.ContractOpsKeeper, q types.SmartQuerier) types.MsgServer {
	return &msgServer{contractSource: poeKeeper, contractKeeper: contractKeeper, smartQuerier: q}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) CreateValidator(goCtx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		if !tmstrings.StringInSlice(pk.Type(), cp.Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				stakingtypes.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	contractAddr, err := m.contractSource.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "valset")
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "delegator address")
	}

	err = contract.RegisterValidator(ctx, contractAddr, pk, delegatorAddress, msg.Description, m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "register validator")
	}
	// delegate
	contractAddr, err = m.contractSource.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "staking")
	}

	err = contract.BondTokens(ctx, contractAddr, delegatorAddress, sdk.NewCoins(msg.Value), m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "self delegation validator")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValOperator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyMoniker, msg.Description.Moniker),
			sdk.NewAttribute(types.AttributeKeyPubKeyHex, hex.EncodeToString(pk.Bytes())),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.Amount.String()),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

func (m msgServer) UpdateValidator(goCtx context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, sdkerrors.Wrap(err, "description")
	}

	contractAddr, err := m.contractSource.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "valset")
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "delegator address")
	}

	// client sends a diff. we need to query the old description and merge it
	current, err := contract.QueryValidator(ctx, m.smartQuerier, contractAddr, delegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "query current description")
	}
	if current == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "operator")
	}
	validator, err := current.ToValidator()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "to validator")
	}
	newDescr, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "merge description")
	}
	// do the update
	err = contract.UpdateValidator(ctx, contractAddr, delegatorAddress, newDescr, m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "update validator")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeUpdateValidator,
			sdk.NewAttribute(types.AttributeKeyValOperator, msg.DelegatorAddress),
			sdk.NewAttribute(types.AttributeKeyMoniker, msg.Description.Moniker),
		),
	})

	return &types.MsgUpdateValidatorResponse{}, nil
}
