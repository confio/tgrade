package testing

import (
	"bufio"
	"bytes"
	"container/ring"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	client "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var workDir string

// SystemUnderTest blockchain provisioning
type SystemUnderTest struct {
	blockListener EventListener
	currentHeight int64
	chainID       string
	outputDir     string
	blockTime     time.Duration
	rpcAddr       string
	nodesCount    int
	minGasPrice   string
	cleanupFn     []CleanupFn
	outBuff       *ring.Ring
	errBuff       *ring.Ring
	out           io.Writer
	verbose       bool
}

func NewSystemUnderTest(verbose bool) *SystemUnderTest {
	return &SystemUnderTest{
		chainID:    "testing",
		outputDir:  "./testnet",
		blockTime:  1500 * time.Millisecond,
		rpcAddr:    "tcp://localhost:26657",
		nodesCount: 1,
		outBuff:    ring.New(100),
		errBuff:    ring.New(100),
		out:        os.Stdout,
		verbose:    verbose,
	}
}

func (s SystemUnderTest) SetupChain() {
	s.Log("Setup chain")
	args := []string{
		"testnet",
		"--chain-id=" + s.chainID,
		"--output-dir=" + s.outputDir,
		"--v=" + strconv.Itoa(s.nodesCount),
		"--keyring-backend=test",
		"--commit-timeout=" + s.blockTime.String(),
		"--minimum-gas-prices=" + s.minGasPrice,
		"--starting-ip-address", "", // empty to use host systems
	}
	cmd := exec.Command(
		locateExecutable("tgrade"),
		args...,
	)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	s.Log(string(out))

	// backup genesis
	src := filepath.Join(workDir, s.nodePath(0), "config", "genesis.json")
	dest := filepath.Join(workDir, s.nodePath(0), "config", "genesis.json.orig")
	if _, err := copyFile(src, dest); err != nil {
		panic(fmt.Sprintf("copy failed :%#+v", err))
	}
}

func (s SystemUnderTest) StartChain(t *testing.T) {
	s.Log("Start chain")
	s.forEachNodesExecAsync(t, "start", "--trace", "--log_level=info")

	s.awaitChainUp(t)

	t.Log("Start new block listener")
	s.blockListener = NewEventListener(t, s.rpcAddr)
	s.cleanupFn = append(s.cleanupFn,
		s.blockListener.Subscribe("tm.event='NewBlock'", func(e ctypes.ResultEvent) (more bool) {
			newBlock, ok := e.Data.(types.EventDataNewBlock)
			require.True(t, ok, "unexpected type %T", e.Data)
			atomic.StoreInt64(&s.currentHeight, newBlock.Block.Height)
			return true
		}),
	)
}

func (s *SystemUnderTest) watchLogs(cmd *exec.Cmd) {
	errReader, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}
	go appendToBuf(errReader, s.errBuff)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}
	go appendToBuf(outReader, s.outBuff)
}

func appendToBuf(r io.ReadCloser, b *ring.Ring) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b.Value = scanner.Text()
		b = b.Next()
	}
}

// awaitChainUp ensures the chain is running
func (s SystemUnderTest) awaitChainUp(t *testing.T) {
	t.Log("Await chain starts")
	timeout := defaultWaitTime
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()

	started := make(chan struct{})
	go func() { // query for a non empty block on status page
		t.Logf("Checking node status: %s\n", s.rpcAddr)
		for {
			con, err := client.New(s.rpcAddr, "/websocket")
			if err != nil || con.Start() != nil {
				time.Sleep(time.Second)
				continue
			}
			result, err := con.Status(ctx)
			if err != nil || result.SyncInfo.LatestBlockHeight < 1 {
				con.Stop()
				continue
			}
			t.Logf("Node started. Current block: %d\n", result.SyncInfo.LatestBlockHeight)
			con.Stop()
			started <- struct{}{}
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	case <-time.NewTimer(timeout).C:
		t.Fatalf("timeout waiting for chain start: %s", timeout)
	}
}

// StopChain stops the system under test and executes all registered cleanup callbacks
func (s SystemUnderTest) StopChain() {
	s.Log("Stop chain")
	for _, c := range s.cleanupFn {
		c()
	}
	s.cleanupFn = nil
	cmd := exec.Command(locateExecutable("pkill"), "-15", "tgrade")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to stop chain: %s\n", err)
	}
	s.Log(string(out))
}

