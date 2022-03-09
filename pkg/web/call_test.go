package web

import (
	"fmt"
	"os"
	"testing"
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
	got := WatchBash(bash, env)
	if got != nil {
		t.Errorf("Got %d", got)
	}
}
func TestString2StatusReport(t *testing.T) {
	sr, err := String2StatusReport(`_STATUS {"counter": 123}`)
	check(err)
	if sr.normal.Counter != 123 {
		t.Errorf("Got %d", sr.normal.Counter)
	}

}
