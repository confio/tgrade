package stakingadapter

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"time"
)

var _ evidencetypes.SlashingKeeper = &SlashingAdapter{}

type SlashingAdapter struct{}

func (s SlashingAdapter) GetPubkey(context sdk.Context, address cryptotypes.Address) (cryptotypes.PubKey, error) {
	panic("implement me")
}

func (s SlashingAdapter) IsTombstoned(context sdk.Context, address sdk.ConsAddress) bool {
	panic("implement me")
}

func (s SlashingAdapter) HasValidatorSigningInfo(context sdk.Context, address sdk.ConsAddress) bool {
	panic("implement me")
}

func (s SlashingAdapter) Tombstone(context sdk.Context, address sdk.ConsAddress) {
	panic("implement me")
}

func (s SlashingAdapter) Slash(context sdk.Context, address sdk.ConsAddress, dec sdk.Dec, i int64, i2 int64) {
	panic("implement me")
}

func (s SlashingAdapter) SlashFractionDoubleSign(context sdk.Context) sdk.Dec {
	panic("implement me")
}

func (s SlashingAdapter) Jail(context sdk.Context, address sdk.ConsAddress) {
	panic("implement me")
}

func (s SlashingAdapter) JailUntil(context sdk.Context, address sdk.ConsAddress, time time.Time) {
	panic("implement me")
}
