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
				d.RegisteredCallbacks = nil
			}),
		},
		"multiple callbacks": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{
					{Position: 1, CallbackType: "begin_block"},
					{Position: 1, CallbackType: "end_block"},
				}
			}),
		},
		"duplicate callbacks": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{
					{Position: 1, CallbackType: "begin_block"},
					{Position: 2, CallbackType: "begin_block"},
				}
			}),
			expErr: true,
		},
		"unknown callback": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{{Position: 1, CallbackType: "unknown"}}
			}),
			expErr: true,
		},
		"empty callback": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{{Position: 1}}
			}),
			expErr: true,
		},
		"invalid callback position": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{{Position: math.MaxUint8 + 1, CallbackType: "begin_block"}}
			}),
			expErr: true,
		},
		"empty callback position": {
			src: TgradeContractDetailsFixture(t, func(d *TgradeContractDetails) {
				d.RegisteredCallbacks = []*RegisteredCallback{{CallbackType: "begin_block"}}
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
