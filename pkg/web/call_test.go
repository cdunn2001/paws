package web

import (
	"fmt"
	"os"
	"testing"
	//"time"
)

var testdataDir string

func init() {
	wd, err := os.Getwd()
	check(err)
	testdataDir = wd + "/testdata"
	fmt.Printf("testdata='%s'\n", testdataDir)
	requireFile(testdataDir)
}
func requireFile(fn string) {
	_, err := os.Stat(fn)
	check(err)
}

func TestPlus1(t *testing.T) {
	// Just an example.
	got := plus1(2)
	if got != 3 {
		t.Errorf("Got %d", got)
	}
}
func TestRunBash(t *testing.T) {
	// We might not use this, but it works.
	bash := testdataDir + "/dummy-basic.sh"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=.1",
	}
	got := RunBash(bash, env)
	if got != nil {
		t.Errorf("Got %d", got)
	}
}
func TestWatchBash(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env)
	if err != nil {
		t.Errorf("Got %d", err)
	}
	fmt.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	fmt.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	if code != 0 {
		t.Errorf("Expected 0, got exit-code %d", code)
	}
}
func TestWatchBashKill(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env)
	if err != nil {
		t.Errorf("Got %d", err)
	}
	fmt.Printf("Sending to chanKill\n")
	cp.chanKill <- true
	fmt.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	fmt.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	if code != -1 {
		t.Errorf("Expected -1, got exit-code %d", code)
	}
}
func TestString2StatusReport(t *testing.T) {
	{
		sr, err := String2StatusReport(`_STATUS {"counter": 123}`)
		check(err)
		if sr.State == "exception" {
			t.Errorf("Got %v", sr)
		} else if sr.Counter != 123 {
			t.Errorf("Got %d", sr.Counter)
		}
	}
	{
		sr, err := String2StatusReport(`_STATUS {"state": "exception", "message": "HELLO"}`)
		check(err)
		if sr.State != "exception" {
			t.Errorf("Got %v", sr)
		} else if sr.Message != "HELLO" {
			t.Errorf("Got %s", sr.Message)
		}
	}
}
