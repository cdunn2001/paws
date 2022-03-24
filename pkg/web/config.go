package web

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// This might replace config.TopStruct soon. Not sure yet.
type PpaConfig struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reducestats     string
	Binary_smrt_basecaller string
}

func (cfg *PpaConfig) SetDefaults() {
	cfg.Binary_baz2bam = "baz2bam"
	cfg.Binary_smrt_basecaller = "smrt-basecaller"
	cfg.Binary_pa_cal = "pa-cal"
	cfg.Binary_reducestats = "ppa-reducestats"
}
func UpdatePpaConfig(raw []byte, current *PpaConfig) error {
	err := json.Unmarshal(raw, current)
	return err
}
func UpdatePpaConfigFromFile(fn string, current *PpaConfig) error {
	b, err := ioutil.ReadFile(fn)
	check(err)
	UpdatePpaConfig(b, current)
	log.Printf("Config now:%#v\n", *current)
	return nil
}

func Config2Json(cfg PpaConfig) string {
	raw, err := json.Marshal(cfg)
	check(err)
	return string(raw)
}
