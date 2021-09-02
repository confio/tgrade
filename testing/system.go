package testing

import (
	"bufio"
	"bytes"
	"container/ring"
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/sync"
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
	blockListener     *EventListener
	currentHeight     int64
	chainID           string
	outputDir         string
	blockTime         time.Duration
	rpcAddr           string
	initialNodesCount int
	nodesCount        int
	minGasPrice       string
	cleanupFn         []CleanupFn
	outBuff           *ring.Ring
	errBuff           *ring.Ring
	out               io.Writer
	verbose           bool
}

func NewSystemUnderTest(verbose bool, nodesCount int, blockTime time.Duration) *SystemUnderTest {
	return &SystemUnderTest{
		chainID:           "testing",
		outputDir:         "./testnet",
		blockTime:         blockTime,
		rpcAddr:           "tcp://localhost:26657",
		initialNodesCount: nodesCount,
		outBuff:           ring.New(100),
		errBuff:           ring.New(100),
		out:               os.Stdout,
		verbose:           verbose,
	}
}

func (s *SystemUnderTest) SetupChain() {
	s.Logf("Setup chain: %s\n", s.outputDir)
	if err := os.RemoveAll(filepath.Join(workDir, s.outputDir)); err != nil {
		panic(err.Error())
	}
	args := []string{
		"testnet",
		"--chain-id=" + s.chainID,
		"--output-dir=" + s.outputDir,
		"--v=" + strconv.Itoa(s.initialNodesCount),
		"--keyring-backend=test",
		"--commit-timeout=" + s.blockTime.String(),
		"--minimum-gas-prices=" + s.minGasPrice,
		"--starting-ip-address", "", // empty to use host systems
		"--single-host",
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

func (s *SystemUnderTest) StartChain(t *testing.T, xargs ...string) {
	t.Helper()
	s.Log("Start chain\n")
	s.forEachNodesExecAsync(t, append([]string{"start", "--trace", "--log_level=info"}, xargs...)...)

	s.AwaitNodeUp(t, s.rpcAddr)

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
	stopRingBuffer := make(chan struct{})
	go appendToBuf(io.TeeReader(errReader, logfile), s.errBuff, stopRingBuffer)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("stdout reader error %#+v", err))
	}
	go appendToBuf(io.TeeReader(outReader, logfile), s.outBuff, stopRingBuffer)
	s.cleanupFn = append(s.cleanupFn, func() {
		close(stopRingBuffer)
		logfile.Close()
	})
}

func appendToBuf(r io.Reader, b *ring.Ring, stop <-chan struct{}) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case _, _ = <-stop:
			return
		default:
		}
		b.Value = scanner.Text()
		b = b.Next()
	}
}

// AwaitNodeUp ensures the node is running
func (s *SystemUnderTest) AwaitNodeUp(t *testing.T, rpcAddr string) {
	t.Helper()
	t.Logf("Await node is up: %s", rpcAddr)
	timeout := defaultWaitTime
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()

	started := make(chan struct{})
	go func() { // query for a non empty block on status page
		t.Logf("Checking node status: %s\n", rpcAddr)
		for {
			con, err := client.New(rpcAddr, "/websocket")
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
		t.Fatalf("timeout waiting for node start: %s", timeout)
	}
}

// StopChain stops the system under test and executes all registered cleanup callbacks
func (s *SystemUnderTest) StopChain() {
	s.Log("Stop chain")
	for _, c := range s.cleanupFn {
		c()
	}
	s.cleanupFn = nil
	cmd := exec.Command(locateExecutable("pkill"), "-15", "tgrade")
	cmd.Dir = workDir
	_, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to stop chain: %s\n", err)
	}
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

// AwaitNextBlock is a first class function that any caller can use to ensure a new block was minted.
// Returns the new height
func (s *SystemUnderTest) AwaitNextBlock(t *testing.T, timeout ...time.Duration) int64 {
	t.Helper()
	var maxWaitTime = s.blockTime * 3
	if len(timeout) != 0 { // optional argument to overwrite default timeout
		maxWaitTime = timeout[0]
	}
	done := make(chan int64)
	go func() {
		for start, current := atomic.LoadInt64(&s.currentHeight), atomic.LoadInt64(&s.currentHeight); current == start; current = atomic.LoadInt64(&s.currentHeight) {
			time.Sleep(s.blockTime)
		}
		done <- atomic.LoadInt64(&s.currentHeight)
		close(done)
	}()
	select {
	case v := <-done:
		return v
	case <-time.NewTimer(maxWaitTime).C:
		t.Fatalf("Timeout - no block within %s", maxWaitTime)
		return -1
	}
}

// ResetChain stops and clears all nodes state via 'unsafe-reset-all'
func (s *SystemUnderTest) ResetChain(t *testing.T) {
	t.Helper()
	t.Log("ResetChain chain")
	s.StopChain()
	restoreOriginalGenesis(t, *s)
	restoreOriginalKeyring(t, *s)
	s.resetBuffers()

	cmd := exec.Command(locateExecutable("pkill"), "tgrade")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to kill process: %s", err)
	}
	s.Log(string(out))

	// remove all additional nodes
	for i := s.initialNodesCount; i < s.nodesCount; i++ {
		os.Remove(s.nodePath(i))
	}
	s.nodesCount = s.initialNodesCount

	// reset all validataor nodes
	s.ForEachNodeExecAndWait(t, []string{"unsafe-reset-all"})
	s.withEachNodeHome(func(i int, home string) {
		os.Remove(filepath.Join(workDir, home, "wasm"))
	})
}

