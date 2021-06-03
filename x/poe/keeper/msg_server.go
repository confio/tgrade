package keeper

import (
	"context"
	"encoding/json"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	poecontract "github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm/contract"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
)

type msgServer struct {
	Keeper
	contractKeeper wasmtypes.ContractOpsKeeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, contractKeeper wasmtypes.ContractOpsKeeper) types.MsgServer {
	return &msgServer{Keeper: keeper, contractKeeper: contractKeeper}
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

	// register validator
	if pk.Type() != "ed25519" { // todo (Alex): revisit
		return nil, sdkerrors.Wrap(wasmtypes.ErrInvalid, "only ed25519 supported currently")
	}
	registerValidator := poecontract.TG4ValsetExecute{
		RegisterValidatorKey: poecontract.RegisterValidatorKey{
			PubKey: contract.ValidatorPubkey{
				Ed25519: pk.Bytes(),
			},
		},
	}
	payloadBz, err := json.Marshal(&registerValidator)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "serialize payload msg")
	}

	contractAddr, err := m.GetPoEContractAddress(ctx, types.PoEContractTypes_VALSET)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "valset")
	}
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "delegator address")
	}
	_, err = m.contractKeeper.Execute(ctx, contractAddr, delegatorAddress, payloadBz, nil)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "register validator")
	}
	// delegate
	initialStake := poecontract.TG4StakeExecute{
		Bond: &struct{}{},
	}
	payloadBz, err = json.Marshal(&initialStake)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "serialize payload msg")
	}
	contractAddr, err = m.GetPoEContractAddress(ctx, types.PoEContractTypes_STAKING)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "staking")
	}

	_, err = m.contractKeeper.Execute(ctx, contractAddr, delegatorAddress, payloadBz, sdk.NewCoins(msg.Value))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "self delegation validator")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}
