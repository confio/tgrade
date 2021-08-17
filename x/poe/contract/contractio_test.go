package contract

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	"testing"
)

func TestConvertToTendermintPubKey(t *testing.T) {
	var (
		ed25519pubkeybz   = ed25519.GenPrivKey().PubKey().Bytes()
		secp256k1pubkeybz = secp256k1.GenPrivKey().PubKey().Bytes()
	)
	specs := map[string]struct {
		src    ValidatorPubkey
		assert func(t *testing.T, v crypto.PublicKey)
		expErr bool
	}{
		"ed25519": {
			src: ValidatorPubkey{Ed25519: ed25519pubkeybz},
			assert: func(t *testing.T, v crypto.PublicKey) {
				require.Nil(t, v.GetSecp256K1())
				assert.Equal(t, ed25519pubkeybz, v.GetEd25519())
			},
		},
		"secp256k1": {
			src: ValidatorPubkey{Secp256k1: secp256k1pubkeybz},
			assert: func(t *testing.T, v crypto.PublicKey) {
				require.Nil(t, v.GetEd25519())
				assert.Equal(t, secp256k1pubkeybz, v.GetSecp256K1())
			},
		},
		"unsupported": {
			src:    ValidatorPubkey{},
			expErr: true,
		}}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotRes, gotErr := convertToTendermintPubKey(spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			spec.assert(t, gotRes)
		})
	}

}
