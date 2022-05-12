package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"pacb.com/seq/paws/pkg/config"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// This is used only for tests. Ignored when "".
var DataDir = ""

func CreateTemplate(source string, name string) *template.Template {
	result := template.Must(template.
		New(name).
		Option("missingkey=error"). // https://pkg.go.dev/text/template#Template.Option
		Parse(source))
	return result
}

type TemplateSub struct {
	Global config.TopStruct
	Local  map[string]string
}

var Template_darkcal = `
{{.Global.Binaries.Binary_pa_cal}} \
  --statusfd 2 \
  --logoutput {{.Local.logoutput}} \
  --sra {{.Local.sra}} \
  --movieNum {{.Local.movieNum}} \
  --numFrames {{.Local.numFrames}} \
  --cal Dark \
  --outputFile {{.Local.outputFile}}  \
  --timeoutSeconds {{.Local.timeoutSeconds}} \
`

func WriteDarkcalBash(wr io.Writer, tc config.TopStruct, obj *SocketDarkcalObject, SocketId string, so *StorageObject) error {
	t := CreateTemplate(Template_darkcal, "")
	kv := make(map[string]string)

	sra := SocketId2Sra(SocketId)
	kv["sra"] = strconv.Itoa(sra)

	if tc.Values.MovieNumberAlwaysZero || obj.MovieNumber < 0 {
		kv["movieNum"] = "0"
	} else {
		kv["movieNum"] = strconv.Itoa(int(obj.MovieNumber))
	}

	numFrames := int(obj.MovieMaxFrames)
	kv["numFrames"] = strconv.Itoa(numFrames)
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = TranslateUrl(so, obj.CalibFileUrl)
	kv["logoutput"] = TranslateUrl(so, obj.LogUrl)

	timeout := float64(numFrames) * 1.1 / tc.Values.DefaultFrameRate // default
	if obj.MovieMaxSeconds > 0 {
		timeout = obj.MovieMaxSeconds
	}
	kv["timeoutSeconds"] = fmt.Sprintf("%g", timeout)
	// Skip --inputDarkCalFile can be skipped for now.

	ts := TemplateSub{
		Local:  kv,
		Global: tc,
	}
	return t.Execute(wr, &ts)
}

var Template_loadingcal = `
{{.Global.Binaries.Binary_pa_cal}} \
  --statusfd 2 \
  --logoutput {{.Local.logoutput}} \
  --sra {{.Local.sra}} \
  --movieNum {{.Local.movieNum}} \
  --numFrames {{.Local.numFrames}} \
  --cal Loading \
  --outputFile {{.Local.outputFile}}  \
  --inputDarkCalFile {{.Local.inputDarkCalFile}} \
  --timeoutSeconds {{.Local.timeoutSeconds}} \
`

func WriteLoadingcalBash(wr io.Writer, tc config.TopStruct, obj *SocketLoadingcalObject, SocketId string, so *StorageObject) error {
	t := CreateTemplate(Template_loadingcal, "")
	kv := make(map[string]string)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	if tc.Values.MovieNumberAlwaysZero || obj.MovieNumber < 0 {
		kv["movieNum"] = "0"
	} else {
		kv["movieNum"] = strconv.Itoa(int(obj.MovieNumber))
	}

	numFrames := int(obj.MovieMaxFrames)
	kv["numFrames"] = strconv.Itoa(numFrames)
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = TranslateUrl(so, obj.CalibFileUrl)
	kv["logoutput"] = TranslateUrl(so, obj.LogUrl)
	kv["inputDarkCalFile"] = TranslateUrl(so, obj.DarkFrameFileUrl)

	timeout := float64(numFrames) * 1.1 / tc.Values.DefaultFrameRate // default
	if obj.MovieMaxSeconds > 0 {
		timeout = obj.MovieMaxSeconds
	}
	kv["timeoutSeconds"] = fmt.Sprintf("%g", timeout)

	ts := TemplateSub{
		Local:  kv,
		Global: tc,
	}
	return t.Execute(wr, &ts)
}

