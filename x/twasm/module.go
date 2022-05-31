package twasm

import (
	"context"
	"encoding/json"
	"math/rand"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmrest "github.com/CosmWasm/wasmd/x/wasm/client/rest"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmsimulation "github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/confio/tgrade/x/twasm/client/cli"
	"github.com/confio/tgrade/x/twasm/keeper"
	"github.com/confio/tgrade/x/twasm/simulation"
	"github.com/confio/tgrade/x/twasm/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct{}

func (b AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
	if err := wasmtypes.RegisterQueryHandlerClient(context.Background(), serveMux, wasmtypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// Name returns the wasm module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the wasm
// module.
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{
		Params: wasmtypes.DefaultParams(),
	})
}

// ValidateGenesis performs genesis state validation for the wasm module.
func (b AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	if err := marshaler.UnmarshalJSON(message, &data); err != nil {
		return sdkerrors.Wrap(err, "validate genesis")
	}

	return data.ValidateBasic()
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

// AppModule implements an application module for the wasm module.
type AppModule struct {
	AppModuleBasic
	cdc                codec.Codec
	keeper             *keeper.Keeper
	validatorSetSource wasmkeeper.ValidatorSetSource
	accountKeeper      wasmtypes.AccountKeeper // for simulation
	bankKeeper         simulation.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
	validatorSetSource wasmkeeper.ValidatorSetSource,
	ak wasmtypes.AccountKeeper,
	bk simulation.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic:     AppModuleBasic{},
		cdc:                cdc,
		keeper:             keeper,
		validatorSetSource: validatorSetSource,
		accountKeeper:      ak,
		bankKeeper:         bk,
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))
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
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	validators, err := keeper.InitGenesis(ctx, am.keeper, genesisState, am.Route().Handler())
	if err != nil {
		panic(err)
	}
	return validators
}

// ExportGenesis returns the exported genesis state as raw bytes for the wasm
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
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

// GenerateGenesisState creates a randomized GenState of the bank module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized bank param changes for the simulator.
func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return wasmsimulation.ParamChanges(r, am.cdc)
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore()
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(&simState, am.accountKeeper, am.bankKeeper, am.keeper)
}

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

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
