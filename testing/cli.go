package testing

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/confio/tgrade/app"
)

// TgradeCli wraps the command line interface
type TgradeCli struct {
	t             *testing.T
	nodeAddress   string
	chainID       string
	homeDir       string
	Debug         bool
	amino         *codec.LegacyAmino
	assertErrorFn func(t require.TestingT, err error, msgAndArgs ...interface{})
}

func NewTgradeCli(t *testing.T, sut *SystemUnderTest, verbose bool) *TgradeCli {
	return NewTgradeCliX(t, sut.rpcAddr, sut.chainID, filepath.Join(workDir, sut.outputDir), verbose)
}

func NewTgradeCliX(t *testing.T, nodeAddress string, chainID string, homeDir string, debug bool) *TgradeCli {
	return &TgradeCli{
		t:             t,
		nodeAddress:   nodeAddress,
		chainID:       chainID,
		homeDir:       homeDir,
		Debug:         debug,
		amino:         app.MakeEncodingConfig().Amino,
		assertErrorFn: require.NoError,
	}
}

// RunErrorAssert is custom type that is satisfies by testify matchers as well
type RunErrorAssert func(t require.TestingT, err error, msgAndArgs ...interface{})

// WithRunErrorMatcher assert function to ensure run command error value
func (c TgradeCli) WithRunErrorMatcher(f RunErrorAssert) TgradeCli {
	return TgradeCli{
		t:             c.t,
		nodeAddress:   c.nodeAddress,
		chainID:       c.chainID,
		homeDir:       c.homeDir,
		Debug:         c.Debug,
		amino:         c.amino,
		assertErrorFn: f,
	}
}

func (c TgradeCli) WithNodeAddress(addr string) TgradeCli {
	return TgradeCli{
		t:             c.t,
		nodeAddress:   addr,
		chainID:       c.chainID,
		homeDir:       c.homeDir,
		Debug:         c.Debug,
		amino:         c.amino,
		assertErrorFn: c.assertErrorFn,
	}
}

func (c TgradeCli) CustomCommand(args ...string) string {
	args = c.withTXFlags(args...)
	return c.run(args)
}

func (c TgradeCli) Keys(args ...string) string {
	args = c.withKeyringFlags(args...)
	return c.run(args)
}

func (c TgradeCli) CustomQuery(args ...string) string {
	args = c.withQueryFlags(args...)
	return c.run(args)
}

func (c TgradeCli) run(args []string) string {
	if c.Debug {
		c.t.Logf("+++ running `tgrade %s`", strings.Join(args, " "))
	}
	gotOut, gotErr := func() (out []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("recovered from panic: %v", r)
			}
		}()
		cmd := exec.Command(locateExecutable("tgrade"), args...) //nolint:gosec
		cmd.Dir = workDir
		return cmd.CombinedOutput()
	}()
	c.assertErrorFn(c.t, gotErr, string(gotOut))
	return string(gotOut)
}

func (c TgradeCli) withQueryFlags(args ...string) []string {
	args = append(args, "--output", "json")
	return c.withChainFlags(args...)
}

func (c TgradeCli) withTXFlags(args ...string) []string {
	args = append(args,
		"--broadcast-mode", "block",
		"--output", "json",
		"--yes",
	)
	args = c.withKeyringFlags(args...)
	return c.withChainFlags(args...)
}

func (c TgradeCli) withKeyringFlags(args ...string) []string {
	r := append(args,
		"--home", c.homeDir,
		"--keyring-backend", "test",
	)
	for _, v := range args {
		if v == "-a" || v == "--address" { // show address only
			return r
		}
	}
	return append(r, "--output", "json")
}

func (c TgradeCli) withChainFlags(args ...string) []string {
	return append(args,
		"--node", c.nodeAddress,
		"--chain-id", c.chainID,
	)
}

// Execute send MsgExecute to a contract
func (c TgradeCli) Execute(contractAddr, msg, from string, args ...string) string {
	cmd := []string{"tx", "wasm", "execute", contractAddr, msg, "--from", from}
	return c.run(c.withTXFlags(append(cmd, args...)...))
}

