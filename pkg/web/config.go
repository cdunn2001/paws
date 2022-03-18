package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// This might replace TopConfig soon. Not sure yet.
type PpaConfig struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reduce_stats    string
	Binary_smrt_basecaller string
}

func (cfg *PpaConfig) SetDefaults() {
	cfg.Binary_baz2bam = "baz2bam"
	cfg.Binary_smrt_basecaller = "smrt-basecaller"
	cfg.Binary_pa_cal = "pa-cal"
	cfg.Binary_reduce_stats = "reduce-stats"
}
func UpdatePpaConfig(raw []byte, current *PpaConfig) error {
	err := json.Unmarshal(raw, current)
	return err
}
func UpdatePpaConfigFromFile(fn string, current *PpaConfig) error {
	b, err := ioutil.ReadFile(fn)
	check(err)
	UpdatePpaConfig(b, current)
	fmt.Printf("Config now:%#v\n", *current)
	return nil
}

func Config2Json(cfg PpaConfig) string {
	raw, err := json.Marshal(cfg)
	check(err)
	return string(raw)
}
