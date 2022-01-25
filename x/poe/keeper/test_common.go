package keeper

import (
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/app"

	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	"github.com/cosmos/cosmos-sdk/baseapp"
	params2 "github.com/cosmos/cosmos-sdk/simapp/params"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/core"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	ibckeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/confio/tgrade/x/poe/stakingadapter"
	"github.com/confio/tgrade/x/poe/types"
	"github.com/confio/tgrade/x/twasm"
	twasmkeeper "github.com/confio/tgrade/x/twasm/keeper"
	twasmtypes "github.com/confio/tgrade/x/twasm/types"
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
	twasm.AppModuleBasic{},
)

func makeEncodingConfig(t testing.TB) params2.EncodingConfig {
	r := types.MakeEncodingConfig(t)
	moduleBasics.RegisterLegacyAminoCodec(r.Amino)
	moduleBasics.RegisterInterfaces(r.InterfaceRegistry)
	twasmtypes.RegisterInterfaces(r.InterfaceRegistry)
	return r
}

type TestKeepers struct {
	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	GovKeeper      govkeeper.Keeper
	TWasmKeeper    *twasmkeeper.Keeper
	IBCKeeper      *ibckeeper.Keeper
	Router         *baseapp.Router
	PoEKeeper      Keeper
	EncodingConfig params2.EncodingConfig
	UpgradeKeeper  upgradekeeper.Keeper
	BaseApp        *baseapp.BaseApp
}

// CreateDefaultTestInput common settings for CreateTestInput
func CreateDefaultTestInput(t *testing.T, opts ...wasmkeeper.Option) (sdk.Context, TestKeepers) {
	return CreateTestInput(t, false, "staking,iterator,tgrade", opts...)
}

// CreateTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func CreateTestInput(t *testing.T, isCheckTx bool, supportedFeatures string, opts ...wasmkeeper.Option) (sdk.Context, TestKeepers) {
	// Load default wasm config
	return createTestInput(t, isCheckTx, supportedFeatures, twasmtypes.DefaultTWasmConfig(), dbm.NewMemDB(), opts...)
}

// encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func createTestInput(
	t *testing.T,
	isCheckTx bool,
	supportedFeatures string,
	wasmConfig twasmtypes.TWasmConfig,
	db dbm.DB,
	opts ...wasmkeeper.Option,
) (sdk.Context, TestKeepers) {
	tempDir := t.TempDir()

	keyPoE := sdk.NewKVStoreKey(types.StoreKey)
	keyWasm := sdk.NewKVStoreKey(twasmtypes.StoreKey)
	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyBank := sdk.NewKVStoreKey(banktypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	keyGov := sdk.NewKVStoreKey(govtypes.StoreKey)
	keyIBC := sdk.NewKVStoreKey(ibchost.StoreKey)
	keyCapability := sdk.NewKVStoreKey(capabilitytypes.StoreKey)
	keyCapabilityTransient := storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey)
	keyUpgrade := sdk.NewKVStoreKey(upgradetypes.StoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyPoE, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyWasm, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyBank, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyIBC, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyCapability, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyCapabilityTransient, sdk.StoreTypeMemory, db)
	ms.MountStoreWithDB(keyUpgrade, sdk.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())
	encodingConfig := makeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, keyParams, tkeyParams)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distributiontypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(capabilitytypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(types.ModuleName)

	maccPerms := map[string][]string{ // module account permissions
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		twasm.ModuleName:            {authtypes.Minter, authtypes.Burner},
	}
	authSubsp, _ := paramsKeeper.GetSubspace(authtypes.ModuleName)
	authKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keyAcc, // target store
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
		keyBank,
		authKeeper,
		bankSubsp,
		blockedAddrs,
	)
	bankParams := banktypes.DefaultParams()
	bankParams = bankParams.SetSendEnabledParam("utgd", true)
	bankKeeper.SetParams(ctx, bankParams)
	bankKeeper.SetSupply(ctx, banktypes.NewSupply(sdk.NewCoins(sdk.NewCoin("utgd", sdk.NewInt(1_000_000_000_000)))))

	capabilityKeeper := capabilitykeeper.NewKeeper(appCodec, keyCapability, keyCapabilityTransient)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedWasmKeeper := capabilityKeeper.ScopeToModule(twasmtypes.ModuleName)

	ibcSubsp, _ := paramsKeeper.GetSubspace(ibchost.ModuleName)

	var stakingKeeper clienttypes.StakingKeeper
	ibcKeeper := ibckeeper.NewKeeper(
		appCodec, keyIBC, ibcSubsp, stakingKeeper, scopedIBCKeeper,
	)
	upgradeKeeper := upgradekeeper.NewKeeper(map[int64]bool{}, keyUpgrade, appCodec, tempDir)
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(paramsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(upgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientUpdateProposalHandler(ibcKeeper.ClientKeeper))

	router := baseapp.NewRouter()
	bh := bank.NewHandler(bankKeeper)
	router.AddRoute(sdk.NewRoute(banktypes.RouterKey, bh))

	querier := baseapp.NewGRPCQueryRouter()
	banktypes.RegisterQueryServer(querier, bankKeeper)

	stakingAdapter := stakingadapter.NewStakingAdapter(nil, nil)
	twasmSubspace := paramsKeeper.Subspace(twasmtypes.DefaultParamspace)

	consensusParamsUpdater := baseapp.NewBaseApp("testApp", log.NewNopLogger(), db, nil)
	// set the BaseApp's parameter store
	consensusParamsUpdater.SetParamStore(paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	consensusParamsUpdater.StoreConsensusParams(ctx, app.DefaultConsensusParams)

	var twasmKeeper twasmkeeper.Keeper
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
			twasmkeeper.NewTgradeHandler(appCodec, &twasmKeeper, bankKeeper, consensusParamsUpdater, govRouter),
		)
	})

	opts = append([]wasmkeeper.Option{handler}, opts...)

	twasmKeeper = twasmkeeper.NewKeeper(
		appCodec,
		keyWasm,
		twasmSubspace,
		authKeeper,
		bankKeeper,
		stakingAdapter,
		nil,
		ibcKeeper.ChannelKeeper,
		&ibcKeeper.PortKeeper,
		scopedWasmKeeper,
		wasmtesting.MockIBCTransferKeeper{},
		router,
		querier,
		govRouter,
		tempDir,
		wasmConfig,
		supportedFeatures,
		opts...,
	)
	defaultParams := twasmtypes.DefaultParams()
	twasmSubspace.SetParamSet(ctx, &defaultParams)

	poeSubsp, _ := paramsKeeper.GetSubspace(types.ModuleName)
	poeKeeper := NewKeeper(appCodec, keyPoE, poeSubsp, twasmKeeper)
	router.AddRoute(sdk.NewRoute(twasmtypes.RouterKey, wasm.NewHandler(twasmKeeper.GetContractKeeper())))
	govRouter.AddRoute(twasm.RouterKey, twasmkeeper.NewProposalHandler(twasmKeeper))

	keepers := TestKeepers{
		AccountKeeper:  authKeeper,
		TWasmKeeper:    &twasmKeeper,
		BankKeeper:     bankKeeper,
		IBCKeeper:      ibcKeeper,
		Router:         router,
		PoEKeeper:      poeKeeper,
		UpgradeKeeper:  upgradeKeeper,
		EncodingConfig: encodingConfig,
		BaseApp:        consensusParamsUpdater,
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
	return rand.Bytes(sdk.AddrLen)
}

// createMinTestInput minimum integration test setup for this package
func createMinTestInput(t *testing.T) (sdk.Context, simappparams.EncodingConfig, Keeper) {
	var (
		keyPoe     = sdk.NewKVStoreKey(types.StoreKey)
		keyParams  = sdk.NewKVStoreKey(paramstypes.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyPoe, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	require.NoError(t, ms.LoadLatestVersion())
	encodingConfig := types.MakeEncodingConfig(t)

	paramsKeeper := paramskeeper.NewKeeper(encodingConfig.Marshaler, encodingConfig.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	k := NewKeeper(encodingConfig.Marshaler, keyPoe, paramsKeeper.Subspace(types.ModuleName), nil)
	return ctx, encodingConfig, k
}
