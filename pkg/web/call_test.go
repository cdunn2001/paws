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
	bash := testdataDir + "/dummy-basic.sh --status-fd 2"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 3"
	ps := &ProcessStatusObject{}
	setup := ProcessSetupObject{
		Hostname: "localhost",
		Stall:    "0",        //"0.3"
		ScriptFn: "where.sh", // TODO: Clean up this file!
	}
	WriteStringToFile(bash, setup.ScriptFn)
	_ = StartControlledShellProcess(setup, ps)
	/*
		result, _ := WatchBash(bash, ps, nil, "dummy-pa-cal")
		select {
		case <-result.chanComplete:
		}
	*/
}
func TestWatchBashStderrSucceed(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 2"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 2"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBashStderr(bash, ps, env, "dummy-basic")
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
func TestWatchBashSucceed(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env, "dummy-basic")
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
func TestWatchBashStderrKill(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 2"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
		"STATUS_TIMEOUT=0.00001",
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBashStderr(bash, ps, env, "dummy-basic")
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
func TestWatchBashKill(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 3"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=0.05", // Note: ".05" would not be valid.
		"STATUS_TIMEOUT=0.00001",
	}
	ps := &ProcessStatusObject{}
	cp, err := WatchBash(bash, ps, env, "dummy-basic")
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
		sr, err := String2StatusReport(`BLAH BLAH BLAH`)
		if err == nil {
			t.Errorf("Expected err; Got %v", sr)
		}
	}
	{
		sr, err := String2StatusReport(`+ X_STATUS {"counter": 123}`)
		if err == nil {
			t.Errorf("Expected err; Got %v", sr)
		}
	}
	{
		sr, err := String2StatusReport(`X_STATUS {"counter": 123}`)
		check(err)
		if sr.State == "exception" {
			t.Errorf("Got %v", sr)
		} else if sr.Counter != 123 {
			t.Errorf("Got %d", sr.Counter)
		}
	}
	{
		sr, err := String2StatusReport(`X_STATUS {"state": "exception", "message": "HELLO"}`)
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
func TestProgressMetricsObjectFromStatusReport(t *testing.T) {
	{
		sr := StatusReport{
			Counter:      0,
			CounterMax:   0,
			StageWeights: []int32{1, 99},
		}
		expected := ProgressMetricsObject{}
		got := ProgressMetricsObjectFromStatusReport(sr)
		if got.StageProgress != expected.StageProgress {
			t.Errorf("Got StageProgress '%v', expected '%v'", got.StageProgress, expected.StageProgress)
		}
		if got.NetProgress != expected.NetProgress {
			t.Errorf("Got NetProgress '%v', expected '%v'", got.NetProgress, expected.NetProgress)
		}
	}
	{
		sr := StatusReport{
			Counter:      1,
			CounterMax:   2,
			StageNumber:  2,
			StageWeights: []int32{99, 0, 1},
		}
		expected := ProgressMetricsObject{
			StageProgress: 0.5,
			NetProgress:   0.995,
		}
		got := ProgressMetricsObjectFromStatusReport(sr)
		if got.StageProgress != expected.StageProgress {
			t.Errorf("Got StageProgress '%v', expected '%v'", got.StageProgress, expected.StageProgress)
		}
		if got.NetProgress != expected.NetProgress {
			t.Errorf("Got NetProgress '%v', expected '%v'", got.NetProgress, expected.NetProgress)
		}
	}
}
func TestSplitExt(t *testing.T) {
	{
		gotBase, gotExt := SplitExt("foo.bar.baz")
		expectedBase := "foo.bar"
		expectedExt := ".baz"
		if gotBase != expectedBase {
			t.Errorf("Got Base '%v', expected '%v'", gotBase, expectedBase)
		}
		if gotExt != expectedExt {
			t.Errorf("Got Ext '%v', expected '%v'", gotExt, expectedExt)
		}
	}
	{
		gotBase, gotExt := SplitExt("fubar")
		expectedBase := "fubar"
		expectedExt := ""
		if gotBase != expectedBase {
			t.Errorf("Got Base '%v', expected '%v'", gotBase, expectedBase)
		}
		if gotExt != expectedExt {
			t.Errorf("Got Ext '%v', expected '%v'", gotExt, expectedExt)
		}
	}
	{
		gotBase, gotExt := SplitExt(".snafu")
		expectedBase := ""
		expectedExt := ".snafu"
		if gotBase != expectedBase {
			t.Errorf("Got Base '%v', expected '%v'", gotBase, expectedBase)
		}
		if gotExt != expectedExt {
			t.Errorf("Got Ext '%v', expected '%v'", gotExt, expectedExt)
		}
	}
}
func TestChooseLoggerFilenameTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		check(err)
		got := chooseLoggerFilenameTestable("foo.log", mytime, 123)
		expected := "foo.06-01-02.123.log"
		if got != expected {
			t.Errorf("Got '%v', expected '%v'", got, expected)
		}
	}
}
func TestChooseLoggerFilenameLegacyTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		check(err)
		got := chooseLoggerFilenameLegacyTestable("foo.log", mytime)
		expected := "foo.06-01-02.log"
		if got != expected {
			t.Errorf("Got '%v', expected '%v'", got, expected)
		}
	}
}
