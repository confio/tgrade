package twasm

import (
	"context"
	"encoding/json"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/client/cli"
	"github.com/confio/tgrade/x/twasm/keeper"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math/rand"

	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmrest "github.com/CosmWasm/wasmd/x/wasm/client/rest"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmsimuliation "github.com/CosmWasm/wasmd/x/wasm/simulation"
	"github.com/confio/tgrade/x/twasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct {
}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx))
	wasmtypes.RegisterQueryHandlerClient(context.Background(), serveMux, wasmtypes.NewQueryClient(clientCtx))
}

// Name returns the wasm module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the wasm
// module.
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{
		Wasm: wasmtypes.GenesisState{
			Params: wasmtypes.DefaultParams(),
		},
	})
}

// ValidateGenesis performs genesis state validation for the wasm module.
func (b AppModuleBasic) ValidateGenesis(marshaler codec.JSONMarshaler, config client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	return nil
}

// RegisterRESTRoutes registers the REST routes for the wasm module.
func (AppModuleBasic) RegisterRESTRoutes(cliCtx client.Context, rtr *mux.Router) {
	wasmrest.RegisterRoutes(cliCtx, rtr)
}

// GetTxCmd returns the root tx command for the wasm module.
func (b AppModuleBasic) GetTxCmd() *cobra.Command {
	return wasmcli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the wasm module.
func (b AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaceTypes implements InterfaceModule
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

//____________________________________________________________________________

// AppModule implements an application module for the wasm module.
type AppModule struct {
	AppModuleBasic
	cdc                codec.Marshaler
	keeper             *keeper.Keeper
	validatorSetSource wasmkeeper.ValidatorSetSource
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Marshaler, keeper *keeper.Keeper, validatorSetSource wasmkeeper.ValidatorSetSource) AppModule {
	return AppModule{
		AppModuleBasic:     AppModuleBasic{},
		cdc:                cdc,
		keeper:             keeper,
		validatorSetSource: validatorSetSource,
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewGrpcQuerier(am.keeper))
	// wasm services
	wasmtypes.RegisterMsgServer(cfg.MsgServer(), wasmkeeper.NewMsgServerImpl(wasmkeeper.NewDefaultPermissionKeeper(am.keeper)))
	wasmtypes.RegisterQueryServer(cfg.QueryServer(), keeper.WasmQuerier(am.keeper))
}

func (am AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return wasmkeeper.NewLegacyQuerier(am.keeper, am.keeper.QueryGasLimit())
}

// RegisterInvariants registers the wasm module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the message routing key for the wasm module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, wasm.NewHandler(wasmkeeper.NewDefaultPermissionKeeper(am.keeper)))
}

// QuerierRoute returns the wasm module's querier route name.
func (AppModule) QuerierRoute() string {
	return wasmtypes.QuerierRoute
}

// InitGenesis performs genesis initialization for the wasm module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	validators, err := keeper.InitGenesis(ctx, am.keeper, genesisState, am.validatorSetSource, am.Route().Handler())
	if err != nil {
		panic(err)
	}
	return validators
}

// ExportGenesis returns the exported genesis state as raw bytes for the wasm
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := keeper.ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the wasm module.
func (am AppModule) BeginBlock(ctx sdk.Context, b abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper, b)
}

// EndBlock returns the end blocker for the wasm module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, b abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, am.keeper)
}

//____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the bank module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	wasmsimuliation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized bank param changes for the simulator.
func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return wasmsimuliation.ParamChanges(r, am.cdc)
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

//____________________________________________________________________________

// AddModuleInitFlags implements servertypes.ModuleInitFlags interface.
func AddModuleInitFlags(startCmd *cobra.Command) {
	wasm.AddModuleInitFlags(startCmd)
}

// ReadWasmConfig reads the wasm specifig configuration
func ReadWasmConfig(opts servertypes.AppOptions) (types.TWasmConfig, error) {
	cfg := types.TWasmConfig{}
	wasmConfig, err := wasm.ReadWasmConfig(opts)
	if err != nil {
		return cfg, sdkerrors.Wrap(err, "wasm")
	}
	cfg.WasmConfig = wasmConfig

	return cfg, nil
}
