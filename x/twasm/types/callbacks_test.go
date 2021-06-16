package types

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestPrivilegedCallbackTypeFrom(t *testing.T) {
	specs := map[string]struct {
		src    string
		expVal PrivilegeType
		expNil bool
	}{
		"begin blocker": {
			src:    "begin_blocker",
			expVal: PrivilegeType(0x1),
		},
		"end blocker": {
			src:    "end_blocker",
			expVal: PrivilegeType(0x2),
		},
		"validator update": {
			src:    "validator_set_updater",
			expVal: PrivilegeType(0x3),
		},
		"unknown value": {
			src:    "unknown",
			expNil: true,
		},
		"invalid case": {
			src:    "BEGIN_BLOCKER",
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
		src    PrivilegeType
		expErr bool
	}{
		"registered": {
			src: PrivilegeTypeBeginBlock,
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

func TestPrivilegedCallbackTypeSingletons(t *testing.T) {
	// sanity check with manually curated list
	specs := map[PrivilegeType]bool{
		PrivilegeTypeBeginBlock:          false,
		PrivilegeTypeEndBlock:            false,
		PrivilegeTypeValidatorSetUpdate:  true,
		PrivilegeTypeGovProposalExecutor: false,
	}
	for c, exp := range specs {
		t.Run(c.String(), func(t *testing.T) {
			assert.Equal(t, c.IsSingleton(), exp)
		})
	}
	require.Len(t, specs, len(AllCallbackTypeNames()), "got %v", AllCallbackTypeNames())
}

func TestPrivilegeTypeMarshalJson(t *testing.T) {
	unknownType := PrivilegeType(0)
	specs := map[string]struct {
		src     interface{}
		expJson []byte
		expErr  bool
	}{
		"all good": {
			src:     &PrivilegeTypeBeginBlock,
			expJson: []byte(`"begin_blocker"`),
		},
		"empty": {
			src:     (*PrivilegeType)(nil),
			expJson: []byte(`null`),
		},
		"empty obj": {
			src: struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{},
			expJson: []byte(`{}`),
		},
		"unknown": {
			src:    &unknownType,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotVal, gotErr := json.Marshal(spec.src)
			if spec.expErr {
				require.Error(t, gotErr)
				assert.Nil(t, gotVal)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, spec.expJson, gotVal)
		})
	}
}

func TestPrivilegeTypeUnmarshalJson(t *testing.T) {
	var x PrivilegeType
	specs := map[string]struct {
		src    []byte
		target interface{}
		exp    interface{}
		expErr bool
	}{
		"all good": {
			src:    []byte(`"begin_blocker"`),
			target: &x,
			exp:    &PrivilegeTypeBeginBlock,
		},
		"null": {
			src:    []byte(`null`),
			expErr: true,
		},
		"empty obj": {
			src: []byte(`{}`),
			target: &struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{},
			exp: &struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{},
		},
		"empty obj value": {
			src: []byte(`{"foo":""}`),
			target: &struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{},
			expErr: true,
		},
		"obj value set": {
			src: []byte(`{"foo":"begin_blocker"}`),
			target: &struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{},
			exp: &struct {
				Foo *PrivilegeType `json:"foo,omitempty"`
			}{Foo: &PrivilegeTypeBeginBlock},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := json.Unmarshal(spec.src, spec.target)
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			require.Equal(t, spec.exp, spec.target)
		})
	}
}
