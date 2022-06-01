package contract

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/confio/tgrade/x/poe/types"
)

// RegisterValidator calls valset contract to register a new validator key and address
func RegisterValidator(ctx sdk.Context, contractAddr sdk.AccAddress, pk cryptotypes.PubKey, operatorAddress sdk.AccAddress, description stakingtypes.Description, k types.Executor) error {
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

	_, err = k.Execute(ctx, contractAddr, operatorAddress, payloadBz, nil)
	return sdkerrors.Wrap(err, "execute contract")
}

// UpdateValidator calls valset contract to change validator's metadata
func UpdateValidator(ctx sdk.Context, contractAddr sdk.AccAddress, operatorAddress sdk.AccAddress, description stakingtypes.Description, k types.Executor) error {
	metadata := MetadataFromDescription(description)
	updateValidator := TG4ValsetExecute{
		UpdateMetadata: &metadata,
	}
	payloadBz, err := json.Marshal(&updateValidator)
	if err != nil {
		return sdkerrors.Wrap(err, "serialize payload msg")
	}

	_, err = k.Execute(ctx, contractAddr, operatorAddress, payloadBz, nil)
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
		pub, err := ConvertToTendermintPubKey(v.PubKey)
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

// UnbondDelegation unbond the given amount from the operators self delegation
// Amount must be in bonding token denom
func UnbondDelegation(ctx sdk.Context, contractAddr sdk.AccAddress, operatorAddress sdk.AccAddress, amount sdk.Coin, k types.Executor) (*time.Time, error) {
	if amount.Amount.IsNil() || amount.IsZero() || amount.IsNegative() || !amount.Amount.IsInt64() {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "amount")
	}
	msg := TG4StakeExecute{Unbond: &Unbond{Tokens: wasmvmtypes.NewCoin(amount.Amount.Uint64(), amount.Denom)}}
	msgBz, err := json.Marshal(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "TG4StakeExecute message")
	}
	// execute with a custom event manager so that captured events can be parsed for payload
	em := sdk.EventManager{}
	defer func() {
		ctx.EventManager().EmitEvents(em.Events())
	}()

	if _, err = k.Execute(ctx.WithEventManager(&em), contractAddr, operatorAddress, msgBz, nil); err != nil {
		return nil, sdkerrors.Wrap(err, "execute staking contract")
	}
	// parse events for unbound completion time
	for _, e := range em.Events() {
		if e.Type != wasmtypes.WasmModuleEventType {
			continue
		}
		var trusted bool
		for _, a := range e.Attributes {
			if string(a.Key) != wasmtypes.AttributeKeyContractAddr || string(a.Value) != contractAddr.String() {
				continue
			}
			trusted = true
			break
		}
		if !trusted { //  filter out other events
			continue
		}
		for _, a := range e.Attributes {
			if string(a.Key) != "completion_time" {
				continue
			}
			nanos, err := strconv.ParseInt(string(a.Value), 10, 64)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "completion time value")
			}
			completionTime := time.Unix(0, nanos).UTC()
			return &completionTime, nil
		}

	}
	return nil, types.ErrInvalid.Wrap("completion_time event attribute")
}

// BondDelegation sends given amounts to the staking contract to increase the bonded amount for the validator operator
func BondDelegation(ctx sdk.Context, contractAddr sdk.AccAddress, operatorAddress sdk.AccAddress, amount sdk.Coins, vestingAmount *sdk.Coin, k types.Executor) error {
	var vestingTokens *wasmvmtypes.Coin
	if vestingAmount != nil {
		vestingTokens = &wasmvmtypes.Coin{Amount: vestingAmount.Amount.String(), Denom: vestingAmount.Denom}
	}
	bondStake := TG4StakeExecute{
		Bond: &Bond{VestingTokens: vestingTokens},
	}
	payloadBz, err := json.Marshal(&bondStake)
	if err != nil {
		return sdkerrors.Wrap(err, "serialize payload msg")
	}

	_, err = k.Execute(ctx, contractAddr, operatorAddress, payloadBz, amount)
	return sdkerrors.Wrap(err, "execute contract")
}

// SetEngagementPoints set engagement points  If the member already exists, its weight will be reset to the weight sent here
func SetEngagementPoints(ctx sdk.Context, contractAddr sdk.AccAddress, k types.Sudoer, opAddr sdk.AccAddress, points uint64) error {
	msg := TG4EngagementSudoMsg{
		UpdateMember: &TG4Member{Addr: opAddr.String(), Points: points},
	}
	msgBz, err := json.Marshal(msg)
	if err != nil {
		return sdkerrors.Wrap(err, "tg4 group sudo msg")
	}

	_, err = k.Sudo(ctx, contractAddr, msgBz)
	return sdkerrors.Wrap(err, "sudo")
}

