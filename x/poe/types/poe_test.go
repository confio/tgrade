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
			srcType: PoEContractTypeStaking,
		},
		"valset": {
			srcType: PoEContractTypeValset,
		},
		"engagement": {
			srcType: PoEContractTypeEngagement,
		},
		"mixer": {
			srcType: PoEContractTypeMixer,
		},
		"undefined": {
			srcType: PoEContractTypeUndefined,
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
			exp: PoEContractTypeStaking,
		},
		"VALSET": {
			src: "VALSET",
			exp: PoEContractTypeValset,
		},
		"ENGAGEMENT": {
			src: "ENGAGEMENT",
			exp: PoEContractTypeEngagement,
		},
		"MIXER": {
			src: "MIXER",
			exp: PoEContractTypeMixer,
		},
		"lower case ": {
			src: "staking",
			exp: PoEContractTypeUndefined,
		},
		"not in list": {
			src: "foobar",
			exp: PoEContractTypeUndefined,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := PoEContractTypeFrom(spec.src)
			assert.Equal(t, spec.exp, got)
		})
	}
}
