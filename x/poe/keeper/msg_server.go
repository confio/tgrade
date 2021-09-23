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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PoEKeeper is a subset of the keeper
type PoEKeeper interface {
	ContractSource
	SetValidatorInitialEngagementPoints(ctx sdk.Context, address sdk.AccAddress, value sdk.Coin) error
	GetBondDenom(ctx sdk.Context) string
}

type msgServer struct {
	keeper         PoEKeeper
	contractKeeper wasmtypes.ContractOpsKeeper
	twasmKeeper    TwasmKeeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(poeKeeper PoEKeeper, contractKeeper wasmtypes.ContractOpsKeeper, twasmKeeper TwasmKeeper) types.MsgServer {
	return &msgServer{keeper: poeKeeper, contractKeeper: contractKeeper, twasmKeeper: twasmKeeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) CreateValidator(c context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

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

	valsetContractAddr, err := m.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "valset")
	}
	operatorAddress, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "operator address")
	}

	err = contract.RegisterValidator(ctx, valsetContractAddr, pk, operatorAddress, msg.Description, m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "register validator")
	}
	// delegate
	stakingContractAddr, err := m.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "staking contract")
	}

	err = contract.BondDelegation(ctx, stakingContractAddr, operatorAddress, sdk.NewCoins(msg.Value), m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "self delegation validator")
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.OperatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValOperator, msg.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyMoniker, msg.Description.Moniker),
			sdk.NewAttribute(types.AttributeKeyPubKeyHex, hex.EncodeToString(pk.Bytes())),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.Amount.String()),
		),
	})
	if err := m.keeper.SetValidatorInitialEngagementPoints(ctx, operatorAddress, msg.Value); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.MsgCreateValidatorResponse{}, nil
}

func (m msgServer) UpdateValidator(c context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, sdkerrors.Wrap(err, "description")
	}

	valsetContractAddr, err := m.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "valset")
	}
	operatorAddress, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "operator address")
	}

	// client sends a diff. we need to query the old description and merge it
	current, err := contract.QueryValidator(ctx, m.twasmKeeper, valsetContractAddr, operatorAddress)
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
	err = contract.UpdateValidator(ctx, valsetContractAddr, operatorAddress, newDescr, m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "update validator")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.OperatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeUpdateValidator,
			sdk.NewAttribute(types.AttributeKeyValOperator, msg.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyMoniker, msg.Description.Moniker),
		),
	})

	return &types.MsgUpdateValidatorResponse{}, nil
}

func (m msgServer) Delegate(c context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	operatorAddress, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "operator address")
	}

	stakingContractAddr, err := m.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "staking contract")
	}
	err = contract.BondDelegation(ctx, stakingContractAddr, operatorAddress, sdk.NewCoins(msg.Amount), m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "bond delegation")
	}
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.OperatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	})
	return &types.MsgDelegateResponse{}, nil
}

func (m msgServer) Undelegate(c context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	operatorAddress, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "operator address")
	}

	stakingContractAddr, err := m.keeper.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "staking contract")
	}
	if msg.Amount.Denom != m.keeper.GetBondDenom(ctx) {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "denom")
	}
	err = contract.UnbondDelegation(ctx, stakingContractAddr, operatorAddress, msg.Amount.Amount, m.contractKeeper)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "unbond delegation")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.OperatorAddress),
		),
		sdk.NewEvent(
			types.EventTypeUndelegate,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
	})
	return &types.MsgUndelegateResponse{}, nil

}
