package poe

import (
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
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
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
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
}

// DefaultGenesis returns default genesis state as raw bytes for the genutil
// module.
func (b AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	gs := DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

// ValidateGenesis performs genesis state validation for the genutil module.
func (b AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, txEncodingConfig client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	// todo: add PoE validation

	//return types.ValidateGenesis(&data, txEncodingConfig.TxJSONDecoder())
	return nil
}

// RegisterRESTRoutes registers the REST routes for the genutil module.
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the genutil module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {
}

// GetTxCmd returns no root tx command for the genutil module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the genutil module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

//____________________________________________________________________________

// AppModule implements an application module for the genutil module.
type AppModule struct {
	AppModuleBasic

	accountKeeper    genutiltypes.AccountKeeper
	stakingKeeper    genutiltypes.StakingKeeper
	deliverTx        deliverTxfn
	txEncodingConfig client.TxEncodingConfig
	twasmKeeper      twasmKeeper
	contractKeeper   wasmtypes.ContractOpsKeeper
	poeKeeper        keeper.Keeper
}

// twasmKeeper subset of keeper to decouple from twasm module
type twasmKeeper interface {
	QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) (*sdk.Result, error)
	SetPrivileged(ctx sdk.Context, contractAddr sdk.AccAddress) error
	HasPrivilegedContractCallback(ctx sdk.Context, contractAddr sdk.AccAddress, callbackType twasmtypes.PrivilegedCallbackType) (bool, error)
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	poeKeeper keeper.Keeper,
	accountKeeper genutiltypes.AccountKeeper,
	stakingKeeper genutiltypes.StakingKeeper,
	twasmKeeper twasmKeeper,
	deliverTx deliverTxfn,
	txEncodingConfig client.TxEncodingConfig,
	contractKeeper wasmtypes.ContractOpsKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic:   AppModuleBasic{},
		accountKeeper:    accountKeeper,
		twasmKeeper:      twasmKeeper,
		contractKeeper:   contractKeeper,
		poeKeeper:        poeKeeper,
		stakingKeeper:    stakingKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
	}
}

func (am AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {
}

func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.poeKeeper, am.contractKeeper))
}

func (am AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return nil
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.poeKeeper, am.contractKeeper))
}

func (am AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
}

func (am AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

// InitGenesis performs genesis initialization for the genutil module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := bootstrapPoEContracts(ctx, am.contractKeeper, am.twasmKeeper, am.poeKeeper, genesisState); err != nil {
		panic(fmt.Sprintf("bootstrap: %s", err))
	}

	_, err := InitGenesis(ctx, am.stakingKeeper, am.deliverTx, genesisState, am.txEncodingConfig)
	if err != nil {
		panic(err)
	}
	addr, err := am.poeKeeper.GetPoeContractAddress(ctx, types.PoEContractTypes_VALSET)
	if err != nil {
		panic(fmt.Sprintf("valset addr: %s", err))
	}
	switch ok, err := am.twasmKeeper.HasPrivilegedContractCallback(ctx, addr, twasmtypes.CallbackTypeValidatorSetUpdate); {
	case err != nil:
		panic(fmt.Sprintf("valset contract: %s", err))
	case !ok:
		panic(fmt.Sprintf("valset contract not registered for valdator updates: %s", addr.String()))
	}

	diff, err := callValidatorSetUpdaterContract(ctx, addr, am.twasmKeeper)
	if err != nil {
		panic(fmt.Sprintf("poe sudo call: %s", err))
	}
	if len(diff) == 0 {
		panic("initial valset must not be empty")
	}
	return diff
}

func getPubKey(key contract.ValidatorPubkey) crypto.PublicKey {
	return crypto.PublicKey{
		Sum: &crypto.PublicKey_Ed25519{
			Ed25519: key.Ed25519,
		},
	}
}

func callValidatorSetUpdaterContract(ctx sdk.Context, contractAddr sdk.AccAddress, k twasmKeeper) ([]abci.ValidatorUpdate, error) {
	sudoMsg := contract.TgradeSudoMsg{EndWithValidatorUpdate: &struct{}{}}
	msgBz, err := json.Marshal(sudoMsg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "tgrade sudo msg")
	}
	resp, err := k.Sudo(ctx, contractAddr, msgBz)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "sudo")
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	var contractResult contract.EndWithValidatorUpdateResponse
	if err := json.Unmarshal(resp.Data, &contractResult); err != nil {
		return nil, sdkerrors.Wrap(err, "contract response")
	}
	if len(contractResult.Diffs) == 0 {
		return nil, nil
	}

	result := make([]abci.ValidatorUpdate, len(contractResult.Diffs))
	for i, v := range contractResult.Diffs {
		result[i] = abci.ValidatorUpdate{
			PubKey: getPubKey(v.PubKey),
			Power:  int64(v.Power),
		}
	}
	keeper.ModuleLogger(ctx).Info("privileged contract callback", "type", "validator-set-update", "result", result)
	return result, nil
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil
// module.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	var gs types.GenesisState
	return cdc.MustMarshalJSON(&gs)
}
