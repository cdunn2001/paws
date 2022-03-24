package config

type BinaryPaths struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reducestats     string
	Binary_smrt_basecaller string
}

type ValuesConfig struct {
	DefaultFrameRate float64 // fps
}

//type StringMap map[string]string // would hide map as 'reference' type

type TopConfigStruct struct {
	Values   ValuesConfig
	Binaries BinaryPaths
	flat     map[string]string // someday maybe put all here?
}

func UpdateWithConfig(kv map[string]string, tc *TopConfigStruct) {
	for k, v := range tc.flat {
		kv[k] = v
	}
}

var TopConfig TopConfigStruct // Should be considered "const", as changes would not be thread-safe.

func FindBinaries() BinaryPaths {
	// TODO: Replace w/ PpaConfig
	return BinaryPaths{
		Binary_baz2bam:         "baz2bam",
		Binary_smrt_basecaller: "smrt-basecaller-launch.sh", // this script is necessary to configure NUMA. don't call smrt-basecaller binary directly.
		Binary_pa_cal:          "pa-cal",
		Binary_reducestats:     "ppa-reducestats",
	}
}

func init() {
	// TODO: These should be configurable.
	TopConfig = TopConfigStruct{
		binaries: FindBinaries(),
		values: ValuesConfig{
			defaultFrameRate: 100.0, // fps
		},
	}
	TopConfig.flat = make(map[string]string)
	TopConfig.flat["Binary_baz2bam"] = TopConfig.binaries.Binary_baz2bam
	TopConfig.flat["Binary_pa_cal"] = TopConfig.binaries.Binary_pa_cal
	TopConfig.flat["Binary_reducestats"] = TopConfig.binaries.Binary_reducestats
	TopConfig.flat["Binary_smrt_basecaller"] = TopConfig.binaries.Binary_smrt_basecaller
}
