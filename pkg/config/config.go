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

type TopStruct struct {
	Values   ValuesConfig
	Binaries BinaryPaths
	flat     map[string]string // someday maybe put all here?
}

func UpdateWithConfig(kv map[string]string, tc TopStruct) {
	for k, v := range tc.flat {
		kv[k] = v
	}
}

var top TopStruct // Should be considered "const", as changes would not be thread-safe.

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
	top = TopStruct{
		Binaries: FindBinaries(),
		Values: ValuesConfig{
			DefaultFrameRate: 100.0, // fps
		},
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
