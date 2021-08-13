package poe

import (
	"context"
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/client/cli"
	"github.com/confio/tgrade/x/poe/contract"
	"github.com/confio/tgrade/x/poe/keeper"
	"github.com/confio/tgrade/x/poe/types"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"sync"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the genutil module.
type AppModuleBasic struct {
}

// Name returns the genutil module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the genutil module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
	slashingtypes.RegisterInterfaces(registry)
	stakingtypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the genutil
// module.
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

// ValidateGenesis performs genesis state validation for the genutil module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, txEncodingConfig client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	// todo: add PoE validation

	return types.ValidateGenesis(data, txEncodingConfig.TxJSONDecoder())
}

// RegisterRESTRoutes registers the REST routes for the genutil module.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the genutil module.
func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns no root tx command for the genutil module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return cli.NewTxCmd() }

// GetQueryCmd returns no root query command for the genutil module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

//____________________________________________________________________________

// AppModule implements an application module for the genutil module.
type AppModule struct {
	AppModuleBasic
	deliverTx        DeliverTxfn
	txEncodingConfig client.TxEncodingConfig
	twasmKeeper      twasmKeeper
	contractKeeper   wasmtypes.ContractOpsKeeper
	poeKeeper        keeper.Keeper
	doOnce           sync.Once
}

// twasmKeeper subset of keeper to decouple from twasm module
type twasmKeeper interface {
	abciKeeper
	types.SmartQuerier
	types.Sudoer
	SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error
	HasPrivilegedContract(ctx sdk.Context, contractAddr sdk.AccAddress, privilegeType twasmtypes.PrivilegeType) (bool, error)
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	poeKeeper keeper.Keeper,
	twasmKeeper twasmKeeper,
	deliverTx DeliverTxfn,
	txEncodingConfig client.TxEncodingConfig,
	contractKeeper wasmtypes.ContractOpsKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic:   AppModuleBasic{},
		twasmKeeper:      twasmKeeper,
		contractKeeper:   contractKeeper,
		poeKeeper:        poeKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
	}
}

func (am AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {
}

func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.poeKeeper, am.contractKeeper, am.twasmKeeper))
}

func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return nil
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewGrpcQuerier(am.poeKeeper, am.twasmKeeper))
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.poeKeeper, am.contractKeeper, am.twasmKeeper))
}

func (am AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
}

func (am AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	am.doOnce.Do(ClearEmbeddedContracts) // release memory
	return EndBlocker(context, am.twasmKeeper)
}

// InitGenesis performs genesis initialization for the genutil module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	if len(genesisState.GenTxs) == 0 {
		panic(sdkerrors.Wrap(wasmtypes.ErrInvalidGenesis, "empty gentx"))
	}

	if genesisState.SeedContracts {
		if err := bootstrapPoEContracts(ctx, am.contractKeeper, am.twasmKeeper, am.poeKeeper, genesisState); err != nil {
			panic(fmt.Sprintf("bootstrap PoE contracts: %s", err))
		}
	} else {
		if err := verifyPoEContracts(ctx, am.contractKeeper, am.twasmKeeper, am.poeKeeper, genesisState); err != nil {
			panic(fmt.Sprintf("verify PoE bootstrap contracts: %s", err))
		}
	}

	if err := keeper.InitGenesis(ctx, am.poeKeeper, am.deliverTx, genesisState, am.txEncodingConfig); err != nil {
		panic(err)
	}
	// verify PoE setup
	addr, err := am.poeKeeper.GetPoEContractAddress(ctx, types.PoEContractTypeValset)
	if err != nil {
		panic(fmt.Sprintf("valset addr: %s", err))
	}
	switch ok, err := am.twasmKeeper.HasPrivilegedContract(ctx, addr, twasmtypes.PrivilegeTypeValidatorSetUpdate); {
	case err != nil:
		panic(fmt.Sprintf("valset contract: %s", err))
	case !ok:
		panic(fmt.Sprintf("valset contract not registered for valdator updates: %s", addr.String()))
	}

	// query validators from PoE for initial abci set
	switch diff, err := contract.CallEndBlockWithValidatorUpdate(ctx, addr, am.twasmKeeper); {
	case err != nil:
		panic(fmt.Sprintf("poe sudo call: %s", err))
	case len(diff) == 0:
		panic("initial valset must not be empty")
	default:
		return diff
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := keeper.ExportGenesis(ctx, am.poeKeeper)
	return cdc.MustMarshalJSON(gs)
}
