package contract_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/poe/contract"
)

func TestToValidator(t *testing.T) {
	// create Public Key
	pk, err := contract.NewValidatorPubkey(ed25519.GenPrivKey().PubKey())
	require.NoError(t, err)
	specs := map[string]struct {
		operatorResponse contract.OperatorResponse
		expStatus        stakingtypes.BondStatus
	}{
		"active validator": {
			operatorResponse: contract.OperatorResponse{
				Pubkey:          pk,
				ActiveValidator: true,
			},
			expStatus: stakingtypes.Bonded,
		},
		"unactive validator": {
			operatorResponse: contract.OperatorResponse{
				Pubkey:          pk,
				ActiveValidator: false,
			},
			expStatus: stakingtypes.Unbonded,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotVal, err := spec.operatorResponse.ToValidator()
			require.NoError(t, err)
			assert.Equal(t, spec.expStatus, gotVal.Status)
		})
	}
}