// AddKey add key to default keyring. Returns address
func (c TgradeCli) AddKey(name string) string {
	cmd := c.withKeyringFlags("keys", "add", name, "--no-backup")
	out := c.run(cmd)
	addr := gjson.Get(out, "address").String()
	require.NotEmpty(c.t, addr, "got %q", out)
	return addr
}

// GetKeyAddr returns address
func (c TgradeCli) GetKeyAddr(name string) string {
	cmd := c.withKeyringFlags("keys", "show", name, "-a")
	out := c.run(cmd)
	addr := strings.Trim(out, "\n")
	require.NotEmpty(c.t, addr, "got %q", out)
	return addr
}

const defaultSrcAddr = "node0"

// FundAddress sends the token amount to the destination address
func (c TgradeCli) FundAddress(destAddr, amount string) string {
	require.NotEmpty(c.t, destAddr)
	require.NotEmpty(c.t, amount)
	cmd := []string{"tx", "bank", "send", defaultSrcAddr, destAddr, amount}
	rsp := c.run(c.withTXFlags(cmd...))
	RequireTxSuccess(c.t, rsp)
	return rsp
}

// StoreWasm uploads a wasm contract to the chain. Returns code id
func (c TgradeCli) StoreWasm(file string, args ...string) int {
	if len(args) == 0 {
		args = []string{"--from=" + defaultSrcAddr, "--gas=2500000"}
	}
	rsp := c.run(c.withTXFlags(append([]string{"tx", "wasm", "store", file}, args...)...))
	RequireTxSuccess(c.t, rsp)
	codeID := gjson.Get(rsp, "logs.#.events.#.attributes.#(key=code_id).value").Array()[0].Array()[0].Int()
	require.NotEmpty(c.t, codeID)
	return int(codeID)
}

// InstantiateWasm create a new contract instance. returns contract address
func (c TgradeCli) InstantiateWasm(codeID int, initMsg string, args ...string) string {
	if len(args) == 0 {
		args = []string{"--label=testing", "--from=" + defaultSrcAddr, "--no-admin"}
	}
	rsp := c.run(c.withTXFlags(append([]string{"tx", "wasm", "instantiate", strconv.Itoa(codeID), initMsg}, args...)...))
	RequireTxSuccess(c.t, rsp)
	addr := gjson.Get(rsp, "logs.#.events.#.attributes.#(key=_contract_address).value").Array()[0].Array()[0].String()
	require.NotEmpty(c.t, addr)
	return addr
}

// QuerySmart run smart contract query
func (c TgradeCli) QuerySmart(contractAddr, msg string, args ...string) string {
	cmd := append([]string{"q", "wasm", "contract-state", "smart", contractAddr, msg}, args...)
	args = c.withQueryFlags(cmd...)
	return c.run(args)
}

// QueryBalances queries all balances for an account. Returns json response
// Example:`{"balances":[{"denom":"node0token","amount":"1000000000"},{"denom":"utgd","amount":"400000003"}],"pagination":{}}`
func (c TgradeCli) QueryBalances(addr string) string {
	return c.CustomQuery("q", "bank", "balances", addr)
}

// QueryBalance returns balance amount for given denom.
// 0 when not found
func (c TgradeCli) QueryBalance(addr, denom string) int64 {
	raw := c.CustomQuery("q", "bank", "balances", addr, "--denom="+denom)
	require.Contains(c.t, raw, "amount", raw)
	return gjson.Get(raw, "amount").Int()
}

// QueryTotalSupply returns total amount of tokens for a given denom.
// 0 when not found
func (c TgradeCli) QueryTotalSupply(denom string) int64 {
	raw := c.CustomQuery("q", "bank", "total", "--denom="+denom)
	require.Contains(c.t, raw, "amount", raw)
	return gjson.Get(raw, "amount").Int()
}

// QueryValidator queries the validator for the given operator address. Returns json response
func (c TgradeCli) QueryValidator(addr string) string {
	return c.CustomQuery("q", "poe", "validator", addr)
}

