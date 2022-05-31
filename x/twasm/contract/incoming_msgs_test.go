package contract

import (
	"encoding/json"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confio/tgrade/x/twasm/types"
)

func TestGetProposalContent(t *testing.T) {
	mySenderContractAddr := types.RandomAddress(t)

	ir := codectypes.NewInterfaceRegistry()
	ibcclienttypes.RegisterInterfaces(ir)
	ibctmtypes.RegisterInterfaces(ir)

	specs := map[string]struct {
		src               string
		expGovProposal    govtypes.Content
		expNotGovType     bool
		skipValidateBasic bool
	}{
		"text": {
			src:            `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"text":{}}}}`,
			expGovProposal: &govtypes.TextProposal{Title: "foo", Description: "bar"},
		},
		"register upgrade": {
			src: `{
"execute_gov_proposal": {
    "title": "myTitle", "description": "myDescription",
    "proposal": {
      "register_upgrade": {
		"name": "myUpgradeName",
        "height": 1,
        "info": "any information"
      }}}}`,
			expGovProposal: &upgradetypes.SoftwareUpgradeProposal{Title: "myTitle", Description: "myDescription", Plan: upgradetypes.Plan{
				Name:   "myUpgradeName",
				Time:   time.Time{},
				Height: 1,
				Info:   "any information",
			}},
		},
		"cancel upgrade": {
			src:            `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"cancel_upgrade":{}}}}`,
			expGovProposal: &upgradetypes.CancelSoftwareUpgradeProposal{Title: "foo", Description: "bar"},
		},
		"change param": {
			src: `{
"execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "change_params": [
        {"subspace": "mySubspace","key": "myKey","value": "myValue"},
        {"subspace": "myOtherSubspace", "key": "myOtherKey","value": "myOtherValue"}
      ]
    }}}`,
			expGovProposal: &proposaltypes.ParameterChangeProposal{Title: "foo", Description: "bar", Changes: []proposaltypes.ParamChange{
				{Subspace: "mySubspace", Key: "myKey", Value: "myValue"},
				{Subspace: "myOtherSubspace", Key: "myOtherKey", Value: "myOtherValue"},
			}},
		},
		"ibc client update": {
			src: `{
"execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "ibc_client_update": {
        "client_id": "myClientID",
        "header": {"type_url": "/ibc.lightclients.tendermint.v1.Header","value": "GgA="}
      }}}}`,
			expGovProposal: &ibcclienttypes.ClientUpdateProposal{
				Title:           "foo",
				Description:     "bar",
				SubjectClientId: "myClientID",
			},
			skipValidateBasic: true,
		},
		"promote to privileged contract": {
			src: `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"promote_to_privileged_contract":{"contract":"cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09"}}}}`,
			expGovProposal: &types.PromoteToPrivilegedContractProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
			},
		},
		"demote privileged contract": {
			src: `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"demote_privileged_contract":{"contract":"cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09"}}}}`,
			expGovProposal: &types.DemotePrivilegedContractProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
			},
		},
		"instantiate contract": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "instantiate_contract": {
        "admin": "cosmos1g6chpwdke3kz69x67jkak5y7gynneqm3nulfrd",
        "code_id": 1,
        "funds": [{"denom": "ALX", "amount": "2"},{"denom": "BLX","amount": "3"}],
        "msg": "e30=",
        "label": "testing"
      }}}}`,
			expGovProposal: &wasmtypes.InstantiateContractProposal{
				Title:       "foo",
				Description: "bar",
				RunAs:       mySenderContractAddr.String(),
				Admin:       "cosmos1g6chpwdke3kz69x67jkak5y7gynneqm3nulfrd",
				CodeID:      1,
				Label:       "testing",
				Msg:         []byte("{}"),
				Funds:       sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(2)), sdk.NewCoin("BLX", sdk.NewInt(3))),
			},
		},
		"store contract": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "store_code": {
        "wasm_byte_code": [0,0,0,0],
		"instantiate_permission": {"permission": "Everybody"}
      }}}}`,
			expGovProposal: &wasmtypes.StoreCodeProposal{
				Title:                 "foo",
				Description:           "bar",
				RunAs:                 mySenderContractAddr.String(),
				InstantiatePermission: &wasmtypes.AllowEverybody,
				WASMByteCode:          []byte{0, 0, 0, 0},
			},
		},
		"migrate contract": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "migrate_contract": {
        "code_id": 1,
		"contract": "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
        "migrate_msg": "e30="
      }}}}`,
			expGovProposal: &wasmtypes.MigrateContractProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
				CodeID:      1,
				Msg:         []byte("{}"),
			},
		},
		"set contract admin": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "set_contract_admin": {
		"contract": "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
        "new_admin": "cosmos1g6chpwdke3kz69x67jkak5y7gynneqm3nulfrd"
      }}}}`,
			expGovProposal: &wasmtypes.UpdateAdminProposal{
				Title:       "foo",
				Description: "bar",
				NewAdmin:    "cosmos1g6chpwdke3kz69x67jkak5y7gynneqm3nulfrd",
				Contract:    "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
			},
		},
		"clear contract admin": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "clear_contract_admin": {
		"contract": "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09"
      }}}}`,
			expGovProposal: &wasmtypes.ClearAdminProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "cosmos1vtg95naqtvf99hj8pe0s9aevy622vt0jmupc09",
			},
		},
		"pin codes": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "pin_codes": {
		"code_ids": [3,2,1]
      }}}}`,
			expGovProposal: &wasmtypes.PinCodesProposal{
				Title:       "foo",
				Description: "bar",
				CodeIDs:     []uint64{3, 2, 1},
			},
		},
		"unpin codes": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "unpin_codes": {
		"code_ids": [3,2,1]
      }}}}`,
			expGovProposal: &wasmtypes.UnpinCodesProposal{
				Title:       "foo",
				Description: "bar",
				CodeIDs:     []uint64{3, 2, 1},
			},
		},
		"unsupported proposal type": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "any_unknown": {
		"foo": "bar"
      }}}}`,
			expGovProposal: nil,
		},
		"no proposal type": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar"
}}`,
			expGovProposal: nil,
		},
		"no gov type": {
			src: `{
  "anything": {
    "foo": "bar"
}}`,
			expNotGovType: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var msg TgradeMsg
			require.NoError(t, msg.UnmarshalWithAny([]byte(spec.src), ir))
			gov := msg.ExecuteGovProposal
			if spec.expNotGovType {
				assert.Nil(t, gov)
				return
			}
			require.NotNil(t, gov)
			gotcontent := gov.GetProposalContent(mySenderContractAddr)
			exp, _ := json.Marshal(spec.expGovProposal)
			assert.Equal(t, spec.expGovProposal, gotcontent, string(exp))

			if spec.expGovProposal != nil && !spec.skipValidateBasic {
				assert.NoError(t, gotcontent.ValidateBasic())
			}
		})
	}
}

func TestConsensusParamsUpdateValidation(t *testing.T) {
	// some integers
	var one, two, three, four, five int64 = 1, 2, 3, 4, 5
	specs := map[string]struct {
		src    ConsensusParamsUpdate
		expErr *sdkerrors.Error
	}{
		"all good": {
			src: ConsensusParamsUpdate{
				Block: &BlockParams{
					MaxBytes: &one,
					MaxGas:   &two,
				},
				Evidence: &EvidenceParams{
					MaxAgeNumBlocks: &three,
					MaxAgeDuration:  &four,
					MaxBytes:        &five,
				},
			},
		},
		"empty msg": {
			src:    ConsensusParamsUpdate{},
			expErr: wasmtypes.ErrEmpty,
		},
		"block - empty": {
			src: ConsensusParamsUpdate{
				Block: &BlockParams{},
				Evidence: &EvidenceParams{
					MaxAgeNumBlocks: &three,
				},
			},
			expErr: wasmtypes.ErrEmpty,
		},
		"block - MaxBytes set": {
			src: ConsensusParamsUpdate{
				Block: &BlockParams{MaxBytes: &one},
			},
		},
		"block - MaxGas set": {
			src: ConsensusParamsUpdate{
				Block: &BlockParams{MaxGas: &one},
			},
		},
		"evidence - empty": {
			src: ConsensusParamsUpdate{
				Block: &BlockParams{
					MaxGas: &two,
				},
				Evidence: &EvidenceParams{},
			},
			expErr: wasmtypes.ErrEmpty,
		},
		"evidence - MaxAgeNumBlocks set": {
			src: ConsensusParamsUpdate{
				Evidence: &EvidenceParams{
					MaxAgeNumBlocks: &one,
				},
			},
		},
		"evidence - MaxAgeDuration set": {
			src: ConsensusParamsUpdate{
				Evidence: &EvidenceParams{
					MaxAgeDuration: &one,
				},
			},
		},
		"evidence - MaxBytes set": {
			src: ConsensusParamsUpdate{
				Evidence: &EvidenceParams{
					MaxBytes: &one,
				},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := spec.src.ValidateBasic()
			require.True(t, spec.expErr.Is(gotErr), "expected %v but got %#+v", spec.expErr, gotErr)
		})
	}
}
