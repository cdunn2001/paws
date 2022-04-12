package config

import (
	"os"
)

var Version string = "0.0.0-local-non-release"

type BinaryPaths struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reducestats     string
	Binary_smrt_basecaller string
}

type ValuesConfig struct {
	DefaultFrameRate float64 // fps
	JustOneBazFile   bool
}

//type StringMap map[string]string // would hide map as 'reference' type

type TopStruct struct {
	Values   ValuesConfig
	Binaries BinaryPaths
	Hostname string
	flat     map[string]string // someday maybe put all here?
}

var top TopStruct // Should be considered "const", as changes would not be thread-safe.

// TODO: Allow config override.
func FindBinaries() BinaryPaths {
	return BinaryPaths{
		Binary_baz2bam:         "baz2bam",
		Binary_smrt_basecaller: "smrt-basecaller-launch.sh", // this script is necessary to configure NUMA. don't call smrt-basecaller binary directly.
		Binary_pa_cal:          "pa-cal",
		Binary_reducestats:     "ppa-reducestats",
	}
}

func init() {
	hostname, err := os.Hostname()
	check(err)
	top = TopStruct{
		Binaries: FindBinaries(),
		Values: ValuesConfig{
			DefaultFrameRate: 100.0, // fps
			JustOneBazFile:   true,
		},
		Hostname: hostname,
	}
	top.flat = make(map[string]string)
	top.flat["Binary_baz2bam"] = top.Binaries.Binary_baz2bam
	top.flat["Binary_pa_cal"] = top.Binaries.Binary_pa_cal
	top.flat["Binary_reducestats"] = top.Binaries.Binary_reducestats
	top.flat["Binary_smrt_basecaller"] = top.Binaries.Binary_smrt_basecaller
}

// Make Top config const by returning only a copy.
func Top() TopStruct {
	return top
}
