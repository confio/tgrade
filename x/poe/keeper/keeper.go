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

	"github.com/confio/tgrade/x/poe/types"
)

type Keeper struct {
	codec       codec.Codec
	storeKey    sdk.StoreKey
	paramStore  paramtypes.Subspace
	twasmKeeper types.TWasmKeeper
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
		codec:       marshaler,
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

// UnbondingTime returns the unbonding period from the staking contract
func (k Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	rsp, err := k.StakeContract(ctx).QueryStakingUnbondingPeriod(ctx)
	if err != nil {
		panic(fmt.Sprintf("unboding period: %s", err))
	}
	return rsp
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
