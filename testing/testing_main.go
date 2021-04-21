// +build system_test

package testing

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var sut *SystemUnderTest
var verbose bool

func RunTestMain(m *testing.M) {
	rebuild := flag.Bool("rebuild", false, "rebuild artifacts")
	waitTime := flag.Duration("wait-time", defaultWaitTime, "time to wait for chain events")
	nodesCount := flag.Int("nodes-count", 4, "number of nodes in the cluster")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workDir = filepath.Join(dir, "../")
	if verbose {
		println("Work dir: ", workDir)
	}
	defaultWaitTime = *waitTime
	sut = NewSystemUnderTest(verbose, *nodesCount)
	if *rebuild {
		sut.BuildNewBinary()
	}
	// setup chain and keyring
	sut.SetupChain()

	// run tests
	exitCode := m.Run()

	// postprocess
	sut.StopChain()
	if verbose || exitCode != 0 {
		sut.PrintBuffer()
		printResultFlag(exitCode == 0)
	}

	os.Exit(exitCode)
}

const (
	successFlag = `
 ___ _   _  ___ ___ ___  ___ ___ 
/ __| | | |/ __/ __/ _ \/ __/ __|
\__ \ |_| | (_| (_|  __/\__ \__ \
|___/\__,_|\___\___\___||___/___/`
	failureFlag = `
  __      _ _          _ 
 / _|    (_) |        | |
| |_ __ _ _| | ___  __| |
|  _/ _| | | |/ _ \/ _| |
| || (_| | | |  __/ (_| |
|_| \__,_|_|_|\___|\__,_|`
)

func printResultFlag(ok bool) {
	if ok {
		fmt.Println(successFlag)
	} else {
		fmt.Println(failureFlag)
	}
}