// ModifyGenesisCLI executes the CLI commands to modify the genesis
func (s *SystemUnderTest) ModifyGenesisCLI(t *testing.T, cmds ...[]string) {
	s.ForEachNodeExecAndWait(t, cmds...)
}

type GenesisMutator func([]byte) []byte

// ModifyGenesisJson executes the callbacks to update the json representation
// The mutator callbacks after each other receive the genesis as raw bytes and return the updated genesis for the next.
// example:
// 	return func(genesis []byte) []byte {
//		val, _ := json.Marshal(sdk.NewDecCoins(fees...))
//		state, _ := sjson.SetRawBytes(genesis, "app_state.globalfee.params.minimum_gas_prices", val)
//		return state
//	}
func (s *SystemUnderTest) ModifyGenesisJson(t *testing.T, mutators ...GenesisMutator) {
	current, err := ioutil.ReadFile(filepath.Join(workDir, s.nodePath(0), "config", "genesis.json"))
	require.NoError(t, err)
	for _, m := range mutators {
		current = m(current)
	}
	out := storeTempFile(t, current)
	defer os.Remove(out.Name())
	s.SetGenesis(t, out.Name())
}

// ReadGenesisJson returns current genesis.json content as raw string
func (s *SystemUnderTest) ReadGenesisJson(t *testing.T) string {
	content, err := ioutil.ReadFile(filepath.Join(workDir, s.nodePath(0), "config", "genesis.json"))
	require.NoError(t, err)
	return string(content)
}

// SetGenesis copy genesis file to all nodes
func (s *SystemUnderTest) SetGenesis(t *testing.T, srcPath string) {
	in, err := os.Open(srcPath)
	require.NoError(t, err)
	defer in.Close()
	var buf bytes.Buffer

	_, err = io.Copy(&buf, in)
	require.NoError(t, err)

	s.withEachNodeHome(func(i int, home string) {
		saveGenesis(t, home, buf.Bytes())
	})
}

func saveGenesis(t *testing.T, home string, content []byte) {
	out, err := os.Create(filepath.Join(workDir, home, "config", "genesis.json"))
	require.NoError(t, err)
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, out.Close())
}

