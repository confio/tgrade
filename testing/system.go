package testing

import (
	"bufio"
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
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var workDir string // this be done better

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
}

func (s SystemUnderTest) StartChain(t *testing.T) {
	s.Log("Start chain")
	s.withEachNodes(t, "start", "--trace", "--log_level=info")

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
			s.Logf("Node has block %d", result.SyncInfo.LatestBlockHeight)
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
		//panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
		s.Logf("failed to stop chain: %s\n", err)
	}
	s.Log(string(out))
}

func (s SystemUnderTest) PrintBuffer() {
	s.outBuff.Do(func(v interface{}) {
		if v != nil {
			s.Logf("err> %s\n", v)
		}
	})
	s.Log("8< chain err -----------------------------------------\n")
	s.errBuff.Do(func(v interface{}) {
		if v != nil {
			s.Logf("err> %s\n", v)
		}
	})
}

func (s SystemUnderTest) BuildNewArtifact() {
	s.Log("install binaries")
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
		t.Fatalf("timeout - no block within %s", s.blockTime*2)
	}
}

// Restart stops and clears all nodes state via 'unsafe-reset-all'
func (s SystemUnderTest) Restart(t *testing.T) {
	t.Log("Restart chain")
	s.StopChain()
	cmd := exec.Command(locateExecutable("pkill"), "tgrade")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Logf("failed to kill process: %s", err)
	}
	s.Log(string(out))

	// reset all nodes
	s.withEachNodes(t, "unsafe-reset-all")
}

func (s SystemUnderTest) withEachNodes(t *testing.T, args ...string) {
	for i := 0; i < s.nodesCount; i++ {
		args = append(args, "--home", fmt.Sprintf("%s/node%d/tgrade", s.outputDir, i))

		s.Logf("execute `tgrade %s` node %d\n", strings.Join(args, ","), i)
		cmd := exec.Command(
			locateExecutable("tgrade"),
			args...,
		)
		cmd.Dir = workDir
		s.watchLogs(cmd)
		require.NoError(t, cmd.Start(), "node %d", i)
	}
}

func (s SystemUnderTest) Log(msg string) {
	if s.verbose {
		fmt.Fprint(s.out, msg)
	}
}

func (s SystemUnderTest) Logf(msg string, args ...interface{}) {
	s.Log(fmt.Sprintf(msg, args...))
}

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
			t.Fatalf("timeout waiting for new events %s", waitTime)
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
