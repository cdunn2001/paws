package web

import (
	"bytes"
	"testing"
)

func TestWriteReduceStatsBash(t *testing.T) {
	expected := `
reduce-stats \
  --input PREFIX.sts.h5 \
  --output PREFIX.rsts.h5 \
  --config=common.chipClass=CHIP \
  --config=common.platform=PLATFORM \
`
	obj := &PostprimaryObject{
		OutputPrefixUrl: "PREFIX",
	}
	job := Job{
		outputPrefix: "PREFIX",
		chipClass:    "CHIP",
		platform:     "PLATFORM",
	}
	var b bytes.Buffer
	tc := &topconfig
	err := WriteReduceStatsBash(&b, tc, obj, job)
	check(err)
	got := b.String()
	if got != expected {
		t.Errorf("Got %s", got)
	}
}
