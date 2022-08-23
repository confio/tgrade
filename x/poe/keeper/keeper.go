package keeper

import (
	"fmt"
	"sync"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/confio/tgrade/x/poe/types"
)

type Keeper struct {
	codec             codec.Codec
	storeKey          sdk.StoreKey
	paramStore        paramtypes.Subspace
	twasmKeeper       types.TWasmKeeper
	contractAddrCache sync.Map
	validatorVotes    validatorVotes
}

type validatorVotes struct {
	ValidatorVotes []abcitypes.VoteInfo
	RwLock         sync.RWMutex
}

// NewKeeper constructor
func NewKeeper(
	marshaler codec.Codec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	twasmK types.TWasmKeeper,
	ak types.AuthKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	// ensure bonded and not bonded module accounts are set
	if addr := ak.GetModuleAddress(types.BondedPoolName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	}

	return Keeper{
		codec:             marshaler,
		storeKey:          key,
		paramStore:        paramSpace,
		twasmKeeper:       twasmK,
		contractAddrCache: sync.Map{},
	}
}

// SetPoEContractAddress stores the contract address for the given type. If one exists already then it is overwritten.
func (k *Keeper) SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(poeContractAddressKey(ctype), contractAddr.Bytes())
	k.contractAddrCache.Store(ctype, contractAddr)
}

// InitContractAddressCache adds all poe contracts to the in memory cache
func (k *Keeper) InitContractAddressCache(ctx sdk.Context) {
	k.IteratePoEContracts(ctx, func(contractType types.PoEContractType, address sdk.AccAddress) bool {
		k.contractAddrCache.Store(contractType, address)
		return false
	})
}

// GetPoEContractAddress get the stored contract address for the given type or returns an error when not exists (yet)
func (k *Keeper) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if err := ctype.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(err, "contract type")
	}

	// try to get addr from cache
	if cachedAddr, ok := k.contractAddrCache.Load(ctype); ok {
		if addr, ok := cachedAddr.(sdk.AccAddress); ok {
			return addr, nil
		}
	}

	// if not in cache, try to get addr from store
	store := ctx.KVStore(k.storeKey)
	addr := store.Get(poeContractAddressKey(ctype))
	if len(addr) == 0 {
		return nil, wasmtypes.ErrNotFound
	}

	return addr, nil
}

// IteratePoEContracts for each persisted PoE contract the given callback is called.
// When the callback returns true, the loop is aborted early.
func (k *Keeper) IteratePoEContracts(ctx sdk.Context, cb func(types.PoEContractType, sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		ctype := types.PoEContractTypeFrom(string(iter.Key()))
		if cb(ctype, iter.Value()) {
			return
		}
	}
}

// UnbondingTime returns the unbonding period from the staking contract
func (k *Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	rsp, err := k.StakeContract(ctx).QueryStakingUnbondingPeriod(ctx)
	if err != nil {
		panic(fmt.Sprintf("unboding period: %s", err))
	}
	return rsp
}

func (k *Keeper) GetBondDenom(ctx sdk.Context) string {
	return types.DefaultBondDenom
}

func (k *Keeper) UpdateValidatorVotes(validatorVotes []abcitypes.VoteInfo) {
	k.validatorVotes.RwLock.Lock()
	k.validatorVotes.ValidatorVotes = validatorVotes
	k.validatorVotes.RwLock.Unlock()
}

func (k *Keeper) GetValidatorVotes() []abcitypes.VoteInfo {
	k.validatorVotes.RwLock.RLock()
	defer k.validatorVotes.RwLock.RUnlock()
	return k.validatorVotes.ValidatorVotes
}

func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func poeContractAddressKey(ctype types.PoEContractType) []byte {
	return append(types.ContractPrefix, []byte(ctype.String())...)
}
