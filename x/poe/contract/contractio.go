package contract

import (
	"encoding/json"
	"github.com/confio/tgrade/x/poe/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// RegisterValidator calls valset contract to register a new validator key and address
func RegisterValidator(ctx sdk.Context, contractAddr sdk.AccAddress, pk cryptotypes.PubKey, delegatorAddress sdk.AccAddress, description stakingtypes.Description, k types.Executor) error {
	pub, err := NewValidatorPubkey(pk)
	if err != nil {
		return err
	}
	registerValidator := TG4ValsetExecute{
		RegisterValidatorKey: &RegisterValidatorKey{
			PubKey:   pub,
			Metadata: MetadataFromDescription(description),
		},
	}
	payloadBz, err := json.Marshal(&registerValidator)
	if err != nil {
		return sdkerrors.Wrap(err, "serialize payload msg")
	}

	_, err = k.Execute(ctx, contractAddr, delegatorAddress, payloadBz, nil)
	return sdkerrors.Wrap(err, "execute contract")
}

// UpdateValidator calls valset contract to change validator's metadata
func UpdateValidator(ctx sdk.Context, contractAddr sdk.AccAddress, delegatorAddress sdk.AccAddress, description stakingtypes.Description, k types.Executor) error {
	metadata := MetadataFromDescription(description)
	updateValidator := TG4ValsetExecute{
		UpdateMetadata: &metadata,
	}
	payloadBz, err := json.Marshal(&updateValidator)
	if err != nil {
		return sdkerrors.Wrap(err, "serialize payload msg")
	}

	_, err = k.Execute(ctx, contractAddr, delegatorAddress, payloadBz, nil)
	return sdkerrors.Wrap(err, "execute contract")
}

// CallEndBlockWithValidatorUpdate calls valset contract for a validator diff
func CallEndBlockWithValidatorUpdate(ctx sdk.Context, contractAddr sdk.AccAddress, k types.Sudoer) ([]abci.ValidatorUpdate, error) {
	sudoMsg := ValidatorUpdateSudoMsg{EndWithValidatorUpdate: &struct{}{}}
	msgBz, err := json.Marshal(sudoMsg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "tgrade sudo msg")
	}

	resp, err := k.Sudo(ctx, contractAddr, msgBz)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sudo")
	}
	if len(resp) == 0 {
		return nil, nil
	}
	var contractResult EndWithValidatorUpdateResponse
	if err := json.Unmarshal(resp, &contractResult); err != nil {
		return nil, sdkerrors.Wrap(err, "contract response")
	}
	if len(contractResult.Diffs) == 0 {
		return nil, nil
	}

	result := make([]abci.ValidatorUpdate, len(contractResult.Diffs))
	for i, v := range contractResult.Diffs {
		pub, err := convertToTendermintPubKey(v.PubKey)
		if err != nil {
			return nil, err
		}
		result[i] = abci.ValidatorUpdate{
			PubKey: pub,
			Power:  int64(v.Power),
		}
	}
	return result, nil
}

// SetEngagementPoints set engagement points  If the member already exists, its weight will be reset to the weight sent here
func SetEngagementPoints(ctx sdk.Context, contractAddr sdk.AccAddress, k types.Sudoer, opAddr sdk.AccAddress, points uint64) error {
	msg := TG4GroupSudoMsg{
		UpdateMember: &TG4GroupSudoUpdateMember{Member: TG4Member{Addr: opAddr.String(), Weight: points}},
	}
	msgBz, err := json.Marshal(msg)
	if err != nil {
		return sdkerrors.Wrap(err, "tg4 group sudo msg")
	}

	_, err = k.Sudo(ctx, contractAddr, msgBz)
	return sdkerrors.Wrap(err, "sudo")
}

func convertToTendermintPubKey(key ValidatorPubkey) (crypto.PublicKey, error) {
	switch {
	case key.Ed25519 != nil:
		return crypto.PublicKey{
			Sum: &crypto.PublicKey_Ed25519{
				Ed25519: key.Ed25519,
			},
		}, nil
	case key.Secp256k1 != nil:
		return crypto.PublicKey{
			Sum: &crypto.PublicKey_Secp256K1{
				Secp256K1: key.Secp256k1,
			},
		}, nil
	default:
		return crypto.PublicKey{}, types.ErrValidatorPubKeyTypeNotSupported
	}
}

// BondTokens sends given amount to the staking contract to increase the bonded amount for the delegator
func BondTokens(ctx sdk.Context, contractAddr sdk.AccAddress, delegatorAddress sdk.AccAddress, amount sdk.Coins, k types.Executor) error {
	bondStake := TG4StakeExecute{
		Bond: &struct{}{},
	}
	payloadBz, err := json.Marshal(&bondStake)
	if err != nil {
		return sdkerrors.Wrap(err, "serialize payload msg")
	}

	_, err = k.Execute(ctx, contractAddr, delegatorAddress, payloadBz, amount)
	return sdkerrors.Wrap(err, "execute contract")
}
