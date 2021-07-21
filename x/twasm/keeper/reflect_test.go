package keeper

import (
	"encoding/json"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"io/ioutil"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ReflectInitMsg is {}

// ReflectHandleMsg is used to encode handle messages
type ReflectHandleMsg struct {
	Reflect        *reflectPayload    `json:"reflect_msg,omitempty"`
	ReflectSubCall *reflectSubPayload `json:"reflect_sub_call,omitempty"`
	Change         *ownerPayload      `json:"change_owner,omitempty"`
}

type ownerPayload struct {
	Owner sdk.Address `json:"owner"`
}

type reflectPayload struct {
	Msgs []wasmvmtypes.CosmosMsg `json:"msgs"`
}

type reflectSubPayload struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}

// ReflectQueryMsg is used to encode query messages
type ReflectQueryMsg struct {
	Owner         *struct{}   `json:"owner,omitempty"`
	Chain         *ChainQuery `json:"chain,omitempty"`
	SubCallResult *SubCall    `json:"sub_call_result,omitempty"`
}

type ChainQuery struct {
	Request *wasmvmtypes.QueryRequest `json:"request,omitempty"`
}

type SubCall struct {
	ID uint64 `json:"id"`
}

type OwnerResponse struct {
	Owner string `json:"owner,omitempty"`
}

type ChainResponse struct {
	Data []byte `json:"data,omitempty"`
}


type HackatomExampleInitMsg struct {
	Verifier    sdk.AccAddress `json:"verifier"`
	Beneficiary sdk.AccAddress `json:"beneficiary"`
}

func (m HackatomExampleInitMsg) GetBytes(t *testing.T) []byte {
	initMsgBz, err := json.Marshal(m)
	require.NoError(t, err)
	return initMsgBz
}


func buildReflectQuery(t *testing.T, query *ReflectQueryMsg) []byte {
	bz, err := json.Marshal(query)
	require.NoError(t, err)
	return bz
}

func mustParse(t *testing.T, data []byte, res interface{}) {
	err := json.Unmarshal(data, res)
	require.NoError(t, err)
}

const ReflectFeatures = "staking,mask,stargate"

func TestReflectContractSend(t *testing.T) {
	ctx, keepers := CreateTestInput(t, false, ReflectFeatures)
	accKeeper, bankKeeper := keepers.AccountKeeper, keepers.BankKeeper
	keeper := wasmkeeper.NewDefaultPermissionKeeper(keepers.TWasmKeeper)

	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
	creator := createFakeFundedAccount(t, ctx, accKeeper, bankKeeper, deposit)
	bob := RandomAddress(t)

	// upload reflect code
	reflectCode, err := ioutil.ReadFile("./testdata/reflect.wasm")
	require.NoError(t, err)
	reflectID, err := keeper.Create(ctx, creator, reflectCode, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), reflectID)

	// upload hackatom escrow code
	escrowCode, err := ioutil.ReadFile("./testdata/hackatom.wasm")
	require.NoError(t, err)
	escrowID, err := keeper.Create(ctx, creator, escrowCode, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, uint64(2), escrowID)

	// creator instantiates a contract and gives it tokens
	reflectStart := sdk.NewCoins(sdk.NewInt64Coin("denom", 40000))
	reflectAddr, _, err := keeper.Instantiate(ctx, reflectID, creator, nil, []byte("{}"), "reflect contract 2", reflectStart)
	require.NoError(t, err)
	require.NotEmpty(t, reflectAddr)

	// now we set contract as verifier of an escrow
	initMsg := HackatomExampleInitMsg{
		Verifier:    reflectAddr,
		Beneficiary: bob,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)
	escrowStart := sdk.NewCoins(sdk.NewInt64Coin("denom", 25000))
	escrowAddr, _, err := keeper.Instantiate(ctx, escrowID, creator, nil, initMsgBz, "escrow contract 2", escrowStart)
	require.NoError(t, err)
	require.NotEmpty(t, escrowAddr)

	// let's make sure all balances make sense
	checkAccount(t, ctx, accKeeper, bankKeeper, creator, sdk.NewCoins(sdk.NewInt64Coin("denom", 35000))) // 100k - 40k - 25k
	checkAccount(t, ctx, accKeeper, bankKeeper, reflectAddr, reflectStart)
	checkAccount(t, ctx, accKeeper, bankKeeper, escrowAddr, escrowStart)
	checkAccount(t, ctx, accKeeper, bankKeeper, bob, nil)

	// now for the trick.... we reflect a message through the reflect to call the escrow
	// we also send an additional 14k tokens there.
	// this should reduce the reflect balance by 14k (to 26k)
	// this 14k is added to the escrow, then the entire balance is sent to bob (total: 39k)
	approveMsg := []byte(`{"release":{}}`)
	msgs := []wasmvmtypes.CosmosMsg{{
		Wasm: &wasmvmtypes.WasmMsg{
			Execute: &wasmvmtypes.ExecuteMsg{
				ContractAddr: escrowAddr.String(),
				Msg:          approveMsg,
				Send: []wasmvmtypes.Coin{{
					Denom:  "denom",
					Amount: "14000",
				}},
			},
		},
	}}
	reflectSend := ReflectHandleMsg{
		Reflect: &reflectPayload{
			Msgs: msgs,
		},
	}
	reflectSendBz, err := json.Marshal(reflectSend)
	require.NoError(t, err)
	_, err = keeper.Execute(ctx, reflectAddr, creator, reflectSendBz, nil)
	require.NoError(t, err)

	// did this work???
	checkAccount(t, ctx, accKeeper, bankKeeper, creator, sdk.NewCoins(sdk.NewInt64Coin("denom", 35000)))     // same as before
	checkAccount(t, ctx, accKeeper, bankKeeper, reflectAddr, sdk.NewCoins(sdk.NewInt64Coin("denom", 26000))) // 40k - 14k (from send)
	checkAccount(t, ctx, accKeeper, bankKeeper, escrowAddr, sdk.Coins{})                                     // emptied reserved
	checkAccount(t, ctx, accKeeper, bankKeeper, bob, sdk.NewCoins(sdk.NewInt64Coin("denom", 39000)))         // all escrow of 25k + 14k

}

