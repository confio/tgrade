package types

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestTgradeContractDetailsValidation(t *testing.T) {
	specs := map[string]struct {
		src    TgradeContractDetails
		expErr bool
	}{
		"all good": {
			src: TgradeContractDetailsFixture(t),
		},
		"empty callbacks": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = nil
			}),
		},
		"multiple callbacks": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{
					{Position: 1, PrivilegeType: "begin_blocker"},
					{Position: 1, PrivilegeType: "end_blocker"},
				}
			}),
		},
		"duplicate callbacks": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{
					{Position: 1, PrivilegeType: "begin_blocker"},
					{Position: 2, PrivilegeType: "begin_blocker"},
				}
			}),
			expErr: true,
		},
		"unknown callback": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{{Position: 1, PrivilegeType: "unknown"}}
			}),
			expErr: true,
		},
		"empty callback": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{{Position: 1}}
			}),
			expErr: true,
		},
		"invalid callback position": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{{Position: math.MaxUint8 + 1, PrivilegeType: "begin_blocker"}}
			}),
			expErr: true,
		},
		"empty callback position": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredPrivileges = []RegisteredPrivilege{{PrivilegeType: "begin_blocker"}}
			}),
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			if spec.expErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
		})
	}

}