// PrintBuffer prints the chain logs to the console
func (s SystemUnderTest) PrintBuffer() {
	s.outBuff.Do(func(v interface{}) {
		if v != nil {
			fmt.Fprintf(s.out, "out> %s\n", v)
		}
	})
	fmt.Fprint(s.out, "8< chain err -----------------------------------------\n")
	s.errBuff.Do(func(v interface{}) {
		if v != nil {
			fmt.Fprintf(s.out, "err> %s\n", v)
		}
	})
}

// BuildNewBinary builds and installs new tgrade binary
func (s SystemUnderTest) BuildNewBinary() {
	s.Log("Install binaries\n")
	makePath := locateExecutable("make")
	cmd := exec.Command(makePath, "clean", "install")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#v : output: %s", err, string(out)))
	}
}

// AwaitNextBlock is a first class function that any caller can use to ensure a new block was minted
func (s SystemUnderTest) AwaitNextBlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for start := atomic.LoadInt64(&s.currentHeight); atomic.LoadInt64(&s.currentHeight) > start; {
			time.Sleep(s.blockTime)
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.NewTimer(s.blockTime * 2).C:
		t.Fatalf("Timeout - no block within %s", s.blockTime*2)
	}
}

// ResetChain stops and clears all nodes state via 'unsafe-reset-all'
func (s SystemUnderTest) ResetChain(t *testing.T) {
	t.Log("ResetChain chain")
	s.StopChain()
	restoreOriginalGenesis(t, s)
	cmd := exec.Command(locateExecutable("pkill"), "tgrade")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to kill process: %s", err)
	}
	s.Log(string(out))

	// reset all nodes
	s.ForEachNodeExecAndWait(t, []string{"unsafe-reset-all"})
}

// ModifyGenesis executes the commands to modify the genesis
func (s SystemUnderTest) ModifyGenesis(t *testing.T, cmds ...[]string) {
	s.ForEachNodeExecAndWait(t, cmds...)
}

// SetGenesis copy genesis file to all nodes
func (s SystemUnderTest) SetGenesis(t *testing.T, srcPath string) {
	in, err := os.Open(srcPath)
	require.NoError(t, err)
	defer in.Close()
	var buf bytes.Buffer

	_, err = io.Copy(&buf, in)
	require.NoError(t, err)

	s.withEachNodeHome(func(i int, home string) {
		out, err := os.Create(filepath.Join(workDir, home, "config", "genesis.json"))
		require.NoError(t, err)
		defer out.Close()

		_, err = io.Copy(out, bytes.NewReader(buf.Bytes()))
		require.NoError(t, err)
		require.NoError(t, out.Close())
	})
}

// ForEachNodeExecAndWait runs the given tgrade commands for all cluster nodes synchronously
func (s SystemUnderTest) ForEachNodeExecAndWait(t *testing.T, cmds ...[]string) {
	s.withEachNodeHome(func(i int, home string) {
		for _, args := range cmds {
			args = append(args, "--home", home)
			s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
			cmd := exec.Command(
				locateExecutable("tgrade"),
				args...,
			)
			cmd.Dir = workDir
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "node %d: %s", i, string(out))
			s.Logf("Result: %s\n", string(out))
		}
	})
}

// forEachNodesExecAsync runs the given tgrade command for all cluster nodes and returns without waiting
func (s SystemUnderTest) forEachNodesExecAsync(t *testing.T, args ...string) []func() error {
	r := make([]func() error, s.nodesCount)
	s.withEachNodeHome(func(i int, home string) {
		args = append(args, "--home", home)
		s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
		cmd := exec.Command(
			locateExecutable("tgrade"),
			args...,
		)
		cmd.Dir = workDir
		s.watchLogs(cmd)
		require.NoError(t, cmd.Start(), "node %d", i)
		r[i] = cmd.Wait
	})
	return r
}

