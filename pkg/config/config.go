package config

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var Version string = "0.0.0-local-non-release"

type BinaryPaths struct {
	Binary_baz2bam         string
	Binary_pa_cal          string
	Binary_reducestats     string
	Binary_smrt_basecaller string
}

type ValuesConfig struct {
	DefaultFrameRate         float64 // fps
	JustOneBazFile           bool
	ApplyDarkCal             bool
	ApplyCrosstalkCorrection bool
	MovieNumberAlwaysZero    bool
	PawsTimeoutMultiplier    float64
}

//type StringMap map[string]string // would hide map as 'reference' type

type TopStruct struct {
	Values   ValuesConfig
	Binaries BinaryPaths
	Hostname string
	//flat     map[string]string // someday maybe put all here?
}

type PpaConfig struct {
	Webservices TopStruct `json:"webservices"`
}

var top *TopStruct   // Should be considered "const", as changes would not be thread-safe.
var ppatop PpaConfig // Should be considered "const", as changes would not be thread-safe.

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
	ppatop.Webservices = TopStruct{
		Binaries: FindBinaries(),
		Values: ValuesConfig{
			DefaultFrameRate:         100.0, // fps
			JustOneBazFile:           true,
			ApplyDarkCal:             true,
			ApplyCrosstalkCorrection: true,
			MovieNumberAlwaysZero:    true,
			PawsTimeoutMultiplier:    1.0,
		},
		Hostname: hostname,
	}
	top = &ppatop.Webservices
	/*
		top.flat = make(map[string]string)
		top.flat["Binary_baz2bam"] = top.Binaries.Binary_baz2bam
		top.flat["Binary_pa_cal"] = top.Binaries.Binary_pa_cal
		top.flat["Binary_reducestats"] = top.Binaries.Binary_reducestats
		top.flat["Binary_smrt_basecaller"] = top.Binaries.Binary_smrt_basecaller
	*/
}

// Make Top config const by returning only a copy.
func Top() TopStruct {
	return *top
}

func UpdateTop(r io.Reader) {
	s, err := io.ReadAll(r)
	check(err)
	Update(&ppatop, []byte(s))
}
func Update(p *PpaConfig, raw []byte) {
	ts := &p.Webservices
	log.Printf("Top Config was:\n%s", Config2Json(ts))
	err := json.Unmarshal(raw, p)
	if err != nil {
		log.Printf("raw JSON has some error:\n%s", raw)
		check(err)
	}
	ts = &p.Webservices
	//log.Printf("Top Config now:\n%s", Config2Json(ts))
	// Something else will dump config later.
}

/*
func PpaConfigJsonFromMap(opts map[string]string) string {
	var values_json strings.Builder
	var binaries_json strings.Builder
	log.Printf("map is (%+v)\n", opts)
	for k, v := range opts {
		log.Printf("--set (%s):(%s)\n", k, v)
		if strings.Contains(k, '.') {
			// Note: Any dots would imply full hiearchy corresponding to the Config file.
			// Dots are not supported yet.
			log.Printf(" We do not yet support hierarchy in single-settings. Skipping '%s:%s'")
		} else if strings.HasPrefix(k, "Binary_") {
			// Should be in Binaries section.
			fmt.Fprintf(&values_json, `,\n      "%s": "%s"`, k, v)
		} else {
			// Must be in Values section.
		}
	}
	var all_json strings.Builder
	fmt.Fprintf(&all_json, `{
  "webservices": {
    "Values": {%s},
    "Binaries": {%s}
  }
}`, values_json.String(), binaries_json.String())
	return all_json.String()
}
*/
// Panic if not understood as bool.
func String2Bool(v string) bool {
	v = strings.ToLower(v)
	if strings.HasPrefix(v, "0") {
		return false
	} else if strings.HasPrefix(v, "f") {
		return false
	} else if strings.HasPrefix(v, "n") {
		return false
	} else if strings.HasPrefix(v, "1") {
		return true
	} else if strings.HasPrefix(v, "t") {
		return true
	} else if strings.HasPrefix(v, "y") {
		return true
	}
	return false
}
func UpdateTopFromMap(opts map[string]string) {
	for k, v := range opts {
		log.Printf("--set (%s):(%s)\n", k, v)
		if strings.Contains(k, ".") {
			// Note: Any dots would imply full hiearchy corresponding to the Config file.
			// Dots are not supported yet.
			log.Printf(" We do not support hierarchy in single-settings. Skipping '%s:%s'",
				k, v)
			continue
		} else if k == "Binary_baz2bam" {
			top.Binaries.Binary_baz2bam = v
		} else if k == "Binary_smrt_basecaller" {
			top.Binaries.Binary_smrt_basecaller = v
		} else if k == "Binary_pa_cal" {
			top.Binaries.Binary_pa_cal = v
		} else if k == "Binary_reducestats" {
			top.Binaries.Binary_reducestats = v
		} else if k == "DefaultFrameRate" {
			fv, err := strconv.ParseFloat(v, 64)
			check(err)
			top.Values.DefaultFrameRate = fv
		} else if k == "PawsTimeoutMultiplier" {
			fv, err := strconv.ParseFloat(v, 64)
			check(err)
			top.Values.PawsTimeoutMultiplier = fv
		} else if k == "JustOneBazFile" {
			fv := String2Bool(v)
			top.Values.JustOneBazFile = fv
		} else if k == "ApplyDarkCal" {
			fv := String2Bool(v)
			top.Values.ApplyDarkCal = fv
		} else if k == "MovieNumberAlwaysZero" {
			fv := String2Bool(v)
			top.Values.MovieNumberAlwaysZero = fv
		} else if k == "ApplyCrosstalkCorrection" {
			fv := String2Bool(v)
			top.Values.ApplyCrosstalkCorrection = fv
		}
	}
}
func Config2Json(ts *TopStruct) string {
	result, err := json.MarshalIndent(ts, "", "  ")
	check(err)
	return string(result)
}
func TopAsJson() string {
	return Config2Json(top)
}
