package web

import (
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 3, got)
}
func TestRunBash(t *testing.T) {
	// We might not use this, but it works.
	bash := testdataDir + "/dummy-basic.sh"
	env := []string{
		"STATUS_COUNT=3",
		"STATUS_DELAY_SECONDS=.1",
	}
	got := RunBash(bash, env)
	assert.Nil(t, got)
}
func TestStartControlledShellProcessSsh(t *testing.T) {
	bash := testdataDir + "/dummy-basic.sh --status-fd 2"
	//bash := testdataDir + "/dummy-pa-cal.sh --statusfd 3"
	ps := &ProcessStatusObject{}
	setup := ProcessSetupObject{
		Hostname: "localhost",
		Stall:    "0",
		ScriptFn: "where.sh",
	}
	if testing.Short() {
		setup.Hostname = ""
		//t.Skip("skipping this test")
	}
	defer func() {
		_ = os.Remove("where.sh")
	}()
	WriteStringToFile(bash, setup.ScriptFn)
	cp := StartControlledShellProcess(setup, ps)
	//result, _ := WatchBash(bash, ps, nil, "dummy-pa-cal")
	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	assert.Equal(t, int32(0), code, "(from timeout)")
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
	assert.Nil(t, err)

	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	assert.Equal(t, int32(0), code, "(from timeout)")
	assert.False(t, ps.Armed, "ProcessStatus.Armed should be false when the process completes.")
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
	assert.Nil(t, err)

	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	assert.Equal(t, int32(0), code, "(from timeout)")
	assert.False(t, ps.Armed, "ProcessStatus.Armed should be false when the process completes.")
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
	assert.Nil(t, err)

	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	assert.Equal(t, int32(-1), code, "(from timeout)")
	assert.False(t, ps.Armed, "ProcessStatus.Armed should be false when the process completes.")
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
	assert.Nil(t, err)

	log.Printf("Waiting for chanComplete '%s'?", cp.cmd)
	select {
	case <-cp.chanComplete:
	}
	log.Printf("Done '%s'!\n", cp.cmd)
	code := ps.ExitCode
	assert.Equal(t, int32(-1), code, "(from timeout)")
	assert.False(t, ps.Armed, "ProcessStatus.Armed should be false when the process completes.")
}
func TestString2StatusReport(t *testing.T) {
	{
		s := `BLAH BLAH BLAH`
		sr, err := String2StatusReport(s)
		assert.NotNil(t, err, "Expected err for %q from %v", s, sr)
	}
	{
		s := `+ X_STATUS {"counter": 123}`
		sr, err := String2StatusReport(s)
		assert.NotNil(t, err, "Expected err for %q from %v", s, sr)
	}
	{
		sr, err := String2StatusReport(`X_STATUS {"counter": 123}`)
		assert.Nil(t, err)
		assert.NotEqual(t, "exception", sr.State)
		assert.Equal(t, uint64(123), sr.Counter, "Counter")
	}
	{
		sr, err := String2StatusReport(`X_STATUS {"state": "exception", "message": "HELLO"}`)
		assert.Nil(t, err)
		assert.Equal(t, "exception", sr.State)
		assert.Equal(t, "HELLO", sr.Message, "Message")
	}
	{
		sr, err := String2StatusReport(`PA_BASECALLER_STATUS {"counter":0,"counterMax":1,"ready":false,"stageName":"StartUp","stageNumber":0,"stageWeights":[10, 80, 10],"state":"progress","timeStamp":"2022-03-21T23:27:40Z","timeoutForNextStatus":300}`)
		assert.Nil(t, err)
		assert.NotEqual(t, "exception", sr.State)
		assert.Equal(t, 300.0, sr.TimeoutForNextStatus, "TimeoutForNextStatus")
	}
}
func TestTimestamp(t *testing.T) {
	layout := "Mon Jan 02 2006 15:04:05 GMT-0700"
	sample := "Fri Sep 23 2017 15:38:22 GMT+0630"
	expected := "2017-09-23T09:08:22Z"
	tt, err := time.Parse(layout, sample)
	assert.Nil(t, err)
	ts := Timestamp(tt)
	assert.Equal(t, expected, ts)
}
func TestFirstWord(t *testing.T) {
	{
		got := FirstWord("hello world")
		expected := "hello"
		assert.Equal(t, expected, got)
	}
	{
		got := FirstWord(" after space ")
		expected := "after"
		assert.Equal(t, expected, got)
	}
	{
		got := FirstWord("")
		expected := ""
		assert.Equal(t, expected, got)
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
		assert.Equal(t, expected.StageProgress, got.StageProgress, "StageProgress from %v", sr)
		assert.Equal(t, expected.NetProgress, got.NetProgress, "NetProgress from %v", sr)
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
		assert.Equal(t, expected.StageProgress, got.StageProgress, "StageProgress from %v", sr)
		assert.Equal(t, expected.NetProgress, got.NetProgress, "NetProgress from %v", sr)
	}
}
func TestSplitExt(t *testing.T) {
	{
		gotBase, gotExt := SplitExt("foo.bar.baz")
		expectedBase := "foo.bar"
		expectedExt := ".baz"
		assert.Equal(t, expectedBase, gotBase)
		assert.Equal(t, expectedExt, gotExt)
	}
	{
		gotBase, gotExt := SplitExt("fubar")
		expectedBase := "fubar"
		expectedExt := ""
		assert.Equal(t, expectedBase, gotBase)
		assert.Equal(t, expectedExt, gotExt)
	}
	{
		gotBase, gotExt := SplitExt(".snafu")
		expectedBase := ""
		expectedExt := ".snafu"
		assert.Equal(t, expectedBase, gotBase)
		assert.Equal(t, expectedExt, gotExt)
	}
}
func TestChooseLoggerFilenameTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		assert.Nil(t, err)
		got := chooseLoggerFilenameTestable("foo.log", mytime, 123)
		expected := "foo.06-01-02.123.log"
		assert.Equal(t, expected, got)
	}
}
func TestChooseLoggerFilenameLegacyTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		assert.Nil(t, err)
		got := chooseLoggerFilenameLegacyTestable("foo.log", mytime)
		expected := "foo.06-01-02.log"
		assert.Equal(t, expected, got)
	}
}
