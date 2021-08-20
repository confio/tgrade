package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	yaml "gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// PoE params default values
const (

	// DefaultHistoricalEntries entries is 10000. Apps that don't use IBC can ignore this
	// value by not adding the staking module to the application module manager's
	// SetOrderBeginBlockers.
	DefaultHistoricalEntries                uint32 = 10000
	DefaultInitialValidatorEngagementPoints uint64 = 1
)

var (
	KeyHistoricalEntries          = []byte("HistoricalEntries")
	KeyInitialValEngagementPoints = []byte("InitialValidatorEngagementPoints")
	KeyMinDelegationAmounts       = []byte("MinDelegationAmounts")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable for PoE module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(historicalEntries uint32, engagementPoints uint64, min sdk.Coins) Params {
	return Params{
		HistoricalEntries:          historicalEntries,
		InitialValEngagementPoints: engagementPoints,
		MinDelegationAmounts:       min,
	}
}

// ParamSetPairs Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateUint32),
		paramtypes.NewParamSetPair(KeyInitialValEngagementPoints, &p.InitialValEngagementPoints, validateUint64),
		paramtypes.NewParamSetPair(KeyMinDelegationAmounts, &p.MinDelegationAmounts, validateSDKCoins),
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultHistoricalEntries,
		DefaultInitialValidatorEngagementPoints,
		sdk.Coins{},
	)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validate a set of params
func (p Params) Validate() error {
	return sdkerrors.Wrap(p.MinDelegationAmounts.Validate(), "min delegation amounts")
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
func validateUint32(i interface{}) error {
	_, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateSDKCoins(i interface{}) error {
	c, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return c.Validate()
}
