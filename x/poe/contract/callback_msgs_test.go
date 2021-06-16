package contract

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewValidatorPubkey(t *testing.T) {
	specs := map[string]struct {
		src    cryptotypes.PubKey
		expErr bool
		assert func(t *testing.T, v ValidatorPubkey)
	}{
		"ed25519": {
			src: ed25519.GenPrivKey().PubKey(),
			assert: func(t *testing.T, v ValidatorPubkey) {
				assert.NotEmpty(t, v.Ed25519)
				assert.Empty(t, v.Secp256k1)
			},
		},
		"secp256k1": {
			src: secp256k1.GenPrivKey().PubKey(),
			assert: func(t *testing.T, v ValidatorPubkey) {
				assert.Empty(t, v.Ed25519)
				assert.NotEmpty(t, v.Secp256k1)
			},
		},
		"unsupported": {
			src:    unknownPubKey{},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRes, gotErr := NewValidatorPubkey(spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			spec.assert(t, gotRes)
		})
	}
}

type unknownPubKey struct {
	cryptotypes.PubKey
}

func (unknownPubKey) Type() string {
	return "unknown"
}
