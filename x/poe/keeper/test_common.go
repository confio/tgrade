package keeper

import (
	"github.com/confio/tgrade/x/poe/types"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
	"time"
)

var _ ContractSource = ContractSourceMock{}

// ContractSourceMock implementes ContractSource interface for testing purpose
type ContractSourceMock struct {
	GetPoEContractAddressFn func(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error)
}

func (m ContractSourceMock) GetPoEContractAddress(ctx sdk.Context, ctype types.PoEContractType) (sdk.AccAddress, error) {
	if m.GetPoEContractAddressFn == nil {
		panic("not expected to be called")
	}
	return m.GetPoEContractAddressFn(ctx, ctype)
}

func CreateDefaultTestInput(t *testing.T) (sdk.Context, simappparams.EncodingConfig, Keeper) {
	keyPoe := sdk.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyPoe, sdk.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	encodingConfig := types.MakeEncodingConfig(t)
	k := NewKeeper(encodingConfig.Marshaler, keyPoe)
	return ctx, encodingConfig, k
}
