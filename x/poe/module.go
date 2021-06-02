package poe

import (
	"encoding/json"
	"fmt"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/poe/contract"
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
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

const ModuleName = genutiltypes.ModuleName // todo (Alex): rename to POE

var (
	_ module.AppModuleGenesis = AppModule{}
	_ module.AppModuleBasic   = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the genutil module.
type AppModuleBasic struct {
}

// Name returns the genutil module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec registers the genutil module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

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
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", ModuleName, err)
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
	accountKeeper genutiltypes.AccountKeeper,
	stakingKeeper genutiltypes.StakingKeeper,
	twasmKeeper twasmKeeper,
	deliverTx deliverTxfn,
	txEncodingConfig client.TxEncodingConfig,
	contractKeeper wasmtypes.ContractOpsKeeper,
) module.AppModule {
	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic:   AppModuleBasic{},
		accountKeeper:    accountKeeper,
		twasmKeeper:      twasmKeeper,
		contractKeeper:   contractKeeper,
		stakingKeeper:    stakingKeeper,
		deliverTx:        deliverTx,
		txEncodingConfig: txEncodingConfig,
	})
}

// InitGenesis performs genesis initialization for the genutil module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	if err := bootstrapPoEContracts(ctx, am.contractKeeper, am.twasmKeeper, genesisState); err != nil {
		panic(fmt.Sprintf("bootstrap: %s", err))
	}

	_, err := InitGenesis(ctx, am.stakingKeeper, am.deliverTx, genesisState, am.txEncodingConfig)
	if err != nil {
		panic(err)
	}
	addr, err := sdk.AccAddressFromBech32(genesisState.ValsetContractAddr)
	if err != nil {
		panic(fmt.Sprintf("invalid valset addr: %s", err))
	}
	switch ok, err := am.twasmKeeper.HasPrivilegedContractCallback(ctx, addr, twasmtypes.CallbackTypeValidatorSetUpdate); {
	case err != nil:
		panic(fmt.Sprintf("valset contract: %s", err))
	case !ok:
		panic(fmt.Sprintf("valset contract not registered for valdator updates: %s", genesisState.ValsetContractAddr))
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
	ModuleLogger(ctx).Info("privileged contract callback", "type", "validator-set-update", "result", result)
	return result, nil
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil
// module.
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	var gs types.GenesisState
	return cdc.MustMarshalJSON(&gs)
}

func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}
