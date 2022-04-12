package config

import (
	//"bytes"
	//"strings"
	"testing"
)

func Expect(t *testing.T, expected string, got string) {
	if got != expected {
		t.Errorf("Expected:\n%#v\nGot:\n%#v", expected, got)
	}
}

var sample_json string = `
{
  "basecaller": {
    "init": {
      "numWorkerThreads_avx512": 88
    }
  },
  "webservices": {
	"Values": {
      "DefaultFrameRate": 99.0,
      "JustOneBazFile": false,
      "ApplyDarkCal": true,
      "ApplyCrosstalkCorrection": false
    },
    "Binaries": {
      "Binary_baz2bam": "NEWPATH"
    },
    "IgnoreMe": null
  },
  "ppa": {
  }
}
`

func TestUpdate(t *testing.T) {
	cfg := &PpaConfig{}
	cfg.Webservices.Values.DefaultFrameRate = 100.0
	//buf := bytes.NewBufferString(sample_json)
	//buf := strings.NewReader(sample_json)
	Update(cfg, []byte(sample_json))
	got := Config2Json(&cfg.Webservices)
	expected := `{
  "Values": {
    "DefaultFrameRate": 99,
    "JustOneBazFile": false,
    "ApplyDarkCal": true,
    "ApplyCrosstalkCorrection": false,
    "PawsTimeoutMultiplier": 0
  },
  "Binaries": {
    "Binary_baz2bam": "NEWPATH",
    "Binary_pa_cal": "",
    "Binary_reducestats": "",
    "Binary_smrt_basecaller": ""
  },
  "Hostname": ""
}`
	Expect(t, expected, got)
}
