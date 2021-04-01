package keeper

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	wasmkeeper.Keeper
	cdc            codec.Marshaler
	storeKey       sdk.StoreKey
	contractKeeper wasmtypes.ContractOpsKeeper
	paramSpace     paramtypes.Subspace
}

func NewKeeper(
	cdc codec.Marshaler,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper wasmtypes.StakingKeeper,
	distKeeper wasmtypes.DistributionKeeper,
	channelKeeper wasmtypes.ChannelKeeper,
	portKeeper wasmtypes.PortKeeper,
	capabilityKeeper wasmtypes.CapabilityKeeper,
	portSource wasmtypes.ICS20TransferPortSource,
	router sdk.Router,
	queryRouter wasmkeeper.GRPCQueryRouter,
	govRouter govtypes.Router,
	homeDir string,
	twasmConfig types.TWasmConfig,
	supportedFeatures string,
	opts ...wasmkeeper.Option,
) Keeper {
	result := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramSpace: paramSpace,
	}
	// configure wasm keeper via options

	var handlerChain wasmkeeper.Messenger = wasmkeeper.NewMessageHandlerChain(
		wasmkeeper.NewDefaultMessageHandler(
			router,
			channelKeeper,
			capabilityKeeper,
			cdc,
			portSource,
		),
	)
	var queryPlugins wasmkeeper.WASMVMQueryHandler = wasmkeeper.DefaultQueryPlugins(bankKeeper, stakingKeeper, distKeeper, channelKeeper, queryRouter, &result.Keeper)

	opts = append([]wasm.Option{
		wasmkeeper.WithMessageHandler(handlerChain),
		wasmkeeper.WithQueryHandler(queryPlugins),
	}, opts...)

	result.Keeper = wasmkeeper.NewKeeper(
		cdc,
		storeKey,
		paramSpace,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distKeeper,
		channelKeeper,
		portKeeper,
		capabilityKeeper,
		portSource,
		router,
		queryRouter,
		homeDir,
		twasmConfig.WasmConfig,
		supportedFeatures,
		opts...,
	)
	result.contractKeeper = wasmkeeper.NewDefaultPermissionKeeper(&result.Keeper)
	return result
}

func (k Keeper) setParams(ctx sdk.Context, ps wasmtypes.Params) {
	k.paramSpace.SetParamSet(ctx, &ps)
}

func WasmQuerier(k *Keeper) wasmtypes.QueryServer {
	return wasmkeeper.NewGrpcQuerier(k.cdc, k.storeKey, k, k.QueryGasLimit())
}

func (Keeper) Logger(ctx sdk.Context) log.Logger {
	return ModuleLogger(ctx)
}
func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