// Translates the arguments into a command line option, or empty string if the URL is discardable.
//  ex. Translate("--outputtrcfile", "discard:", ) returns ""
//  ex. Translate("--foo","file:/bar") returns "--foo /bar"
// Otherwise return the flag with the translated path.
func TranslateDiscardableUrl(so *StorageObject, option string, url string) string {
	path := TranslateUrl(so, url)
	if path == "" {
		return ""
	} else if strings.Contains(option, "=") {
		return fmt.Sprintf("%s%s", option, path)
	} else {
		return fmt.Sprintf("%s %s", option, path)
	}
}

var Template_basecaller = `
export NUMA_NODE={{.Local.numa}}
export GPU_ID={{.Local.gpu}}
{{.Global.Binaries.Binary_smrt_basecaller}} \
  {{.Local.optMultiple}} \
  --statusfd 2 \
  {{.Local.optLogOutput}} \
  --logfilter INFO \
  {{.Local.optTraceFile}} \
  {{.Local.optTraceFileRoi}} \
  {{.Local.optOutputBazFile}} \
  {{.Local.optOutputRtMetricsFile}} \
  --config {{.Local.config_json_fn}} \
  --config source.WXIPCDataSourceConfig.sraIndex={{.Local.sra}} \
  {{.Local.optDarkCalFileName}} \
  {{.Local.optImagePsfKernel}} \
  {{.Local.optCrosstalkFilterKernel}} \
  --config system.analyzerHardware=A100 \
  --maxFrames {{.Local.maxFrames}} \
`

// optional:
//   system.analyzerHardware
//   algorithm

