package testing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TgradeCli wraps the command line interface
type TgradeCli struct {
	t           *testing.T
	nodeAddress string
	chainID     string
	homeDir     string
	Debug       bool
}

func NewTgradeCli(t *testing.T, sut *SystemUnderTest, verbose bool) *TgradeCli {
	return NewTgradeCliX(t, sut.rpcAddr, sut.chainID, filepath.Join(workDir, sut.outputDir), verbose)
}

func NewTgradeCliX(t *testing.T, nodeAddress string, chainID string, homeDir string, debug bool) *TgradeCli {
	return &TgradeCli{t: t, nodeAddress: nodeAddress, chainID: chainID, homeDir: homeDir, Debug: debug}
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
	cmd := exec.Command(locateExecutable("tgrade"), args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	require.NoError(c.t, err, string(out))
	return string(out)
}

func (c TgradeCli) withQueryFlags(args ...string) []string {
	args = append(args, "--output", "json")
	return c.withChainFlags(args...)
}

func (c TgradeCli) withTXFlags(args ...string) []string {
	args = append(args,
		"--broadcast-mode", "block",
		"--yes",
	)
	args = c.withKeyringFlags(args...)
	return c.withChainFlags(args...)
}

func (c TgradeCli) withKeyringFlags(args ...string) []string {
	return append(args,
		"--home", c.homeDir,
		"--keyring-backend", "test",
	)
}

func (c TgradeCli) withChainFlags(args ...string) []string {
	return append(args,
		"--node", c.nodeAddress,
		"--chain-id", c.chainID,
	)
}

// Execute send MsgExecute to a contract
func (c TgradeCli) Execute(contractAddr, msg, from string, amount ...sdk.Coin) string {
	cmd := []string{"tx", "wasm", "execute", contractAddr, msg, "--from", from}
	if len(amount) != 0 {
		cmd = append(cmd, "--amount", sdk.NewCoins(amount...).String())
	}
	return c.run(c.withTXFlags(cmd...))
}

// AddKey add key to default keyring. Returns address
func (c TgradeCli) AddKey(name string) string {
	cmd := c.withKeyringFlags("keys", "add", name, "--no-backup", "--output", "json")
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
func (c TgradeCli) FundAddress(t *testing.T, destAddr, amount string) string {
	require.NotEmpty(t, destAddr)
	require.NotEmpty(t, amount)
	cmd := []string{"tx", "send", defaultSrcAddr, destAddr, amount}
	return c.run(c.withTXFlags(cmd...))
}

// RequireTxSuccess require the received response to contain the success code
func RequireTxSuccess(t *testing.T, got string) {
	t.Helper()
	code := gjson.Get(got, "code")
	details := gjson.Get(got, "raw_log").String()
	if len(details) == 0 {
		details = got
	}
	require.Equal(t, int64(0), code.Int(), details)
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