// ForEachNodeExecAndWait runs the given tgrade commands for all cluster nodes synchronously
// The commands output is returned for each node.
func (s *SystemUnderTest) ForEachNodeExecAndWait(t *testing.T, cmds ...[]string) [][]string {
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
func (s *SystemUnderTest) forEachNodesExecAsync(t *testing.T, xargs ...string) []func() error {
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
	ip, err := server.ExternalIP()
	require.NoError(t, err)

	for i, out := range outs {
		result[i] = Node{
			ID:      strings.TrimSpace(out[0]),
			IP:      ip,
			RPCPort: 25567 + i, // as defined in testnet command
			P2PPort: 16656 + i, // as defined in testnet command
		}
	}
	return result
}

func (s *SystemUnderTest) resetBuffers() {
	s.outBuff = ring.New(100)
	s.errBuff = ring.New(100)
}

// AddFullnode starts a new fullnode that connects to the existing chain but is not a validator.
func (s *SystemUnderTest) AddFullnode(t *testing.T) Node {
	s.nodesCount++
	nodeNumber := s.nodesCount - 1
	nodePath := s.nodePath(nodeNumber)
	// prepare new node
	moniker := fmt.Sprintf("node%d", nodeNumber)
	args := []string{"init", moniker, "--home", nodePath}
	s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
	cmd := exec.Command(
		locateExecutable("tgrade"),
		args...,
	)
	cmd.Dir = workDir
	s.watchLogs(nodeNumber, cmd)
	require.NoError(t, cmd.Run(), "node %d", nodeNumber)
	saveGenesis(t, nodePath, []byte(s.ReadGenesisJson(t)))

	// quick hack: copy config and overwrite by start params
	configFile := filepath.Join(workDir, nodePath, "config", "config.toml")
	_ = os.Remove(configFile)
	_, err := copyFile(filepath.Join(workDir, s.nodePath(0), "config", "config.toml"), configFile)
	require.NoError(t, err)

	// start node
	allNodes := s.AllNodes(t)
	node := allNodes[len(allNodes)-1]
	var peers []string
	for _, n := range allNodes[0 : len(allNodes)-1] {
		peers = append(peers, n.PeerAddr())
	}
	args = []string{
		"start",
		"--p2p.persistent_peers=" + strings.Join(peers, ","),
		fmt.Sprintf("--p2p.laddr=tcp://localhost:%d", node.P2PPort),
		fmt.Sprintf("--rpc.laddr=tcp://localhost:%d", node.RPCPort),
		fmt.Sprintf("--grpc.address=localhost:%d", 9090+nodeNumber),
		"--moniker=" + moniker,
		"--trace", "--log_level=info",
		"--home", nodePath,
	}
	s.Logf("Execute `tgrade %s`\n", strings.Join(args, " "))
	cmd = exec.Command(
		locateExecutable("tgrade"),
		args...,
	)
	cmd.Dir = workDir
	s.watchLogs(nodeNumber, cmd)
	require.NoError(t, cmd.Start(), "node %d", nodeNumber)
	return node
}

// NewEventListener constructor for Eventlistener with system rpc address
func (s *SystemUnderTest) NewEventListener(t *testing.T) *EventListener {
	return NewEventListener(t, s.rpcAddr)
}

type Node struct {
	ID      string
	IP      string
	RPCPort int
	P2PPort int
}

func (n Node) PeerAddr() string {
	return fmt.Sprintf("%s@%s:%d", n.ID, n.IP, n.P2PPort)
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
func NewEventListener(t *testing.T, rpcAddr string) *EventListener {
	httpClient, err := client.New(rpcAddr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	return &EventListener{client: httpClient, t: t}
}

var defaultWaitTime = 30 * time.Second

type (
	CleanupFn     func()
	EventConsumer func(e ctypes.ResultEvent) (more bool)
)

// Subscribe to receive events for a topic. Does not block.
// For query syntax See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
func (l *EventListener) Subscribe(query string, cb EventConsumer) func() {
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

// AwaitQuery blocks and waits for a single result or timeout. This can be used with `broadcast-mode=async`.
// For query syntax See https://docs.cosmos.network/master/core/events.html#subscribing-to-events
func (l *EventListener) AwaitQuery(query string, optMaxWaitTime ...time.Duration) *ctypes.ResultEvent {
	c, result := CaptureSingleEventConsumer()
	maxWaitTime := defaultWaitTime
	if len(optMaxWaitTime) != 0 {
		maxWaitTime = optMaxWaitTime[0]
	}
	cleanupFn := l.Subscribe(query, TimeoutConsumer(l.t, maxWaitTime, c))
	l.t.Cleanup(cleanupFn)
	return result
}

// TimeoutConsumer is an event consumer decorator with a max wait time. Panics when wait time exceeded without
// a result returned
func TimeoutConsumer(t *testing.T, maxWaitTime time.Duration, next EventConsumer) EventConsumer {
	ctx, done := context.WithCancel(context.Background())
	timeout := time.NewTimer(maxWaitTime)
	go func() {
		select {
		case <-ctx.Done():
		case <-timeout.C:
			t.Fatalf("Timeout waiting for new events %s", maxWaitTime)
		}
	}()
	return func(e ctypes.ResultEvent) (more bool) {
		timeout.Reset(maxWaitTime)
		result := next(e)
		if !result {
			done()
		}
		return result
	}
}

// CaptureSingleEventConsumer consumes one event. No timeout
func CaptureSingleEventConsumer() (EventConsumer, *ctypes.ResultEvent) {
	var result ctypes.ResultEvent
	return func(e ctypes.ResultEvent) (more bool) {
		return false
	}, &result
}

// CaptureAllEventsConsumer is an `EventConsumer` that captures all events until `done()` is called to stop or timeout happens.
// The consumer works async in the background and returns all the captured events when `done()` is called.
// This can be used to verify that certain events have happened.
// Example usage:
// 	c, done := CaptureAllEventsConsumer(t)
//	query := `tm.event='Tx'`
//	cleanupFn := l.Subscribe(query, c)
//	t.Cleanup(cleanupFn)
//
//  // do something in your test that create events
//
//	assert.Len(t, done(), 1) // then verify your assumption
func CaptureAllEventsConsumer(t *testing.T, optMaxWaitTime ...time.Duration) (c EventConsumer, done func() []ctypes.ResultEvent) {
	maxWaitTime := defaultWaitTime
	if len(optMaxWaitTime) != 0 {
		maxWaitTime = optMaxWaitTime[0]
	}
	var (
		mu             sync.Mutex
		capturedEvents []ctypes.ResultEvent
		exit           bool
	)
	collectEventsConsumer := func(e ctypes.ResultEvent) (more bool) {
		mu.Lock()
		defer mu.Unlock()
		if exit {
			return false
		}
		capturedEvents = append(capturedEvents, e)
		return true
	}

	return TimeoutConsumer(t, maxWaitTime, collectEventsConsumer), func() []ctypes.ResultEvent {
		mu.Lock()
		defer mu.Unlock()
		exit = true
		return capturedEvents
	}
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

func storeTempFile(t *testing.T, content []byte) *os.File {
	out, err := ioutil.TempFile(t.TempDir(), "genesis")
	require.NoError(t, err)
	_, err = io.Copy(out, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, out.Close())
	return out
}
