package config

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	PAWS_TEST_LOG := os.Getenv("PAWS_TEST_LOG")
	//fmt.Printf("PAWS_TEST_LOG=%s\n", PAWS_TEST_LOG)
	if PAWS_TEST_LOG == "" {
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}