// QueryValidatorRewards queries the validator rewards for the given operator address
func (c TgradeCli) QueryValidatorRewards(addr string) sdk.DecCoin {
	raw := c.CustomQuery("q", "poe", "validator-reward", addr)
	require.NotEmpty(c.t, raw)

	r := gjson.Get(raw, "reward")
	amount, err := sdk.NewDecFromStr(gjson.Get(r.Raw, "amount").String())
	require.NoError(c.t, err)
	denom := gjson.Get(r.Raw, "denom").String()
	return sdk.NewDecCoinFromDec(denom, amount)
}

func (c TgradeCli) GetTendermintValidatorSet() rpc.ResultValidatorsOutput {
	args := []string{"q", "tendermint-validator-set"}
	got := c.run(c.withQueryFlags(args...))

	var res rpc.ResultValidatorsOutput
	require.NoError(c.t, c.amino.UnmarshalJSON([]byte(got), &res), got)
	return res
}

// GetPoEContractAddress query the PoE contract address
func (c TgradeCli) GetPoEContractAddress(v string) string {
	qRes := c.CustomQuery("q", "poe", "contract-address", v)
	addr := gjson.Get(qRes, "address").String()
	require.NotEmpty(c.t, addr, "got %q", addr)
	return addr
}

// IsInTendermintValset returns true when the giben pub key is in the current active tendermint validator set
func (c TgradeCli) IsInTendermintValset(valPubKey cryptotypes.PubKey) (rpc.ResultValidatorsOutput, bool) {
	valResult := c.GetTendermintValidatorSet()
	var found bool
	for _, v := range valResult.Validators {
		if v.PubKey.Equals(valPubKey) {
			found = true
			break
		}
	}
	return valResult, found
}

// RequireTxSuccess require the received response to contain the success code
func RequireTxSuccess(t *testing.T, got string) {
	t.Helper()
	code := gjson.Get(got, "code")
	details := gjson.Get(got, "raw_log").String()
	if len(details) == 0 {
		details = got
	}
	require.Equal(t, int64(0), code.Int(), "non success tx code : %s", details)
}

// RequireTxFailure require the received response to contain any failure code and the passed msgsgs
func RequireTxFailure(t *testing.T, got string, containsMsgs ...string) {
	t.Helper()
	code := gjson.Get(got, "code")
	rawLog := gjson.Get(got, "raw_log").String()
	require.NotEqual(t, int64(0), code.Int(), rawLog)
	for _, msg := range containsMsgs {
		require.Contains(t, rawLog, msg)
	}
}

var (
	// ErrOutOfGasMatcher requires error with out of gas message
	ErrOutOfGasMatcher RunErrorAssert = func(t require.TestingT, err error, args ...interface{}) {
		const oogMsg = "out of gas"
		expErrWithMsg(t, err, args, oogMsg)
	}
	// ErrTimeoutMatcher requires time out message
	ErrTimeoutMatcher RunErrorAssert = func(t require.TestingT, err error, args ...interface{}) {
		const expMsg = "timed out waiting for tx to be included in a block"
		expErrWithMsg(t, err, args, expMsg)
	}
	// ErrPostFailedMatcher requires post failed
	ErrPostFailedMatcher RunErrorAssert = func(t require.TestingT, err error, args ...interface{}) {
		const expMsg = "post failed"
		expErrWithMsg(t, err, args, expMsg)
	}
	// ErrInvalidQuery requires smart query request failed
	ErrInvalidQuery RunErrorAssert = func(t require.TestingT, err error, args ...interface{}) {
		const expMsg = "query wasm contract failed"
		expErrWithMsg(t, err, args, expMsg)
	}
)

func expErrWithMsg(t require.TestingT, err error, args []interface{}, expMsg string) {
	require.Error(t, err, args)
	var found bool
	for _, v := range args {
		if strings.Contains(fmt.Sprintf("%s", v), expMsg) {
			found = true
			break
		}
	}
	require.True(t, found, "expected %q but got: %s", expMsg, args)
}