func WriteBasecallerBash(wr io.Writer, tc config.TopStruct, obj *SocketBasecallerObject, SocketId string, so *StorageObject) error {
	t := CreateTemplate(Template_basecaller, "")
	kv := make(map[string]string)

	var config_json_fn string
	{
		var outdir string
		if DataDir != "" {
			outdir = filepath.Join(DataDir, obj.Mid)
		} else {
			bazpath := TranslateUrl(so, obj.BazUrl)
			outdir = filepath.Dir(bazpath)
		}
		os.MkdirAll(outdir, 0777)
		config_json_fn = filepath.Join(outdir, obj.Mid+".basecaller.config.json")
	}

	// translate PAWS API to smrt-basecaller API JSON. UGH.
	var basecallerConfig SmrtBasecallerConfigObject
	acqConfig := &basecallerConfig.Source.WXIPCDataSourceConfig.AcqConfig
	for _, analog := range obj.Analogs {
		var a *Smrt_AnalogObject
		if analog.BaseLabel == "A" {
			a = &acqConfig.A
		} else if analog.BaseLabel == "C" {
			a = &acqConfig.C
		} else if analog.BaseLabel == "G" {
			a = &acqConfig.G
		} else if analog.BaseLabel == "T" {
			a = &acqConfig.T
		} else {
			a = nil
			log.Print("WARNING: Skipping analog ", analog)
		}
		if a != nil {
			a.BaseLabel = string(analog.BaseLabel)
			a.ExcessNoiseCV = analog.ExcessNoiseCv
			a.InterPulseDistance = analog.InterPulseDistanceSec
			a.Ipd2SlowStepRatio = analog.Ipd2SlowStepRatio
			a.PulseWidth = analog.PulseWidthSec
			a.Pw2SlowStepRatio = analog.Pw2SlowStepRatio
			a.RelAmplitude = analog.RelativeAmp
		}
	}
	acqConfig.RefSnr = obj.RefSnr
	acqConfig.PhotoelectronSensitivity = obj.PhotoelectronSensitivity
	chipLayout := obj.Chiplayout
	if chipLayout == "" || chipLayout == "Spider_1p0_NTO" {
		log.Printf("WARNING: Overriding bad or missing chipLayout name %q with %q",
			chipLayout, "KestrelRTO2")
		chipLayout = "KestrelRTO2"
	}
	acqConfig.ChipLayoutName = chipLayout

	basecallerConfigBytes, err := json.MarshalIndent(basecallerConfig, "", "    ")
	if err != nil {
		return err
	}
	basecallerConfigString := string(basecallerConfigBytes)

	WriteStringToFile(basecallerConfigString, config_json_fn)

	// Note: This file will be over-written on each call.

	sra := SocketId2Sra(SocketId)
	kv["sra"] = strconv.Itoa(sra)
	numNumaNodes := 2 // FIXME This works for rt-8400* machines, but doesn't work for kos-dev01 for example.
	numGpuNodes := 2  // FIXME

	kv["numa"] = strconv.Itoa(sra % numNumaNodes)
	kv["gpu"] = strconv.Itoa(sra % numGpuNodes)
	kv["config_json_fn"] = config_json_fn
	kv["maxFrames"] = strconv.Itoa(int(obj.MovieMaxFrames))

	raw, err := json.Marshal(obj.PixelSpreadFunction)
	if err != nil {
		return err
	}
	if len(raw) == 0 || string(raw) == "null" {
		kv["optImagePsfKernel"] = ""
	} else {
		kv["optImagePsfKernel"] = "--config dataSource.imagePsfKernel=" + string(raw)
	}

	raw, err = json.Marshal(obj.CrosstalkFilter)
	if err != nil {
		return err
	}
	if len(raw) == 0 || string(raw) == "null" {
		kv["optCrosstalkFilterKernel"] = ""
	} else {
		kv["optCrosstalkFilterKernel"] = "--config dataSource.crosstalkFilterKernel=" + string(raw)
	}

	if !tc.Values.ApplyCrosstalkCorrection {
		log.Printf("WARNING: imagePsfKernel is suppressed for basecaller.")
		kv["optImagePsfKernel"] = ""
		kv["optCrosstalkFilterKernel"] = ""
	}
	if !tc.Values.ApplyDarkCal {
		log.Printf("WARNING: darkCalFileName is suppressed for basecaller.")
		kv["optDarkCalFileName"] = ""
	} else {
		kv["optDarkCalFileName"] = TranslateDiscardableUrl(so, "--config dataSource.darkCalFileName=", obj.DarkCalFileUrl)
	}

	// TODO: Fill these from tc.Values first?
	if len(obj.TraceFileRoi) == 0 {
		kv["optTraceFile"] = ""
		kv["optTraceFileRoi"] = ""
	} else {
		optTraceFile := TranslateDiscardableUrl(so, "--outputtrcfile", obj.TraceFileUrl)
		kv["optTraceFile"] = optTraceFile

		raw, err := json.Marshal(obj.TraceFileRoi)
		if err != nil {
			return err
		}
		kv["optTraceFileRoi"] = "--config traceSaver.roi='" + string(raw) + "'"
	}
	kv["optOutputRtMetricsFile"] = "--config realTimeMetrics.jsonOutputFile=" + TranslateUrl(so, obj.RtMetrics.Url)
	if len(obj.BazUrl) == 0 {
		kv["optOutputBazFile"] = ""
	} else {
		kv["optOutputBazFile"] = TranslateDiscardableUrl(so, "--outputbazfile", obj.BazUrl)
	}
	if len(obj.LogUrl) == 0 {
		kv["optLogOutput"] = ""
	} else {
		kv["optLogOutput"] = TranslateDiscardableUrl(so, "--logoutput", obj.LogUrl)
	}
	if !strings.HasSuffix(kv["optLogOutput"], ".log") {
		msg := fmt.Sprintf("ERROR! For smrt-basecaller, log output is %q but must end w/ '.log' (for now).",
			kv["optLogOutput"])
		log.Printf(msg)
		//panic(msg)
	}

	optMultiple := ""
	if tc.Values.JustOneBazFile {
		optMultiple = "--config multipleBazFiles=false"
	}
	kv["optMultiple"] = optMultiple

	ts := TemplateSub{
		Local:  kv,
		Global: tc,
	}
	return t.Execute(wr, &ts)
}

