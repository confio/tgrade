package keeper

import (
	"fmt"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/types"
)

type TwasmKeeper interface {
	types.SmartQuerier
	types.Sudoer
}

type Keeper struct {
	marshaler   codec.Marshaler
	storeKey    sdk.StoreKey
	paramStore  paramtypes.Subspace
	twasmKeeper TwasmKeeper
}

// NewKeeper constructor
func NewKeeper(marshaler codec.Marshaler, key sdk.StoreKey, paramSpace paramtypes.Subspace, twasmK TwasmKeeper) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}
	return Keeper{
		marshaler:   marshaler,
		storeKey:    key,
		paramStore:  paramSpace,
		twasmKeeper: twasmK,
	}
}

func (k Keeper) SetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType, contractAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(poeContractAddressKey(ctype), contractAddr.Bytes())
}

func (k Keeper) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if ctype == types.PoEContractTypeUndefined {
		return nil, sdkerrors.Wrap(wasmtypes.ErrInvalid, "contract type")
	}
	if _, ok := types.PoEContractType_name[int32(ctype)]; !ok {
		return nil, sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract type")
	}
	store := ctx.KVStore(k.storeKey)
	addr := store.Get(poeContractAddressKey(ctype))
	if len(addr) == 0 {
		return nil, wasmtypes.ErrNotFound
	}
	return addr, nil
}

func (k Keeper) IteratePoEContracts(ctx sdk.Context, cb func(types.PoEContractType, sdk.AccAddress) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractPrefix)
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		ctype := types.PoEContractType_value[string(iter.Key())]
		if cb(types.PoEContractType(ctype), iter.Value()) {
			return
		}
	}
}

func (k Keeper) setPoESystemAdminAddress(ctx sdk.Context, admin sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.SystemAdminPrefix, admin.Bytes())
}

func (k Keeper) GetPoESystemAdminAddress(ctx sdk.Context) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	return store.Get(types.SystemAdminPrefix)
}

// UnbondingTime returns the unbonding period from the staking contract
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	contractAddr, err := k.GetPoEContractAddress(ctx, types.PoEContractTypeStaking)
	if err != nil {
		panic(fmt.Sprintf("get contract address: %s", err))
	}

	rsp, err := contract.QueryStakingUnbondingPeriod(ctx, k.twasmKeeper, contractAddr)
	if err != nil {
		panic(fmt.Sprintf("query unponding period from %s: %s", contractAddr.String(), err))
	}
	return time.Duration(rsp) * time.Second
}

func (k Keeper) GetBondDenom(ctx sdk.Context) string {
	return types.DefaultBondDenom
}

func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func poeContractAddressKey(ctype types.PoEContractType) []byte {
	return append(types.ContractPrefix, []byte(ctype.String())...)
}
