package stakingadapter

import (
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

var _ evidencetypes.SlashingKeeper = &SlashingAdapter{}

type SlashingAdapter struct{}

func (s SlashingAdapter) GetPubkey(ctx sdk.Context, address cryptotypes.Address) (cryptotypes.PubKey, error) {
	log(ctx, "GetPubkey")
	return nil, ErrNotImplemented
}

func (s SlashingAdapter) IsTombstoned(ctx sdk.Context, address sdk.ConsAddress) bool {
	log(ctx, "IsTombstoned")
	return false
}

func (s SlashingAdapter) HasValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) bool {
	log(ctx, "HasValidatorSigningInfo")
	return false
}

func (s SlashingAdapter) Tombstone(ctx sdk.Context, address sdk.ConsAddress) {
	log(ctx, "Tombstone")
}

func (s SlashingAdapter) Slash(ctx sdk.Context, address sdk.ConsAddress, dec sdk.Dec, i int64, i2 int64) {
	log(ctx, "Slash")
}

func (s SlashingAdapter) SlashFractionDoubleSign(ctx sdk.Context) sdk.Dec {
	log(ctx, "SlashFractionDoubleSign")
	return sdk.ZeroDec()
}

func (s SlashingAdapter) Jail(ctx sdk.Context, address sdk.ConsAddress) {
	log(ctx, "Jail")
}

func (s SlashingAdapter) JailUntil(ctx sdk.Context, address sdk.ConsAddress, time time.Time) {
	log(ctx, "JailUntil")
}
