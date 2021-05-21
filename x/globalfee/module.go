package globalfee

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

const ModuleName = "globalfee"

var _ module.AppModuleBasic = AppModuleBasic{}
var _ module.AppModuleGenesis = AppModule{}
var _ module.AppModule = AppModule{}

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct{}

func (a AppModuleBasic) Name() string {
	return ModuleName
}

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return cdc.MustMarshalJSON(&GenesisState{
		Params: DefaultParams(),
	})
}

func (a AppModuleBasic) ValidateGenesis(marshaler codec.JSONMarshaler, config client.TxEncodingConfig, message json.RawMessage) error {
	var data GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	if err := data.Params.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "params")
	}
	return nil
}

func (a AppModuleBasic) RegisterInterfaces(registry types.InterfaceRegistry) {
}

func (a AppModuleBasic) RegisterRESTRoutes(context client.Context, router *mux.Router) {
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux) {
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
}

type AppModule struct {
	AppModuleBasic
	paramSpace paramstypes.Subspace
}

// NewAppModule constructor
func NewAppModule(paramSpace paramstypes.Subspace) *AppModule {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(ParamKeyTable())
	}

	return &AppModule{paramSpace: paramSpace}
}

func (a AppModule) InitGenesis(ctx sdk.Context, marshaler codec.JSONMarshaler, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	marshaler.MustUnmarshalJSON(message, &genesisState)
	a.paramSpace.SetParamSet(ctx, &genesisState.Params)
	return nil
}

func (a AppModule) ExportGenesis(ctx sdk.Context, marshaler codec.JSONMarshaler) json.RawMessage {
	var genState GenesisState
	a.paramSpace.GetParamSet(ctx, &genState.Params)
	return marshaler.MustMarshalJSON(&genState)
}
func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {
}

func (a AppModule) Route() sdk.Route {
	return sdk.Route{}
}

func (a AppModule) QuerierRoute() string {
	return ""
}

func (a AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return nil
}

func (a AppModule) RegisterServices(configurator module.Configurator) {
}

func (a AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
}

func (a AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}
