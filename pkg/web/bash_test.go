package web

import (
	"bytes"
	"pacb.com/seq/paws/pkg/config"
	"testing"
)

func TestWriteReduceStatsBash(t *testing.T) {
	expected := `
ppa-reducestats \
  --input PREFIX.sts.h5 \
  --output PREFIX.rsts.h5 \
  --config=common.chipClass=Kestrel \
  --config=common.platform=Kestrel \
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
	tc := config.Top()
	err := WriteReduceStatsBash(&b, tc, obj, job)
	check(err)
	got := b.String()
	if got != expected {
		t.Errorf("Got %s", got)
	}
}
