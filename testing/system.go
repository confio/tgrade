package testing

import (
	"bufio"
	"context"
	"fmt"
	client "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync/atomic"
	"time"
)

const workDir = "../" // this be done better

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
	out           io.Writer
	cleanupFn     []CleanupFn
}

func NewSystemUnderTest() *SystemUnderTest {
	return &SystemUnderTest{
		chainID:    "testing",
		outputDir:  "./testnet",
		blockTime:  1500 * time.Millisecond,
		rpcAddr:    "http://localhost:26657",
		nodesCount: 1,
		out:        os.Stdout,
	}
}

func (s SystemUnderTest) SetupChain() {
	s.Log("Setup chain")
	tgradePath := locateExecutable("tgrade")
	cmd := exec.Command(
		tgradePath,
		"testnet",
		"--chain-id="+s.chainID,
		"--output-dir="+s.outputDir,
		"--v="+strconv.Itoa(s.nodesCount),
		"--keyring-backend=test",
		"--commit-timeout="+s.blockTime.String(),
		"--minimum-gas-prices="+s.minGasPrice,
	)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	s.Log(string(out))
}

func (s SystemUnderTest) StartChain() {
	s.Log("Start chain")
	dockerComposePath := locateExecutable("docker-compose")
	cmd := exec.Command(dockerComposePath, "up")
	cmd.Dir = workDir
	watchLogs(cmd)
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}

	s.awaitChainUp()

	s.Log("Start new block listener")
	s.blockListener = NewEventListener(s.rpcAddr)
	s.cleanupFn = append(s.cleanupFn,
		s.blockListener.Subscribe("tm.event='NewBlock'", func(e ctypes.ResultEvent) (more bool) {
			newBlock, ok := e.Data.(types.EventDataNewBlock)
			if !ok {
				panic(fmt.Sprintf("unexpected type %T", e.Data))
			}
			atomic.StoreInt64(&s.currentHeight, newBlock.Block.Height)
			return true
		}))
}

func watchLogs(cmd *exec.Cmd) {
	errReader, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}
	go printToStdout(errReader)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}
	go printToStdout(outReader)

}

func printToStdout(r io.ReadCloser) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		fmt.Println(s)
	}
}

// awaitChainUp ensures the chain is running
func (s SystemUnderTest) awaitChainUp() {
	s.Log("Await chain starts")
	timeout := defaultWaitTime
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()

	started := make(chan struct{})
	go func() { // query for a non empty block on status page
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
			con.Stop()
			started <- struct{}{}
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			panic(fmt.Sprintf("unexpected error: %#+v", err))
		}
	case <-time.NewTimer(timeout).C:
		panic(fmt.Sprintf("timeout waiting for chain start: %s", timeout))
	}
}

// StopChain stops the system under test and executes all registered cleanup callbacks
func (s SystemUnderTest) StopChain() {
	s.Log("stop chain")
	for _, c := range s.cleanupFn {
		c()
	}
	s.cleanupFn = nil
	dockerComposePath := locateExecutable("docker-compose")
	cmd := exec.Command(dockerComposePath, "stop")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	s.Log(string(out))
}

func (s SystemUnderTest) CompileBinaries() {
	s.Log("compile binaries")
	makePath := locateExecutable("make")
	cmd := exec.Command(makePath, "clean", "install", "build-docker")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#v : output: %s", err, string(out)))
	}
}

// AwaitNextBlock is a first class function that any caller can use to ensure a new block was minted
func (s SystemUnderTest) AwaitNextBlock() {
	done := make(chan struct{})
	go func() {
		for start := atomic.LoadInt64(&s.currentHeight); atomic.LoadInt64(&s.currentHeight) > start; {
			time.Sleep(s.blockTime) // block time?
		}
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.NewTimer(s.blockTime * 2).C:
		panic(fmt.Sprintf("timeout - no block within %s", s.blockTime*2))
	}
}

// Restart stops and clears all nodes state via 'unsafe-reset-all'
func (s SystemUnderTest) Restart() {
	s.Log("Restart chain")
	s.StopChain()
	dockerComposePath := locateExecutable("docker-compose")
	cmd := exec.Command(dockerComposePath, "kill")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#v : output: %s", err, string(out)))
	}

	s.Log(string(out))
	// reset all nodes
	tgradePath := locateExecutable("tgrade")
	for i := 0; i < s.nodesCount; i++ {
		s.Logf("unsafe reset node %d", i)
		cmd := exec.Command(
			tgradePath,
			"unsafe-reset-all", "--home", fmt.Sprintf("%s/node%d/tgrade", s.outputDir, i))
		cmd.Dir = workDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("unexpected error %#v : output: %s", err, string(out)))
		}
		s.Log(string(out))
	}
}

func (s SystemUnderTest) Log(msg string) {
	fmt.Fprint(s.out, msg)
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
	client *client.HTTP
}

// NewEventListener event listener
func NewEventListener(rpcAddr string) EventListener {
	httpClient, err := client.New(rpcAddr, "/websocket")
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}
	if err := httpClient.Start(); err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}

	return EventListener{client: httpClient}
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
	if err != nil {
		panic(fmt.Sprintf("unexpected error %#+v", err))
	}

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
	l.Subscribe(query, TimeoutConsumer(defaultWaitTime, c))
	return result
}

// TimeoutConsumer is an event consumer decorator with a max wait time. Panics when wait time exceeded without
// a result returned
func TimeoutConsumer(waitTime time.Duration, next EventConsumer) EventConsumer {
	ctx, done := context.WithCancel(context.Background())
	timeout := time.NewTimer(waitTime)
	go func() {
		select {
		case <-ctx.Done():
		case <-timeout.C:
			panic(fmt.Sprintf("timeout waiting for new events %s", waitTime))
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
