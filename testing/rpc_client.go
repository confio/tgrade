package testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	client "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

// RPCClient is a test helper to interact with a node via the RPC endpoint.
type RPCClient struct {
	client *client.HTTP
	t      *testing.T
}

func NewRPCCLient(t *testing.T, addr string) RPCClient {
	httpClient, err := client.New(addr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	return RPCClient{client: httpClient, t: t}
}

func (r RPCClient) Validators() []*tmtypes.Validator {
	v, err := r.client.Validators(context.Background(), nil, nil, nil)
	require.NoError(r.t, err)
	return v.Validators
}
