package types

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestPrivilegedCallbackTypeFrom(t *testing.T) {
	specs := map[string]struct {
		src    string
		expVal PrivilegedCallbackType
		expNil bool
	}{
		"begin block": {
			src:    "begin_block",
			expVal: PrivilegedCallbackType(0x1),
		},
		"end block": {
			src:    "end_block",
			expVal: PrivilegedCallbackType(0x2),
		},
		"validator update": {
			src:    "validator_set_update",
			expVal: PrivilegedCallbackType(0x3),
		},
		"unknown value": {
			src:    "unknown",
			expNil: true,
		},
		"invalid case": {
			src:    "BEGIN_BLOCK",
			expNil: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			got := PrivilegedCallbackTypeFrom(spec.src)
			if spec.expNil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, spec.expVal, *got)
		})
	}
}

func TestPrivilegedCallbackTypeValidation(t *testing.T) {
	specs := map[string]struct {
		src    PrivilegedCallbackType
		expErr bool
	}{
		"registered": {
			src: CallbackTypeBeginBlock,
		},
		"unregistered": {
			src:    math.MaxUint8,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}

}
