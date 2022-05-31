package keeper

import (
	"testing"
	"time"

	wasmapp "github.com/CosmWasm/wasmd/app"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	params2 "github.com/cosmos/cosmos-sdk/simapp/params"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
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
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	transfer "github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	poestakingadapter "github.com/confio/tgrade/x/poe/stakingadapter"
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
	PoEKeeper      Keeper
	EncodingConfig params2.EncodingConfig
	UpgradeKeeper  upgradekeeper.Keeper
	Faucet         *wasmkeeper.TestFaucet
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
	t testing.TB,
	isCheckTx bool,
	supportedFeatures string,
	wasmConfig twasmtypes.TWasmConfig,
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
		twasmtypes.StoreKey,
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
	ctx = wasmtypes.WithTXCounter(ctx, 0)

	encodingConfig := makeEncodingConfig(t)
	appCodec, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	paramsKeeper := paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	for _, m := range []string{
		authtypes.ModuleName,
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
		twasmtypes.ModuleName,
		types.ModuleName,
	} {
		paramsKeeper.Subspace(m)
	}
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}
	maccPerms := map[string][]string{ // module account permissions
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		twasm.ModuleName:            {authtypes.Minter, authtypes.Burner},
		types.ModuleName:            {authtypes.Minter, authtypes.Burner},
		types.BondedPoolName:        {authtypes.Burner, authtypes.Staking},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey], // target store
		subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		accountKeeper,
		subspace(banktypes.ModuleName),
		blockedAddrs,
	)
	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		keys[upgradetypes.StoreKey],
		appCodec,
		tempDir,
		nil,
	)
	stakingAdapter := poestakingadapter.StakingAdapter{}

	bankParams := banktypes.DefaultParams()
	bankParams = bankParams.SetSendEnabledParam("utgd", true)
	bankKeeper.SetParams(ctx, bankParams)

	capabilityKeeper := capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	scopedIBCKeeper := capabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedWasmKeeper := capabilityKeeper.ScopeToModule(twasmtypes.ModuleName)

	var twasmKeeper twasmkeeper.Keeper
	poeKeeper := NewKeeper(
		appCodec,
		keys[types.StoreKey],
		subspace(types.ModuleName),
		&twasmKeeper,
		accountKeeper,
	)
	poeKeeper.setParams(ctx, types.DefaultParams())

	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		subspace(ibchost.ModuleName),
		&poeKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
	)
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(paramsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(upgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(ibcKeeper.ClientKeeper))

	querier := baseapp.NewGRPCQueryRouter()
	querier.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	configurator := module.NewConfigurator(appCodec, msgRouter, querier)
	bank.NewAppModule(appCodec, bankKeeper, accountKeeper).RegisterServices(configurator)

	cfg := sdk.GetConfig()
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())

	consensusParamsUpdater := baseapp.NewBaseApp("testApp", log.NewNopLogger(), db, nil)
	// set the BaseApp's parameter store
	consensusParamsUpdater.SetParamStore(paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))

	consensusParamsUpdater.StoreConsensusParams(ctx, wasmapp.DefaultConsensusParams)

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
		keys[twasmtypes.StoreKey],
		subspace(twasmtypes.ModuleName),
		accountKeeper,
		bankKeeper,
		stakingAdapter,
		nil,
		ibcKeeper.ChannelKeeper,
		&ibcKeeper.PortKeeper,
		scopedWasmKeeper,
		wasmtesting.MockIBCTransferKeeper{},
		msgRouter,
		querier,
		govRouter,
		tempDir,
		wasmConfig,
		supportedFeatures,
		opts...,
	)
	twasmKeeper.SetParams(ctx, twasmtypes.DefaultParams())

	twasm.NewAppModule(appCodec, &twasmKeeper, poestakingadapter.StakingAdapter{}, accountKeeper, bankKeeper).RegisterServices(configurator)
	govRouter.AddRoute(twasm.RouterKey, twasmkeeper.NewProposalHandler(twasmKeeper))

	faucet := wasmkeeper.NewTestFaucet(t, ctx, bankKeeper, types.ModuleName, sdk.NewCoin("utgd", sdk.NewInt(100_000_000_000)))

	keepers := TestKeepers{
		AccountKeeper:  accountKeeper,
		TWasmKeeper:    &twasmKeeper,
		BankKeeper:     bankKeeper,
		IBCKeeper:      ibcKeeper,
		PoEKeeper:      poeKeeper,
		UpgradeKeeper:  upgradeKeeper,
		EncodingConfig: encodingConfig,
		BaseApp:        consensusParamsUpdater,
		Faucet:         faucet,
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

// createMinTestInput minimum integration test setup for this package
func createMinTestInput(t *testing.T) (sdk.Context, simappparams.EncodingConfig, Keeper) {
	var (
		keyPoe     = sdk.NewKVStoreKey(types.StoreKey)
		keyParams  = sdk.NewKVStoreKey(paramstypes.StoreKey)
		tkeyParams = sdk.NewTransientStoreKey(paramstypes.TStoreKey)
		keyAuth    = sdk.NewKVStoreKey(authtypes.StoreKey)
	)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyPoe, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyAuth, sdk.StoreTypeIAVL, db)

	require.NoError(t, ms.LoadLatestVersion())
	encodingConfig := types.MakeEncodingConfig(t)

	appCodec := encodingConfig.Marshaler

	paramsKeeper := paramskeeper.NewKeeper(encodingConfig.Marshaler, encodingConfig.Amino, keyParams, tkeyParams)

	paramsKeeper.Subspace(authtypes.ModuleName)
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}
	maccPerms := map[string][]string{ // module account permissions
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		twasm.ModuleName:            {authtypes.Minter, authtypes.Burner},
		minttypes.ModuleName:        {authtypes.Minter, authtypes.Burner}, // for the faucet only
		types.BondedPoolName:        {authtypes.Burner, authtypes.Staking},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		keyPoe,
		subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
	)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	k := NewKeeper(encodingConfig.Marshaler, keyPoe, paramsKeeper.Subspace(types.ModuleName), nil, accountKeeper)
	return ctx, encodingConfig, k
}
