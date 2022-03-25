package web

import (
	"log"
	"os"
	"testing"
	"time"
)

var testdataDir string

func init() {
	wd, err := os.Getwd()
	check(err)
	testdataDir = wd + "/testdata"
	log.Printf("testdata='%s'\n", testdataDir)
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
func TestCont(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 3"
	ps := &ProcessStatusObject{}
	stall := "0" //"0.3"
	_ = StartControlledBashProcess(bash, ps, stall)
	/*
		result, _ := WatchBash(bash, ps, nil)
		select {
		case <-result.chanComplete:
		}
	*/
}
func TestWatchBashSucceed(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env)
	if err != nil {
		t.Errorf("Got %d", err)
	}
	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	if code != 0 {
		t.Errorf("Expected 0, got exit-code %d", code)
	}
	if ps.Armed {
		t.Errorf("ProcessStatus.Armed should be false when the process completes.")
	}
}
func TestWatchBashKill(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
		"STATUS_TIMEOUT=0.001",
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env)
	if err != nil {
		t.Errorf("Got %d", err)
	}
	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	if code != -1 {
		t.Errorf("Expected -1 (from timeout), got exit-code %d", code)
	}
	if ps.Armed {
		t.Errorf("ProcessStatus.Armed should be false when the process completes.")
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
	{
		sr, err := String2StatusReport(`PA_BASECALLER_STATUS {"counter":0,"counterMax":1,"ready":false,"stageName":"StartUp","stageNumber":0,"stageWeights":[10, 80, 10],"state":"progress","timeStamp":"2022-03-21T23:27:40Z","timeoutForNextStatus":300}`)
		check(err)
		if sr.State == "exception" {
			t.Errorf("Got %v", sr)
		} else if sr.TimeoutForNextStatus != 300 {
			t.Errorf("Got TimeoutForNextStatus=%f", sr.TimeoutForNextStatus)
		}
	}
}
func TestTimestamp(t *testing.T) {
	layout := "Mon Jan 02 2006 15:04:05 GMT-0700"
	sample := "Fri Sep 23 2017 15:38:22 GMT+0630"
	expected := "2017-09-23T09:08:22Z"
	tt, err := time.Parse(layout, sample)
	check(err)
	ts := Timestamp(tt)
	if ts != expected {
		t.Errorf("Got %v", ts)
	}
}
func TestFirstWord(t *testing.T) {
	{
		got := FirstWord("hello world")
		expected := "hello"
		if got != expected {
			t.Errorf("Got '%v', expected '%v'", got, expected)
		}
	}
	{
		got := FirstWord(" after space ")
		expected := "after"
		if got != expected {
			t.Errorf("Got '%v', expected '%v'", got, expected)
		}
	}
	{
		got := FirstWord("")
		expected := ""
		if got != expected {
			t.Errorf("Got '%v', expected '%v'", got, expected)
		}
	}
}