func (s SystemUnderTest) withEachNodeHome(cb func(i int, home string)) {
	for i := 0; i < s.nodesCount; i++ {
		cb(i, s.nodePath(i))
	}
}

func (s SystemUnderTest) nodePath(i int) string {
	return fmt.Sprintf("%s/node%d/tgrade", s.outputDir, i)
}

func (s SystemUnderTest) Log(msg string) {
	if s.verbose {
		fmt.Fprint(s.out, msg)
	}
}

func (s SystemUnderTest) Logf(msg string, args ...interface{}) {
	s.Log(fmt.Sprintf(msg, args...))
}

// locateExecutable looks up the binary on the OS path.
func locateExecutable(file string) string {
	path, err := exec.LookPath(file)
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#v", err))
	}
	if path == "" {
		panic(fmt.Sprintf("%q not founc", file))
	}
	return path
}

// EventListener watches for events on the chain
type EventListener struct {
	t      *testing.T
	client *client.HTTP
}

// NewEventListener event listener
func NewEventListener(t *testing.T, rpcAddr string) EventListener {
	httpClient, err := client.New(rpcAddr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	return EventListener{client: httpClient, t: t}
}

var defaultWaitTime = 30 * time.Second

type (
	CleanupFn     func()
	EventConsumer func(e ctypes.ResultEvent) (more bool)
)

// Subscribe to receive events for a topic.
// For query syntax See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
func (l EventListener) Subscribe(query string, cb EventConsumer) func() {
	ctx, done := context.WithCancel(context.Background())
	eventsChan, err := l.client.WSEvents.Subscribe(ctx, "testing", query)
	require.NoError(l.t, err)
	cleanup := func() {
		ctx, _ := context.WithTimeout(ctx, defaultWaitTime)
		go l.client.WSEvents.Unsubscribe(ctx, "testing", query)
		done()
	}
	go func() {
		for {
			select {
			case e := <-eventsChan:
				if !cb(e) {
					return
				}
			}
		}
	}()
	return cleanup
}

// AwaitQuery waits for single result or timeout
func (l EventListener) AwaitQuery(query string) *ctypes.ResultEvent {
	c, result := CapturingEventConsumer()
	l.Subscribe(query, TimeoutConsumer(l.t, defaultWaitTime, c))
	return result
}

// TimeoutConsumer is an event consumer decorator with a max wait time. Panics when wait time exceeded without
// a result returned
func TimeoutConsumer(t *testing.T, waitTime time.Duration, next EventConsumer) EventConsumer {
	ctx, done := context.WithCancel(context.Background())
	timeout := time.NewTimer(waitTime)
	go func() {
		select {
		case <-ctx.Done():
		case <-timeout.C:
			t.Fatalf("Timeout waiting for new events %s", waitTime)
		}
	}()
	return func(e ctypes.ResultEvent) (more bool) {
		timeout.Reset(waitTime)
		result := next(e)
		if !result {
			done()
		}
		return result
	}
}

// CapturingEventConsumer consumes one event. No timeout
func CapturingEventConsumer() (EventConsumer, *ctypes.ResultEvent) {
	var result ctypes.ResultEvent
	return func(e ctypes.ResultEvent) (more bool) {
		return false
	}, &result
}

// restoreOriginalGenesis replace nodes genesis by the one created on setup
func restoreOriginalGenesis(t *testing.T, s SystemUnderTest) {
	src := filepath.Join(workDir, s.nodePath(0), "config", "genesis.json.orig")
	s.SetGenesis(t, src)
}

// copyFile copy source file to dest file path
func copyFile(src, dest string) (*os.File, error) {
	in, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	out, err := os.Create(dest)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return out, err
}
