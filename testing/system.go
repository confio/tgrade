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
	tmtypes "github.com/tendermint/tendermint/types"
	"io"
	"io/ioutil"
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

func NewSystemUnderTest(verbose bool, nodesCount int) *SystemUnderTest {
	return &SystemUnderTest{
		chainID:    "testing",
		outputDir:  "./testnet",
		blockTime:  1500 * time.Millisecond,
		rpcAddr:    "tcp://localhost:26657",
		nodesCount: nodesCount,
		outBuff:    ring.New(100),
		errBuff:    ring.New(100),
		out:        os.Stdout,
		verbose:    verbose,
	}
}

func (s SystemUnderTest) SetupChain() {
	s.Logf("Setup chain: %s\n", s.outputDir)
	if err := os.RemoveAll(filepath.Join(workDir, s.outputDir)); err != nil {
		panic(err.Error())
	}
	args := []string{
		"testnet",
		"--chain-id=" + s.chainID,
		"--output-dir=" + s.outputDir,
		"--v=" + strconv.Itoa(s.nodesCount),
		"--keyring-backend=test",
		"--commit-timeout=" + s.blockTime.String(),
		"--minimum-gas-prices=" + s.minGasPrice,
		"--starting-ip-address", "", // empty to use host systems
		"--single-machine",
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
	// backup keyring
	src = filepath.Join(workDir, s.nodePath(0), "keyring-test")
	dest = filepath.Join(workDir, s.outputDir, "keyring-test")
	if err := copyFilesInDir(src, dest); err != nil {
		panic(fmt.Sprintf("copy files from dir :%#+v", err))
	}
}

func (s *SystemUnderTest) StartChain(t *testing.T) {
	s.Log("Start chain\n")
	s.forEachNodesExecAsync(t, "start", "--trace", "--log_level=info")

	s.awaitChainUp(t)

	t.Log("Start new block listener")
	s.blockListener = NewEventListener(t, s.rpcAddr)
	s.cleanupFn = append(s.cleanupFn,
		s.blockListener.Subscribe("tm.event='NewBlock'", func(e ctypes.ResultEvent) (more bool) {
			newBlock, ok := e.Data.(tmtypes.EventDataNewBlock)
			require.True(t, ok, "unexpected type %T", e.Data)
			atomic.StoreInt64(&s.currentHeight, newBlock.Block.Height)
			return true
		}),
	)
	s.AwaitNextBlock(t)
}

// watchLogs stores stdout/stderr in a file and in a ring buffer to output the last n lines on test error
func (s *SystemUnderTest) watchLogs(node int, cmd *exec.Cmd) {
	logfile, err := os.Create(filepath.Join(workDir, s.outputDir, fmt.Sprintf("node%d.out", node)))
	if err != nil {
		panic(fmt.Sprintf("open logfile error %#+v", err))
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("stderr reader error %#+v", err))
	}
	go appendToBuf(io.TeeReader(errReader, logfile), s.errBuff)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("stdout reader error %#+v", err))
	}
	go appendToBuf(io.TeeReader(outReader, logfile), s.outBuff)
}

