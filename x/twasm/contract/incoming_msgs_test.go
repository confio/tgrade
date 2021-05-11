package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetProposalContent(t *testing.T) {
	cs, err := clienttypes.PackClientState(&ibctmtypes.ClientState{})
	require.NoError(t, err)

	require.NoError(t, err)
	anyType.ClearCachedValue()
	jm := &jsonpb.Marshaler{OrigName: false, EmitDefaults: true, AnyResolver: nil}
	var buf bytes.Buffer
	require.NoError(t, jm.Marshal(&buf, anyType))
	require.NoError(t, err)
	xxx := fmt.Sprintf(`{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"register_upgrade":{"height":1, "info":"any information", "upgraded_client_state": %s}}}}`, buf.String())
	t.Log("___ ", xxx)
	specs := map[string]struct {
		src string
		exp govtypes.Content
	}{
		"text": {
			src: `{"execute_gov_proposal":{"title":"foo", "description":"bar", "proposal":{"text":{}}}}`,
			exp: &govtypes.TextProposal{Title: "foo", Description: "bar"},
		},
		"register upgrade": {
			src: xxx,
			exp: &upgradetypes.SoftwareUpgradeProposal{Title: "foo", Description: "bar", Plan: upgradetypes.Plan{
				Name:                "",
				Time:                time.Time{},
				Height:              1,
				Info:                "any information",
				UpgradedClientState: anyType,
			}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var msg TgradeMsg
			require.NoError(t, json.Unmarshal([]byte(spec.src), &msg))
			gov := msg.ExecuteGovProposal
			require.NotNil(t, gov)
			assert.Equal(t, spec.exp, gov.GetProposalContent())
		})
	}

}
