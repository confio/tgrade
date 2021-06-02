package types

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPoEContractTypeValidate(t *testing.T) {
	specs := map[string]struct {
		srcType PoEContractType
		expErr  bool
	}{
		"staking": {
			srcType: PoEContractType_STAKING,
		},
		"valset": {
			srcType: PoEContractType_VALSET,
		},
		"engagement": {
			srcType: PoEContractType_ENGAGEMENT,
		},
		"mixer": {
			srcType: PoEContractType_MIXER,
		},
		"undefined": {
			srcType: PoEContractType_UNDEFINED,
			expErr:  true,
		},
		"unsupported type": {
			srcType: PoEContractType(9999),
			expErr:  true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.srcType.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestPoEContractTypeFrom(t *testing.T) {
	specs := map[string]struct {
		src string
		exp PoEContractType
	}{
		"STAKING": {
			src: "STAKING",
			exp: PoEContractType_STAKING,
		},
		"VALSET": {
			src: "VALSET",
			exp: PoEContractType_VALSET,
		},
		"ENGAGEMENT": {
			src: "ENGAGEMENT",
			exp: PoEContractType_ENGAGEMENT,
		},
		"MIXER": {
			src: "MIXER",
			exp: PoEContractType_MIXER,
		},
		"lower case ": {
			src: "staking",
			exp: PoEContractType_UNDEFINED,
		},
		"not in list": {
			src: "foobar",
			exp: PoEContractType_UNDEFINED,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := PoEContractTypeFrom(spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}
