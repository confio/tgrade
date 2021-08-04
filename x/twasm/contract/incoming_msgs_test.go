package contract

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/confio/tgrade/x/twasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetProposalContent(t *testing.T) {
	ir := codectypes.NewInterfaceRegistry()
	clienttypes.RegisterInterfaces(ir)
	ibctmtypes.RegisterInterfaces(ir)

	ib, err := ibcclienttypes.PackHeader(&ibctmtypes.Header{})
	require.NoError(t, err)
	cs, err := clienttypes.PackClientState(&ibctmtypes.ClientState{})
	require.NoError(t, err)

	specs := map[string]struct {
		src            string
		expGovProposal govtypes.Content
		expNotGovType  bool
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
        "info": "any information",
        "upgraded_client_state": {
          "type_url": "/ibc.lightclients.tendermint.v1.ClientState", "value": "EgAaACIAKgAyADoA"
        }
      }}}}`,
			expGovProposal: &upgradetypes.SoftwareUpgradeProposal{Title: "myTitle", Description: "myDescription", Plan: upgradetypes.Plan{
				Name:                "myUpgradeName",
				Time:                time.Time{},
				Height:              1,
				Info:                "any information",
				UpgradedClientState: cs,
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
				Title:       "foo",
				Description: "bar",
				ClientId:    "myClientID",
				Header:      ib,
			},
		},
		"promote to privileged contract": {
			src: `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"promote_to_privileged_contract":{"contract":"myContractAddress"}}}}`,
			expGovProposal: &types.PromoteToPrivilegedContractProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "myContractAddress",
			},
		},
		"demote privileged contract": {
			src: `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"demote_privileged_contract":{"contract":"myContractAddress"}}}}`,
			expGovProposal: &types.DemotePrivilegedContractProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "myContractAddress",
			},
		},
		"instantiate contract": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "instantiate_contract": {
        "admin": "myAdminAddress",
        "code_id": 1,
        "funds": [{"denom": "ALX", "amount": "2"},{"denom": "BLX","amount": "3"}],
        "msg": "e30=",
        "label": "testing",
        "run_as": "myRunAsAddress"
      }}}}`,
			expGovProposal: &wasmtypes.InstantiateContractProposal{
				Title:       "foo",
				Description: "bar",
				RunAs:       "myRunAsAddress",
				Admin:       "myAdminAddress",
				CodeID:      1,
				Label:       "testing",
				Msg:         []byte("{}"),
				Funds:       sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(2)), sdk.NewCoin("BLX", sdk.NewInt(3))),
			},
		},
		"migrate contract": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "migrate_contract": {
        "code_id": 1,
		"contract": "myContractAddr",
        "msg": "e30=",
        "run_as": "myRunAsAddress"
      }}}}`,
			expGovProposal: &wasmtypes.MigrateContractProposal{
				Title:       "foo",
				Description: "bar",
				RunAs:       "myRunAsAddress",
				Contract:    "myContractAddr",
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
		"contract": "myContractAddr",
        "new_admin": "myNewAdminAddress"
      }}}}`,
			expGovProposal: &wasmtypes.UpdateAdminProposal{
				Title:       "foo",
				Description: "bar",
				NewAdmin:    "myNewAdminAddress",
				Contract:    "myContractAddr",
			},
		},
		"clear contract admin": {
			src: `{
  "execute_gov_proposal": {
    "title": "foo", "description": "bar",
    "proposal": {
      "clear_contract_admin": {
		"contract": "myContractAddr"
      }}}}`,
			expGovProposal: &wasmtypes.ClearAdminProposal{
				Title:       "foo",
				Description: "bar",
				Contract:    "myContractAddr",
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
			assert.Equal(t, spec.expGovProposal, gov.GetProposalContent())

		})
	}
}
