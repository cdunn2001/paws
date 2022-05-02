package web

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var testtmpdir string

func TestMain(m *testing.M) {
	var err error
	testtmpdir, err = ioutil.TempDir("", "pawsgo.*.tmpdir")
	check(err)
	PAWS_TEST_LOG := os.Getenv("PAWS_TEST_LOG")
	fmt.Printf("PAWS_TEST_LOG=%s\n", PAWS_TEST_LOG)
	if PAWS_TEST_LOG == "" {
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("testtmpdir=%q", testtmpdir)
	rc := m.Run()
	if PAWS_TEST_LOG == "" {
		os.RemoveAll(testtmpdir)
	}
	os.Exit(rc)
}
