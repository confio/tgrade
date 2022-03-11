package keeper

import (
	"testing"
	"time"

	appparams "github.com/confio/tgrade/app/params"

	"github.com/cosmos/cosmos-sdk/types/address"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/cosmos/ibc-go/v2/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v2/modules/core"
	ibchost "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v2/modules/core/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/confio/tgrade/x/twasm/types"
)

var moduleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	mint.AppModuleBasic{},
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
)

func MakeEncodingConfig(_ testing.TB) appparams.EncodingConfig {
	encodingConfig := appparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	moduleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	moduleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	return encodingConfig
}

var TestingStakeParams = stakingtypes.Params{
	UnbondingTime:     100,
	MaxValidators:     10,
	MaxEntries:        10,
	HistoricalEntries: 10,
	BondDenom:         "stake",
}

type TestKeepers struct {
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	GovKeeper     govkeeper.Keeper
	TWasmKeeper   *Keeper
	IBCKeeper     *ibckeeper.Keeper
	Router        wasmkeeper.MessageRouter
	Faucet        *wasmkeeper.TestFaucet
}

// CreateDefaultTestInput common settings for CreateTestInput
func CreateDefaultTestInput(t *testing.T, opts ...wasmkeeper.Option) (sdk.Context, TestKeepers) {
	return CreateTestInput(t, false, "staking", opts...)
}

// CreateTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func CreateTestInput(t *testing.T, isCheckTx bool, supportedFeatures string, opts ...wasmkeeper.Option) (sdk.Context, TestKeepers) {
	// Load default wasm config
	return createTestInput(t, isCheckTx, supportedFeatures, types.DefaultTWasmConfig(), dbm.NewMemDB(), opts...)
}

// encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func createTestInput(
	t *testing.T,
	isCheckTx bool,
	supportedFeatures string,
	wasmConfig types.TWasmConfig,
	db dbm.DB,
	opts ...wasmkeeper.Option,
) (sdk.Context, TestKeepers) {
	tempDir := t.TempDir()
	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distributiontypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey, feegrant.StoreKey, authzkeeper.StoreKey,
		types.StoreKey,
	)

	ms := store.NewCommitMultiStore(db)
	for _, v := range keys {
		ms.MountStoreWithDB(v, sdk.StoreTypeIAVL, db)
	}
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	for _, v := range tkeys {
		ms.MountStoreWithDB(v, sdk.StoreTypeTransient, db)
	}
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, sdk.StoreTypeMemory, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())
	encodingConfig := MakeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Codec, encodingConfig.Amino

	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	for _, m := range []string{authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		distributiontypes.ModuleName,
		slashingtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		capabilitytypes.ModuleName,
		ibchost.ModuleName,
		govtypes.ModuleName,
		types.ModuleName} {
		paramsKeeper.Subspace(m)
	}
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:     nil,
		distributiontypes.ModuleName:   nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		types.ModuleName:               {authtypes.Minter, authtypes.Burner},
		minttypes.ModuleName:           {authtypes.Minter, authtypes.Burner}, // for the faucet only
	}
	authSubsp, _ := paramsKeeper.GetSubspace(authtypes.ModuleName)
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey], // target store
		authSubsp,
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		allowReceivingFunds := acc != distributiontypes.ModuleName
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = allowReceivingFunds
	}

	bankSubsp, _ := paramsKeeper.GetSubspace(banktypes.ModuleName)
	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		accountKeeper,
		bankSubsp,
		blockedAddrs,
	)
	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		keys[upgradetypes.StoreKey],
		appCodec,
		tempDir,
		nil,
	)
	bankParams := banktypes.DefaultParams()
	bankParams = bankParams.SetSendEnabledParam("stake", true)
	bankKeeper.SetParams(ctx, bankParams)

	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedWasmKeeper := capabilityKeeper.ScopeToModule(types.ModuleName)

	faucet := wasmkeeper.NewTestFaucet(t, ctx, bankKeeper, types.ModuleName, sdk.NewCoin("utgd", sdk.NewInt(100_000_000_000)))

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		subspace(ibchost.ModuleName),
		nil,
		upgradeKeeper,
		scopedIBCKeeper,
	)

	querier := baseapp.NewGRPCQueryRouter()
	querier.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	configurator := module.NewConfigurator(appCodec, msgRouter, querier)
	bank.NewAppModule(appCodec, bankKeeper, accountKeeper).RegisterServices(configurator)

	var keeper Keeper
	handler := wasmkeeper.WithMessageHandlerDecorator(func(nested wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(
			// disable staking messages
			wasmkeeper.MessageHandlerFunc(func(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, err error) {
				if msg.Staking != nil {
					return nil, nil, sdkerrors.Wrap(wasmtypes.ErrExecuteFailed, "not supported, yet")
				}
				return nil, nil, wasmtypes.ErrUnknownMsg
			}),
			nested,
			// append our custom message handler
			NewTgradeHandler(appCodec, &keeper, bankKeeper, nil, nil),
		)
	})

	opts = append([]wasmkeeper.Option{handler}, opts...)
	keeper = NewKeeper(
		appCodec,
		keys[types.StoreKey],
		subspace(types.ModuleName),
		accountKeeper,
		bankKeeper,
		nil,
		nil,
		ibcKeeper.ChannelKeeper,
		&ibcKeeper.PortKeeper,
		scopedWasmKeeper,
		wasmtesting.MockIBCTransferKeeper{},
		msgRouter,
		querier,
		nil,
		tempDir,
		wasmConfig,
		supportedFeatures,
		opts...,
	)
	keeper.setParams(ctx, types.DefaultParams())

	types.RegisterQueryServer(querier, NewGrpcQuerier(keeper))
	// wasm services
	wasmtypes.RegisterMsgServer(querier, wasmkeeper.NewMsgServerImpl(wasmkeeper.NewDefaultPermissionKeeper(keeper)))
	wasmtypes.RegisterQueryServer(querier, WasmQuerier(&keeper))

	keepers := TestKeepers{
		AccountKeeper: accountKeeper,
		TWasmKeeper:   &keeper,
		BankKeeper:    bankKeeper,
		IBCKeeper:     ibcKeeper,
		Router:        msgRouter,
		Faucet:        faucet,
	}
	return ctx, keepers
}

// NewWasmVMMock creates a new WasmerEngine mock with basic ops for create/instantiation set to noops.
func NewWasmVMMock(mutators ...func(*wasmtesting.MockWasmer)) *wasmtesting.MockWasmer {
	mock := &wasmtesting.MockWasmer{
		CreateFn:      wasmtesting.HashOnlyCreateFn,
		InstantiateFn: wasmtesting.NoOpInstantiateFn,
		AnalyzeCodeFn: wasmtesting.HasIBCAnalyzeFn,
	}
	for _, m := range mutators {
		m(mock)
	}
	return mock

}

func RandomAddress(_ *testing.T) sdk.AccAddress {
	return rand.Bytes(address.Len)
}
