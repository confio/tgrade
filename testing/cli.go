package testing

import (
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
	return NewTgradeCliX(t, sut.rpcAddr, sut.chainID, filepath.Join(sut.outputDir, "node0", "tgrade"), verbose)
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

func RequireTxSuccess(t *testing.T, got string) {
	code := gjson.Get(got, "code")
	rawLog := gjson.Get(got, "raw_log")
	require.Equal(t, int64(0), code.Int(), rawLog)
}