// I don't expect you to have to change these.  These are Sequel-II
// parameters that may or may not be updated for Kestrel, but it's
// the kind of thing you'll hard code and never update again. -Ben
const (
	BAZ_THREADS = 32    // -j
	PBI_THREADS = 8     // -b
	OUT_QUEUE   = 15000 // --maxOutputQueueMB
	IN_QUEUE    = 7000  // --maxInputQueueMB
	BATCH_SIZE  = 50000 // --zmwBatchMB
	HEADER_SIZE = 30000 // --zmwHeaderBatchMB
)

var Template_baz2bam = `
{{.Global.Binaries.Binary_baz2bam}} \
  {{.Local.bazFile}} \
  {{.Local.logoutput}} \
  {{.Local.logfilter}} \
  -o {{.Local.outputPrefix}} \
  --statusfd 2 \
  {{.Local.metadata}} \
  --uuid {{.Local.acqId}} \
  -j 128 \
  -b 32 \
  --inlinePbi \
  --maxInputQueueMB 7000 \
  --zmwBatchMB 50000 \
  --zmwHeaderBatchMB 30000 \
  --maxOutputQueueMB 15000 \

  {{.Local.moveOutputStatsXml}}
  {{.Local.moveOutputStatsH5}}
`

// alternatively, replace bazFile(s) w/
// --filelist ${FILE_LIST}

// --silent //?