//func TestReflectStargateQuery(t *testing.T) {
//	cdc := MakeEncodingConfig(t).Marshaler
//	ctx, keepers := CreateTestInput(t, false, ReflectFeatures, WithMessageEncoders(reflectEncoders(cdc)), WithQueryPlugins(reflectPlugins()))
//	accKeeper, keeper, bankKeeper := keepers.AccountKeeper, keepers.WasmKeeper, keepers.BankKeeper
//
//	funds := sdk.NewCoins(sdk.NewInt64Coin("denom", 320000))
//	contractStart := sdk.NewCoins(sdk.NewInt64Coin("denom", 40000))
//	expectedBalance := funds.Sub(contractStart)
//	creator := createFakeFundedAccount(t, ctx, accKeeper, bankKeeper, funds)
//
//	// upload code
//	reflectCode, err := ioutil.ReadFile("./testdata/reflect.wasm")
//	require.NoError(t, err)
//	codeID, err := keepers.ContractKeeper.Create(ctx, creator, reflectCode, "", "", nil)
//	require.NoError(t, err)
//	require.Equal(t, uint64(1), codeID)
//
//	// creator instantiates a contract and gives it tokens
//	contractAddr, _, err := keepers.ContractKeeper.Instantiate(ctx, codeID, creator, nil, []byte("{}"), "reflect contract 1", contractStart)
//	require.NoError(t, err)
//	require.NotEmpty(t, contractAddr)
//
//	// first, normal query for the bank balance (to make sure our query is proper)
//	bankQuery := wasmvmtypes.QueryRequest{
//		Bank: &wasmvmtypes.BankQuery{
//			AllBalances: &wasmvmtypes.AllBalancesQuery{
//				Address: creator.String(),
//			},
//		},
//	}
//	simpleQueryBz, err := json.Marshal(ReflectQueryMsg{
//		Chain: &ChainQuery{Request: &bankQuery},
//	})
//	require.NoError(t, err)
//	simpleRes, err := keeper.QuerySmart(ctx, contractAddr, simpleQueryBz)
//	require.NoError(t, err)
//	var simpleChain ChainResponse
//	mustParse(t, simpleRes, &simpleChain)
//	var simpleBalance wasmvmtypes.AllBalancesResponse
//	mustParse(t, simpleChain.Data, &simpleBalance)
//	require.Equal(t, len(expectedBalance), len(simpleBalance.Amount))
//	assert.Equal(t, simpleBalance.Amount[0].Amount, expectedBalance[0].Amount.String())
//	assert.Equal(t, simpleBalance.Amount[0].Denom, expectedBalance[0].Denom)
//
//	// now, try to build a protobuf query
//	protoQuery := banktypes.QueryAllBalancesRequest{
//		Address: creator.String(),
//	}
//	protoQueryBin, err := proto.Marshal(&protoQuery)
//	protoRequest := wasmvmtypes.QueryRequest{
//		Stargate: &wasmvmtypes.StargateQuery{
//			Path: "/cosmos.bank.v1beta1.Query/AllBalances",
//			Data: protoQueryBin,
//		},
//	}
//	protoQueryBz, err := json.Marshal(ReflectQueryMsg{
//		Chain: &ChainQuery{Request: &protoRequest},
//	})
//	require.NoError(t, err)
//
//	// make a query on the chain
//	protoRes, err := keeper.QuerySmart(ctx, contractAddr, protoQueryBz)
//	require.NoError(t, err)
//	var protoChain ChainResponse
//	mustParse(t, protoRes, &protoChain)
//
//	// unmarshal raw protobuf response
//	var protoResult banktypes.QueryAllBalancesResponse
//	err = proto.Unmarshal(protoChain.Data, &protoResult)
//	require.NoError(t, err)
//	assert.Equal(t, expectedBalance, protoResult.Balances)
//}
//
//type reflectState struct {
//	Owner string `json:"owner"`
//}
//
//func TestMaskReflectWasmQueries(t *testing.T) {
//	cdc := MakeEncodingConfig(t).Marshaler
//	ctx, keepers := CreateTestInput(t, false, ReflectFeatures, WithMessageEncoders(reflectEncoders(cdc)), WithQueryPlugins(reflectPlugins()))
//	accKeeper, keeper := keepers.AccountKeeper, keepers.WasmKeeper
//
//	deposit := sdk.NewCoins(sdk.NewInt64Coin("denom", 100000))
//	creator := createFakeFundedAccount(t, ctx, accKeeper, keepers.BankKeeper, deposit)
//
//	// upload reflect code
//	reflectCode, err := ioutil.ReadFile("./testdata/reflect.wasm")
//	require.NoError(t, err)
//	reflectID, err := keepers.ContractKeeper.Create(ctx, creator, reflectCode, "", "", nil)
//	require.NoError(t, err)
//	require.Equal(t, uint64(1), reflectID)
//
//	// creator instantiates a contract and gives it tokens
//	reflectStart := sdk.NewCoins(sdk.NewInt64Coin("denom", 40000))
//	reflectAddr, _, err := keepers.ContractKeeper.Instantiate(ctx, reflectID, creator, nil, []byte("{}"), "reflect contract 2", reflectStart)
//	require.NoError(t, err)
//	require.NotEmpty(t, reflectAddr)
//
//	// for control, let's make some queries directly on the reflect
//	ownerQuery := buildReflectQuery(t, &ReflectQueryMsg{Owner: &struct{}{}})
//	res, err := keeper.QuerySmart(ctx, reflectAddr, ownerQuery)
//	require.NoError(t, err)
//	var ownerRes OwnerResponse
//	mustParse(t, res, &ownerRes)
//	require.Equal(t, ownerRes.Owner, creator.String())
//
//	// and a raw query: cosmwasm_storage::Singleton uses 2 byte big-endian length-prefixed to store data
//	configKey := append([]byte{0, 6}, []byte("config")...)
//	raw := keeper.QueryRaw(ctx, reflectAddr, configKey)
//	var stateRes reflectState
//	mustParse(t, raw, &stateRes)
//	require.Equal(t, stateRes.Owner, creator.String())
//
//	// now, let's reflect a smart query into the x/wasm handlers and see if we get the same result
//	reflectOwnerQuery := ReflectQueryMsg{Chain: &ChainQuery{Request: &wasmvmtypes.QueryRequest{Wasm: &wasmvmtypes.WasmQuery{
//		Smart: &wasmvmtypes.SmartQuery{
//			ContractAddr: reflectAddr.String(),
//			Msg:          ownerQuery,
//		},
//	}}}}
//	reflectOwnerBin := buildReflectQuery(t, &reflectOwnerQuery)
//	res, err = keeper.QuerySmart(ctx, reflectAddr, reflectOwnerBin)
//	require.NoError(t, err)
//	// first we pull out the data from chain response, before parsing the original response
//	var reflectRes ChainResponse
//	mustParse(t, res, &reflectRes)
//	var reflectOwnerRes OwnerResponse
//	mustParse(t, reflectRes.Data, &reflectOwnerRes)
//	require.Equal(t, reflectOwnerRes.Owner, creator.String())
//
//	// and with queryRaw
//	reflectStateQuery := ReflectQueryMsg{Chain: &ChainQuery{Request: &wasmvmtypes.QueryRequest{Wasm: &wasmvmtypes.WasmQuery{
//		Raw: &wasmvmtypes.RawQuery{
//			ContractAddr: reflectAddr.String(),
//			Key:          configKey,
//		},
//	}}}}
//	reflectStateBin := buildReflectQuery(t, &reflectStateQuery)
//	res, err = keeper.QuerySmart(ctx, reflectAddr, reflectStateBin)
//	require.NoError(t, err)
//	// first we pull out the data from chain response, before parsing the original response
//	var reflectRawRes ChainResponse
//	mustParse(t, res, &reflectRawRes)
//	// now, with the raw data, we can parse it into state
//	var reflectStateRes reflectState
//	mustParse(t, reflectRawRes.Data, &reflectStateRes)
//	require.Equal(t, reflectStateRes.Owner, creator.String())
//}

func checkAccount(t *testing.T, ctx sdk.Context, accKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper, addr sdk.AccAddress, expected sdk.Coins) {
	acct := accKeeper.GetAccount(ctx, addr)
	if expected == nil {
		assert.Nil(t, acct)
	} else {
		assert.NotNil(t, acct)
		if expected.Empty() {
			// there is confusion between nil and empty slice... let's just treat them the same
			assert.True(t, bankKeeper.GetAllBalances(ctx, acct.GetAddress()).Empty())
		} else {
			assert.Equal(t, bankKeeper.GetAllBalances(ctx, acct.GetAddress()), expected)
		}
	}
}
