package testing

import (
	"encoding/json"
	"fmt"
	poetypes "github.com/confio/tgrade/x/poe/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"testing"
)

// SetPoEParamsMutator set params in genesis
func SetPoEParamsMutator(t *testing.T, params poetypes.Params) GenesisMutator {
	return func(genesis []byte) []byte {
		t.Helper()
		val, err := json.Marshal(params)
		require.NoError(t, err)
		state, err := sjson.SetRawBytes(genesis, "app_state.poe.params", val)
		require.NoError(t, err)
		return state
	}
}

// SetGlobalMinFee set the passed coins to the global minimum fee
func SetGlobalMinFee(t *testing.T, fees ...sdk.DecCoin) GenesisMutator {
	return func(genesis []byte) []byte {
		t.Helper()
		coins := sdk.NewDecCoins(fees...)
		require.NoError(t, coins.Validate())
		val, err := json.Marshal(coins)
		require.NoError(t, err)
		state, err := sjson.SetRawBytes(genesis, "app_state.globalfee.params.minimum_gas_prices", val)
		require.NoError(t, err)
		return state
	}
}

// SetAllEngagementPoints set the given value for all members of the engament group (default = validators)
func SetAllEngagementPoints(t *testing.T, points int) GenesisMutator {
	return func(raw []byte) []byte {
		group := gjson.GetBytes(raw, "app_state.poe.engagement").Array()
		for i := range group {
			var err error
			raw, err = sjson.SetRawBytes(raw, fmt.Sprintf("app_state.poe.engagement.%d.weight", i), []byte(fmt.Sprintf(`"%d"`, points)))
			require.NoError(t, err)
		}
		return raw
	}
}