func MoveIfDifferent(implicitFn, desiredFn string) string {
	if desiredFn == "" || implicitFn == desiredFn || desiredFn == "discard:" {
		return ""
	}
	return fmt.Sprintf("mv -f '%s' '%s'", implicitFn, desiredFn)
}
func WriteMetadata(fn string, content string) {
	f, err := os.Create(fn)
	defer f.Close()
	if err != nil {
		msg := fmt.Sprintf("Could not open metadata file '%s' for write: %v", fn, err)
		log.Printf(msg)
		panic(msg)
	}
	f.WriteString(content)
}
func HasFullDataModel(content string) bool {
	// https://jira.pacificbiosciences.com/browse/ICS-1079
	return strings.Contains(content, "PacBioDataModel")
}
func HandleMetadata(fn string, content string) {
	gotFullDataModel := HasFullDataModel(content)
	if !gotFullDataModel {
		msg := fmt.Sprintf("We no longer support the partial datamodel for metadata. We require 'PacBioDataModel', not just a subreadset.\n%s", content)
		panic(msg)
	}
	log.Printf("Metadatafile: %q", fn)
	WriteMetadata(fn, content)
}
func GetPostprimaryHostname(hostname string, rundir string) string {
	if strings.HasPrefix(hostname, "rt") {
		// Substitute "nrt(a|b)" for "rt".
		ab := ""
		if strings.Contains(rundir, "nrta") {
			ab = "a"
		} else if strings.Contains(rundir, "nrtb") {
			ab = "b"
		} else {
			panic("rundir cant be deciphered" + rundir)
		}
		nrt := "nrt" + ab
		return nrt
	} else {
		return ""
	}
}
func DumpBash(setup ProcessSetupObject, bash string) {
	// For now, ignore RunDir. (Needs work.) Use cwd.
	//absRunDir, err := filepath.Abs(setup.RunDir)
	absRunDir, err := os.Getwd()
	check(err)
	content := "set -vex\n"
	content += "cd " + absRunDir + "\n"
	content += "export PATH=" + os.Getenv("PATH") + ":$PATH"
	content += bash
	WriteStringToFile(content, setup.ScriptFn)
}
func DumpBasecallerScript(tc config.TopStruct, obj *SocketBasecallerObject, sid string, so *StorageObject) ProcessSetupObject {
	setup := ProcessSetupObject{
		Tool: "smrt-basecaller",
	}

	// Choose and register any output paths first.
	mid := obj.Mid
	obj.BazUrl = ChooseUrlThenRegister(so, obj.BazUrl, StoragePathNrt, mid+".baz")
	obj.LogUrl = ChooseUrlThenRegister(so, obj.LogUrl, StoragePathNrt, mid+".basecaller.log")
	obj.TraceFileUrl = ChooseUrlThenRegister(so, obj.TraceFileUrl, StoragePathNrt, mid+".trc.h5")
	obj.RtMetrics.Url = ChooseUrlThenRegister(so, obj.RtMetrics.Url, StoragePathNrt, mid+".rtmetrics.json")

	// Now we can use the output Urls.
	var rundir string
	if obj.BazUrl != "discard:" && obj.BazUrl != "/dev/null" {
		rundir = filepath.Dir(TranslateUrl(so, obj.BazUrl))
	} else if obj.TraceFileUrl != "discard:" && obj.TraceFileUrl != "/dev/null" {
		rundir = filepath.Dir(TranslateUrl(so, obj.TraceFileUrl))
	} else {
		rundir = filepath.Join("/tmp", "pawsgo", sid, obj.Mid)
	}
	setup.RunDir = rundir
	setup.ScriptFn = filepath.Join(setup.RunDir, "run.basecaller.sh")
	setup.Hostname = ""
	wr := new(bytes.Buffer)
	if err := WriteBasecallerBash(wr, config.Top(), obj, sid, so); err != nil {
		err = errors.Wrapf(err, "Error in WriteBasecallerBash(%v, %v, %v, %v)", wr, config.Top(), obj, sid)
		check(err)
	}
	DumpBash(setup, wr.String())
	return setup
}
func DumpDarkcalScript(tc config.TopStruct, obj *SocketDarkcalObject, sid string, so *StorageObject) ProcessSetupObject {
	setup := ProcessSetupObject{
		Tool: "darkcal",
	}

	// Choose and register any output paths first.
	mid := obj.Mid
	obj.CalibFileUrl = ChooseUrlThenRegister(so, obj.CalibFileUrl, StoragePathIcc, mid+".darkcal.h5")
	obj.LogUrl = ChooseUrlThenRegister(so, obj.LogUrl, StoragePathIcc, mid+".darkcal.log")

	// Now we can use the output Urls.
	rundir := filepath.Dir(TranslateUrl(so, obj.CalibFileUrl))
	setup.RunDir = rundir
	setup.ScriptFn = filepath.Join(setup.RunDir, "run.darkcal.sh")
	setup.Hostname = ""
	wr := new(bytes.Buffer)
	if err := WriteDarkcalBash(wr, config.Top(), obj, sid, so); err != nil {
		err = errors.Wrapf(err, "Error in WriteDarkcalBash(%v, %v, %v, %v)", wr, config.Top(), obj, sid)
		check(err)
	}
	DumpBash(setup, wr.String())
	return setup
}
func UniqueLabel(so *StorageObject) string {
	prev := so.Counter
	so.Counter++
	return fmt.Sprintf(".%02d", prev)
}
func DumpLoadingcalScript(tc config.TopStruct, obj *SocketLoadingcalObject, sid string, so *StorageObject) ProcessSetupObject {
	setup := ProcessSetupObject{
		Tool: "loadingcal",
	}

	// Choose and register any output paths first.
	mid := obj.Mid
	ul := UniqueLabel(so)
	obj.CalibFileUrl = ChooseUrlThenRegister(so, obj.CalibFileUrl, StoragePathIcc, mid+ul+".loadingcal.h5")
	obj.LogUrl = ChooseUrlThenRegister(so, obj.LogUrl, StoragePathIcc, mid+ul+".loadingcal.log")

	// Now we can use the output Urls.
	rundir := filepath.Dir(TranslateUrl(so, obj.CalibFileUrl))
	setup.RunDir = rundir
	setup.ScriptFn = filepath.Join(setup.RunDir, "run.loadingcal.sh")
	setup.Hostname = ""
	wr := new(bytes.Buffer)
	if err := WriteLoadingcalBash(wr, config.Top(), obj, sid, so); err != nil {
		err = errors.Wrapf(err, "Error in WriteLoadingcalBash(%v, %v, %v, %v)", wr, config.Top(), obj, sid)
		check(err)
	}
	DumpBash(setup, wr.String())
	return setup
}
func DumpPostprimaryScript(tc config.TopStruct, obj *PostprimaryObject, so *StorageObject) ProcessSetupObject {
	setup := ProcessSetupObject{
		Tool: "ppa(baz2bam)",
	}

	// Choose and register any output paths first.
	mid := obj.Mid
	obj.OutputPrefixUrl = ChooseUrlThenRegister(so, obj.OutputPrefixUrl, StoragePathIcc, mid)
	obj.LogUrl = ChooseUrlThenRegister(so, obj.LogUrl, StoragePathIcc, mid+".baz2bam.log")
	//log.Printf("OutputPrefixUrl: %q", obj.OutputPrefixUrl)

	//obj.LogReducdStatsUrl = ChooseUrlThenRegister(so, obj.LogReduceStatsUrl, StoragePathIcc, mid+".reducestats.log")
	//logoutput := TranslateUrl(so, obj.LogReduceStatsUrl)

	OutputStatsH5Url := obj.OutputStatsH5Url
	if OutputStatsH5Url == "" {
		OutputStatsH5Url = obj.OutputPrefixUrl + ".sts.h5"
	}
	OutputStatsH5Url = ChooseUrlThenRegister(so, OutputStatsH5Url, StoragePathIcc, mid+".sts.h5")
	obj.OutputStatsH5Url = OutputStatsH5Url
	OutputReduceStatsH5Url := obj.OutputReduceStatsH5Url
	if OutputReduceStatsH5Url == "" {
		OutputReduceStatsH5Url = obj.OutputPrefixUrl + ".rsts.h5"
	}
	OutputReduceStatsH5Url = ChooseUrlThenRegister(so, OutputReduceStatsH5Url, StoragePathIcc, mid+".rsts.h5")
	//log.Printf("OutputReduceStatsH5Url: %q", OutputReduceStatsH5Url)
	obj.OutputReduceStatsH5Url = OutputReduceStatsH5Url

	// Now we can use the output Urls.
	rundir := filepath.Dir(TranslateUrl(so, obj.OutputPrefixUrl))
	setup.RunDir = rundir
	setup.ScriptFn = filepath.Join(setup.RunDir, "run.ppa.sh")
	setup.Hostname = GetPostprimaryHostname(tc.Hostname, TranslateUrl(so, obj.BazFileUrl))
	wr := new(bytes.Buffer)
	if err := WriteBaz2bamBash(wr, config.Top(), obj, so); err != nil {
		err = errors.Wrapf(err, "Error in WriteBaz2BamBash(%v, %v, %v)", wr, config.Top(), obj)
		check(err)
	}
	if err := WriteReduceStatsBash(wr, config.Top(), obj, so); err != nil {
		err = errors.Wrapf(err, "Error in WriteReduceStatsBash(%v, %v, %v)", wr, config.Top(), obj)
		check(err)
	}
	DumpBash(setup, wr.String())
	return setup
}
func WriteBaz2bamBash(wr io.Writer, tc config.TopStruct, obj *PostprimaryObject, so *StorageObject) error {
	t := CreateTemplate(Template_baz2bam, "")
	kv := make(map[string]string)

	// Now we can use the output Urls.
	outputPrefix := TranslateUrl(so, obj.OutputPrefixUrl)
	kv["outputPrefix"] = outputPrefix
	{
		outdir := filepath.Dir(outputPrefix)
		if outdir == "" {
			return errors.Errorf("Got empty dir for OutputPrefixUrl '%s'", outputPrefix)
		}
		os.MkdirAll(outdir, 0777)
	}
	metadata_xml := outputPrefix + ".metadata.xml"
	HandleMetadata(metadata_xml, obj.SubreadsetMetadataXml)
	kv["metadata"] = "--metadata " + metadata_xml
	kv["acqId"] = obj.Uuid
	kv["bazFile"] = TranslateUrl(so, obj.BazFileUrl)
	loglevel := obj.LogLevel
	logoutput := TranslateUrl(so, obj.LogUrl)
	if loglevel == "" {
		kv["logfilter"] = ""
	} else {
		kv["logfilter"] = "--logfilter " + string(loglevel)
	}
	kv["logoutput"] = "--logoutput " + logoutput
	// Note: baz2bam will actually append to this, not over-write,
	// as of e78d019b (9c896e1e), 11Apr2022.

	kv["moveOutputStatsXml"] = MoveIfDifferent(obj.OutputPrefixUrl+".sts.xml",
		obj.OutputStatsXmlUrl)
	kv["moveOutputStatsH5"] = MoveIfDifferent(obj.OutputPrefixUrl+".sts.h5",
		obj.OutputStatsH5Url)
	//kv["baz2bamComputingThreads"] = "16"
	//kv["bamThreads"] = "16"
	//kv["inlinePbi"] = "true"
	//kv["maxInputQueueMB"] = "39"
	//kv["zmwBatchMB"] = "40"
	//kv["headerBatchMB"] = "41"
	//kv["baz2BamMaxOutputQueueMB"] = "42"

	// --progress # for IPC messages
	// --silent   # do we want this?
	// ppaConfig.Baz2BamArgs();
	// This envar is not to be used except for unit testing.
	// getenv("PPA_BAZ2BAM_OPTIONS");

	ts := TemplateSub{
		Local:  kv,
		Global: tc,
	}
	return t.Execute(wr, &ts)
}

