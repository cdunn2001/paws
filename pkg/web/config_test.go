package web

import (
	"testing"
)

func Expect(t *testing.T, expected string, got string) {
	if got != expected {
		t.Errorf("Expected:\n%v\nGot:\n%#v", got, expected)
	}
}
func TestPpaConfigUpdate(t *testing.T) {
	cfg := &PpaConfig{}
	cfg.SetDefaults()
	raw := []byte(`{"Binary_baz2bam": "SNAFU"}`)
	err := UpdatePpaConfig(raw, cfg)
	check(err)
	got := Config2Json(*cfg)
	expected := `{"Binary_baz2bam":"SNAFU","Binary_pa_cal":"pa-cal","Binary_reduce_stats":"reduce-stats","Binary_smrt_basecaller":"smrt-basecaller"}`
	Expect(t, expected, got)
}
