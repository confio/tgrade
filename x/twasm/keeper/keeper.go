package keeper

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	govRouter      govtypes.Router
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
		govRouter:  govRouter,
	}
	// configure wasm keeper via options

	handlerChainOpt := wasmkeeper.WithMessageHandlerDecorator(func(vanillaMessenger wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			// disable staking messages
			wasmkeeper.MessageHandlerFunc(func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
				if msg.Staking != nil {
					return nil, nil, sdkerrors.Wrap(wasmtypes.ErrExecuteFailed, "not supported, yet")
				}
				return nil, nil, wasmtypes.ErrUnknownMsg
			}),
			vanillaMessenger,
			// append our custom message handler
			NewTgradeHandler(cdc, &result, bankKeeper, govRouter),
		)
	})
	queryChainOpt := wasmkeeper.WithQueryHandlerDecorator(func(old wasmkeeper.WasmVMQueryHandler) wasmkeeper.WasmVMQueryHandler {
		return wasmkeeper.WasmVMQueryHandlerFn(func(ctx sdk.Context, caller sdk.AccAddress, request wasmvmtypes.QueryRequest) ([]byte, error) {
			if request.Staking != nil {
				return nil, wasmvmtypes.UnsupportedRequest{Kind: "not supported, yet"}
			}
			return old.HandleQuery(ctx, caller, request)
		})
	})
	opts = append([]wasm.Option{
		handlerChainOpt,
		queryChainOpt,
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

// setContractDetails stores new tgrade data with the contract info.
func (k Keeper) setContractDetails(ctx sdk.Context, contract sdk.AccAddress, details *types.TgradeContractDetails) error {
	return k.contractKeeper.SetContractInfoExtension(ctx, contract, details)
}

// getContractDetails loads tgrade details. This method should only be used when no ContractInfo is used anywhere.
func (k Keeper) getContractDetails(ctx sdk.Context, contract sdk.AccAddress) (*types.TgradeContractDetails, error) {
	contractInfo := k.GetContractInfo(ctx, contract)
	if contractInfo == nil {
		return nil, sdkerrors.Wrap(wasmtypes.ErrNotFound, "contract info")
	}

	var details types.TgradeContractDetails
	if err := contractInfo.ReadExtension(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

// GetContractKeeper returns the contract keeper instance with default permissions set
func (k *Keeper) GetContractKeeper() wasmtypes.ContractOpsKeeper {
	return k.contractKeeper
}
