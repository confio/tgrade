package keeper

import (
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"

	"github.com/confio/tgrade/x/poe/types"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGetSetHistoricalInfo(t *testing.T) {
	ctx, example := CreateDefaultTestInput(t)
	keeper := example.PoEKeeper
	var header tmproto.Header
	f := fuzz.New()
	f.Fuzz(&header)

	ctx = ctx.WithBlockHeight(1).WithBlockHeader(header)
	exp := stakingtypes.NewHistoricalInfo(ctx.BlockHeader(), nil)
	keeper.SetHistoricalInfo(ctx, 1, &exp)

	// when
	got, exists := keeper.GetHistoricalInfo(ctx, 1)

	// then
	require.True(t, exists)
	assert.Equal(t, exp, got)
}

func TestTrackHistoricalInfo(t *testing.T) {
	ctx, example := CreateDefaultTestInput(t)
	keeper := example.PoEKeeper
	const maxEntries = 2
	keeper.setParams(ctx, types.Params{HistoricalEntries: maxEntries})

	// fill all slots
	expEntries := make([]stakingtypes.HistoricalInfo, 0, maxEntries+1)
	for i := 0; i < maxEntries; i++ {
		var header tmproto.Header
		f := fuzz.New()
		f.Fuzz(&header)
		header.Height = int64(1 + i)
		header.Time = time.Now().UTC()
		keeper.TrackHistoricalInfo(ctx.WithBlockHeader(header))
		expEntries = append(expEntries, stakingtypes.NewHistoricalInfo(header, nil))
	}

	// when new element added
	var header tmproto.Header
	f := fuzz.New()
	f.Fuzz(&header)
	header.Height = int64(1 + maxEntries)
	header.Time = time.Now().UTC()
	keeper.TrackHistoricalInfo(ctx.WithBlockHeader(header))

	// then only last max entries stored
	_, exists := keeper.GetHistoricalInfo(ctx, 1)
	require.False(t, exists)
	expEntries = append(expEntries, stakingtypes.NewHistoricalInfo(header, nil))
	assert.Equal(t, expEntries[1:], keeper.getAllHistoricalInfo(ctx))
}