func appendToBuf(r io.Reader, b *ring.Ring) {
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
func (s *SystemUnderTest) AwaitNextBlock(t *testing.T, timeout ...time.Duration) {
	t.Helper()
	var maxWaitTime = s.blockTime * 2
	if len(timeout) != 0 { // optional argument to overwrite default timeout
		maxWaitTime = timeout[0]
	}
	done := make(chan struct{})
	go func() {
		for start := atomic.LoadInt64(&s.currentHeight); atomic.LoadInt64(&s.currentHeight) == start; {
			time.Sleep(s.blockTime)
		}
		done <- struct{}{}
		defer close(done)
	}()
	select {
	case <-done:
	case <-time.NewTimer(maxWaitTime).C:
		t.Fatalf("Timeout - no block within %s", maxWaitTime)
	}
}

// ResetChain stops and clears all nodes state via 'unsafe-reset-all'
func (s SystemUnderTest) ResetChain(t *testing.T) {
	t.Helper()
	t.Log("ResetChain chain")
	s.StopChain()
	restoreOriginalGenesis(t, s)
	restoreOriginalKeyring(t, s)

	cmd := exec.Command(locateExecutable("pkill"), "tgrade")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to kill process: %s", err)
	}
	s.Log(string(out))

	// reset all nodes
	s.ForEachNodeExecAndWait(t, []string{"unsafe-reset-all"})
	s.withEachNodeHome(func(i int, home string) {
		os.Remove(filepath.Join(workDir, home, "wasm"))
	})
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
// The commands output is returned for each node.
func (s SystemUnderTest) ForEachNodeExecAndWait(t *testing.T, cmds ...[]string) [][]string {
	result := make([][]string, s.nodesCount)
	s.withEachNodeHome(func(i int, home string) {
		result[i] = make([]string, len(cmds))
		for j, xargs := range cmds {
			args := append(xargs, "--home", home)
			s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
			cmd := exec.Command(
				locateExecutable("tgrade"),
				args...,
			)
			cmd.Dir = workDir
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "node %d: %s", i, string(out))
			s.Logf("Result: %s\n", string(out))
			result[i][j] = string(out)
		}
	})
	return result
}

// forEachNodesExecAsync runs the given tgrade command for all cluster nodes and returns without waiting
func (s SystemUnderTest) forEachNodesExecAsync(t *testing.T, xargs ...string) []func() error {
	r := make([]func() error, s.nodesCount)
	s.withEachNodeHome(func(i int, home string) {
		args := append(xargs, "--home", home)
		s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
		cmd := exec.Command(
			locateExecutable("tgrade"),
			args...,
		)
		cmd.Dir = workDir
		s.watchLogs(i, cmd)
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

// nodePath returns the path of the node within the work dir. not absolute
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

func (s SystemUnderTest) RPCClient(t *testing.T) RPCClient {
	return NewRPCCLient(t, s.rpcAddr)
}

func (s SystemUnderTest) AllPeers(t *testing.T) []string {
	result := make([]string, s.nodesCount)
	for i, n := range s.AllNodes(t) {
		result[i] = n.PeerAddr()
	}
	return result
}

func (s SystemUnderTest) AllNodes(t *testing.T) []Node {
	result := make([]Node, s.nodesCount)
	outs := s.ForEachNodeExecAndWait(t, []string{"tendermint", "show-node-id"})
	for i, out := range outs {
		result[i] = Node{
			ID:      out[0],
			IP:      "127.0.0.1",
			RPCPort: 25567 + i,
			P2PPort: 15566 + i,
		}
	}
	return result
}

type Node struct {
	ID      string
	IP      string
	RPCPort int
	P2PPort int
}

func (n Node) PeerAddr() string {
	return fmt.Sprintf("%s@%s:%d", n.ID, n.IP, n.RPCPort)
}

// locateExecutable looks up the binary on the OS path.
func locateExecutable(file string) string {
	path, err := exec.LookPath(file)
	if err != nil {
		panic(fmt.Sprintf("unexpected error %s", err.Error()))
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

// restoreOriginalKeyring replaces test keyring with original
func restoreOriginalKeyring(t *testing.T, s SystemUnderTest) {
	dest := filepath.Join(workDir, s.outputDir, "keyring-test")
	require.NoError(t, os.RemoveAll(dest))

	src := filepath.Join(workDir, s.nodePath(0), "keyring-test")
	require.NoError(t, copyFilesInDir(src, dest))
}

// copyFile copy source file to dest file path
func copyFile(src, dest string) (*os.File, error) {
	in, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return out, err
}

// copyFilesInDir copy files in src dir to dest path
func copyFilesInDir(src, dest string) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("mkdirs: %s", err)
	}
	fs, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir: %s", err)
	}
	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		if _, err := copyFile(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name())); err != nil {
			return fmt.Errorf("copy file: %q: %s", f.Name(), err)
		}
	}
	return nil
}