func ConvertToTendermintPubKey(key ValidatorPubkey) (crypto.PublicKey, error) {
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

// BaseContractAdapter is the base contract adapter type that contains common methods to interact with the contract
type BaseContractAdapter struct {
	contractAddr     sdk.AccAddress
	twasmKeeper      types.TWasmKeeper
	addressLookupErr error
}

// NewBaseContractAdapter constructor
func NewBaseContractAdapter(contractAddr sdk.AccAddress, twasmKeeper types.TWasmKeeper, addressLookupErr error) BaseContractAdapter {
	return BaseContractAdapter{contractAddr: contractAddr, twasmKeeper: twasmKeeper, addressLookupErr: addressLookupErr}
}

// send message via execute entry point
func (a BaseContractAdapter) doExecute(ctx sdk.Context, msg interface{}, sender sdk.AccAddress, coin ...sdk.Coin) error {
	if err := a.addressLookupErr; err != nil {
		return err
	}
	msgBz, err := json.Marshal(msg)
	if err != nil {
		return sdkerrors.Wrap(err, "encode execute msg")
	}
	_, err = a.twasmKeeper.GetContractKeeper().Execute(ctx, a.contractAddr, sender, msgBz, coin)
	return sdkerrors.Wrap(err, "execute")
}

// PageableResult is a query response where the cursor is a subset of the raw last element.
type PageableResult interface {
	// PaginationCursor pagination cursor
	PaginationCursor(raw []byte) (PaginationCursor, error)
}

// execute a smart query with the contract that returns multiple elements
// returns a cursor whenever the result set has more than 1 element
func (a BaseContractAdapter) doPageableQuery(ctx sdk.Context, query interface{}, result interface{}) (PaginationCursor, error) {
	if err := a.addressLookupErr; err != nil {
		return nil, err
	}
	bz, err := json.Marshal(query)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "marshal query")
	}
	res, err := a.twasmKeeper.QuerySmart(ctx, a.contractAddr, bz)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(res, result); err != nil {
		return nil, sdkerrors.Wrap(err, "unmarshal result")
	}
	if p, ok := result.(PageableResult); ok {
		return p.PaginationCursor(res)
	}
	// no pagination support (in tgrade, yet)
	return nil, nil
}

// execute a smart query with the contract
func (a BaseContractAdapter) doQuery(ctx sdk.Context, query interface{}, result interface{}) error {
	if err := a.addressLookupErr; err != nil {
		return err
	}
	bz, err := json.Marshal(query)
	if err != nil {
		return err
	}
	res, err := a.twasmKeeper.QuerySmart(ctx, a.contractAddr, bz)
	if err != nil {
		return err
	}
	return json.Unmarshal(res, result)
}

// Address returns contract address
func (a BaseContractAdapter) Address() (sdk.AccAddress, error) {
	return a.contractAddr, a.addressLookupErr
}

// PaginationCursor is the contracts "last element" as raw data that can be used to navigate through the result set.
type PaginationCursor []byte

// Empty is nil or zero size
func (p PaginationCursor) Empty() bool {
	return len(p) == 0
}

// String convert to string representation
func (p PaginationCursor) String() string {
	if p.Empty() {
		return ""
	}
	return string(p)
}

func (p PaginationCursor) Equal(o PaginationCursor) bool {
	return bytes.Equal(p, o)
}

type Paginator struct {
	StartAfter PaginationCursor `json:"start_after,omitempty"`
	Limit      uint64           `json:"limit,omitempty"`
}

// ToQuery converts to poe contract query format
func (p *Paginator) ToQuery() (string, int) {
	if p == nil {
		return "", 0
	}
	return string(p.StartAfter), int(p.Limit)
}

// NewPaginator constructor
func NewPaginator(pag *query.PageRequest) (*Paginator, error) {
	if pag == nil {
		return nil, nil
	}
	if pag.Offset != 0 {
		return nil, status.Error(codes.InvalidArgument, "pagination offset not supported")
	}
	if pag.CountTotal {
		return nil, status.Error(codes.InvalidArgument, "pagination count total not supported")
	}
	return &Paginator{
		StartAfter: pag.Key,
		Limit:      pag.Limit,
	}, nil
}