var Template_reducestats = `
echo 'PA_PPA_STATUS {"counter":0,"counterMax":1,"stageName":"Bash","stageNumber":0,"stageWeights":[100],"state":"progress","timeoutForNextStatus":300}' >&2

{{.Global.Binaries.Binary_reducestats}} \
  {{.Local.OutputStatsH5}} \
  {{.Local.OutputReduceStatsH5}} \
`

// Skip --logoutput for now.

func WriteReduceStatsBash(wr io.Writer, tc config.TopStruct, obj *PostprimaryObject, so *StorageObject) error {
	t := CreateTemplate(Template_reducestats, "")
	kv := make(map[string]string)

	if obj.OutputStatsH5Url == "discard:" {
		kv["OutputStatsH5"] = ""
	} else {
		kv["OutputStatsH5"] = "--input " + TranslateUrl(so, obj.OutputStatsH5Url)
		// Output from baz2bam; Input for reducestats. But Output of PPA.
	}

	if obj.OutputReduceStatsH5Url == "discard:" {
		kv["OutputReduceStatsH5"] = ""
	} else {
		kv["OutputReduceStatsH5"] = "--output " + TranslateUrl(so, obj.OutputReduceStatsH5Url)
	}

	ts := TemplateSub{
		Local:  kv,
		Global: tc,
	}
	return t.Execute(wr, &ts)
}
func CheckBaz2bam(tc config.TopStruct) {
	call := fmt.Sprintf("which %s; %s --version",
		tc.Binaries.Binary_baz2bam,
		tc.Binaries.Binary_baz2bam)
	//hostname = GetPostprimaryHostname(tc.Hostname, "/data/nrta")
	//captured := CaptureRemoteBash(hostname, call)
	captured := CaptureBash(call)
	log.Printf("Captured:%s\n", captured)
	os.Exit(0)
}

// TODO: reimplement
func SocketId2Sra(socketId string) int {
	socketIdInt, err := strconv.Atoi(socketId)
	if err != nil {
		msg := fmt.Sprintf("Must provide a valid socketId, not %q: %v", socketId, err)
		panic(msg)
	}
	sra := socketIdInt - 1
	return sra
}
