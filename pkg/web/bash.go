package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

func CreateTemplate(source string, name string) *template.Template {
	result := template.Must(template.
		New(name).
		Option("missingkey=error"). // https://pkg.go.dev/text/template#Template.Option
		Parse(source))
	return result
}

var Template_darkcal = `
{{.Binary_pa_cal}} \
  --statusfd 3 \
  --logoutput {{.logoutput}} \
  --sra {{.sra}} \
  --movieNum {{.movieNum}} \
  --numFrames {{.numFrames}} \
  --cal Dark \
  --outputFile {{.outputFile}}  \
  --timeoutSeconds {{.timeoutSeconds}} \
`

func WriteDarkcalBash(wr io.Writer, tc *TopConfig, obj *SocketDarkcalObject, SocketId string) error {
	t := CreateTemplate(Template_darkcal, "")
	kv := make(map[string]string)
	UpdateWithConfig(kv, tc)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	kv["movieNum"] = "0" // for now
	// assert if obj.movieNum not nil, then it is 0.

	numFrames := int(obj.MaxMovieFrames)
	kv["numFrames"] = strconv.Itoa(numFrames)
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = obj.CalibFileUrl // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl        // TODO: Convert from URL!

	timeout := float64(numFrames) * 1.1 / tc.values.defaultFrameRate // default
	if obj.MaxMovieSeconds > 0 {
		timeout = obj.MaxMovieSeconds
	}
	kv["timeoutSeconds"] = fmt.Sprintf("%g", timeout)

	// Skip --inputDarkCalFile can be skipped for now.
	return t.Execute(wr, kv)
}

var Template_loadingcal = `
{{.Binary_pa_cal}} \
  --statusfd 3 \
  --logoutput {{.logoutput}} \
  --sra {{.sra}} \
  --movieNum {{.movieNum}} \
  --numFrames {{.numFrames}} \
  --cal Loading \
  --outputFile {{.outputFile}}  \
  --inputDarkCalFile {{.inputDarkCalFile}} \
  --timeoutSeconds {{.timeoutseconds}} \
`

func WriteLoadingcalBash(wr io.Writer, tc *TopConfig, obj *SocketLoadingcalObject, SocketId string) error {
	t := CreateTemplate(Template_loadingcal, "")
	kv := make(map[string]string)

	UpdateWithConfig(kv, tc)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	kv["movieNum"] = "0" // for now
	// assert if obj.movieNum not nil, then it is 0.

	numFrames := int(obj.MaxMovieFrames)
	kv["numFrames"] = strconv.Itoa(numFrames)
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = obj.CalibFileUrl           // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl                  // TODO: Convert from URL!
	kv["inputDarkCalFile"] = obj.DarkFrameFileUrl // TODO: Convert from URL!

	timeout := float64(numFrames) * 1.1 / tc.values.defaultFrameRate // default
	if obj.MaxMovieSeconds > 0 {
		timeout = obj.MaxMovieSeconds
	}
	kv["timeoutSeconds"] = fmt.Sprintf("%g", timeout)

	return t.Execute(wr, kv)
}

var defaultBasecallerConfig = `
{
	"source" :
	{
		"WXIPCDataSourceConfig" :
		{
			"acqConfig" :
			{
				"A" :
				{
					"baseLabel" : "A",
					"excessNoiseCV" : 0.1,
					"interPulseDistance" : 0.08,
					"ipd2SlowStepRatio" : 0,
					"pulseWidth" : 0.166,
					"pw2SlowStepRatio" : 3.2,
					"relAmplitude" : 0.67
				},
				"C" :
				{
					"baseLabel" : "C",
					"excessNoiseCV" : 0.1,
					"interPulseDistance" : 0.07,
					"ipd2SlowStepRatio" : 0,
					"pulseWidth" : 0.209,
					"pw2SlowStepRatio" : 3.2,
					"relAmplitude" : 1.0
				},
				"G" :
				{
					"baseLabel" : "G",
					"excessNoiseCV" : 0.1,
					"interPulseDistance" : 0.07,
					"ipd2SlowStepRatio" : 0,
					"pulseWidth" : 0.193,
					"pw2SlowStepRatio" : 3.2,
					"relAmplitude" : 0.26
				},
				"T" :
				{
					"baseLabel" : "T",
					"excessNoiseCV" : 0.1,
					"interPulseDistance" : 0.08,
					"ipd2SlowStepRatio" : 0,
					"pulseWidth" : 0.163,
					"pw2SlowStepRatio" : 3.2,
					"relAmplitude" : 0.445
				},
				"refSnr" : 12
			}
		}
	}
}
`

func CopyDefaultBasecallerConfig(dest_fn string) {
	log.Printf("Copy basecaller config to file '%s'", dest_fn)
	WriteStringToFile(defaultBasecallerConfig, dest_fn)
}
func TranslateDiscardableUrl(option string, url string) string {
	// ex. Translate("discard:", "--outputtrcfile")
	// If "discard:", then return "".
	// Otherwise return the flag with the translated path.
	if url == "discard:" {
		return ""
	} else {
		// TODO: Convert from URL!
		return fmt.Sprintf("%s %s", option, url)
	}
}

var Template_basecaller = `
{{.Binary_smrt_basecaller}} \
  --statusfd 3 \
  --logoutput {{.logoutput}} \
  --logfilter INFO \
  {{.optTraceFile}} \
  {{.optTraceFileRoi}} \
  --outputbazfile {{.outputbazfile}} \
  --config {{.config_json_fn}} \
  --config source.WXIPCDataSourceConfig.sraIndex={{.sra}} \
  --config system.analyzerHardware=A100 \
  --maxFrames {{.maxFrames}} \
`

// Maybe better:
// --config source.WXIPCDataSourceConfig.acqConfig=Info-About-Chemistry \

// optional:
//   system.analyzerHardware
//   algorithm

// Doesn't this need the darkcalfile?
func WriteBasecallerBash(wr io.Writer, tc *TopConfig, obj *SocketBasecallerObject, SocketId string) error {
	t := CreateTemplate(Template_basecaller, "")
	kv := make(map[string]string)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	sraName := strconv.Itoa(sra)

	UpdateWithConfig(kv, tc)

	outdir := filepath.Join("/data/nrta", sraName)
	os.MkdirAll(outdir, 0777)
	config_json_fn := filepath.Join(outdir, obj.Mid+".basecaller.config.json")
	CopyDefaultBasecallerConfig(config_json_fn)
	// Note: This file will be over-written on each call.

	kv["sra"] = strconv.Itoa(sra)
	kv["config_json_fn"] = config_json_fn
	kv["outputbazfile"] = obj.BazUrl // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl     // TODO: Convert from URL!
	kv["maxFrames"] = strconv.Itoa(int(obj.MaxMovieFrames))

	if kv["optTraceFileRoi"] == "" || kv["optTraceFileRoi"] == "[]" {
		kv["optTraceFile"] = ""
		kv["optTraceFileRoi"] = ""
	} else {
		optTraceFile := TranslateDiscardableUrl("--outputtrcfile", obj.TraceFileUrl)
		kv["optTraceFile"] = optTraceFile

		raw, err := json.Marshal(obj.TraceFileRoi)
		check(err)
		kv["optTraceFileRoi"] = "--traceFileRoi=" + string(raw)
	}

	return t.Execute(wr, kv)
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
{{.Binary_baz2bam}} \
  {{.bazFile}} \
  --statusfd 3 \
  --metadata {{.metadataFile}} \
  --uuid {{.acqId}} \
  -j 32 \
  -b 8 \
  --inlinePbi \
  --maxInputQueueMB 7000 \
  --zmwBatchMB 50000 \
  --zmwHeaderBatchMB 30000 \
  --maxOutputQueueMB 15000 \
`

//  -o OUT_SUFFIX (e.g. 'out')

// alternatively, replace bazFile(s) w/
// --filelist ${FILE_LIST}

// --silent //?

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
func WriteBaz2bamBash(wr io.Writer, tc *TopConfig, obj *PostprimaryObject) error {
	t := CreateTemplate(Template_baz2bam, "")
	kv := make(map[string]string)
	UpdateWithConfig(kv, tc)
	outdir := obj.OutputPrefixUrl // TODO: Translate URL
	os.MkdirAll(outdir, 0777)
	metadata_xml := filepath.Join(outdir, obj.Mid+".metadata.subreadset.xml")
	WriteMetadata(metadata_xml, obj.SubreadsetMetadataXml)
	kv["metadataFile"] = metadata_xml
	kv["acqId"] = obj.Uuid
	kv["bazFile"] = obj.BazFileUrl // TODO
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
	return t.Execute(wr, kv)
}

type Job struct {
	outputPrefix string
	chipClass    string
	platform     string
}

func UpdateJob(kv map[string]string, job Job) {
	// TODO: Where should these come from? (outputPrefix is on PpObj.)
	kv["job_outputPrefix"] = job.outputPrefix
	kv["job_chipClass"] = job.chipClass
	kv["job_platform"] = job.platform
}

var Template_reducestats = `
{{.Binary_reducestats}} \
  --input {{.job_outputPrefix}}.sts.h5 \
  --output {{.job_outputPrefix}}.rsts.h5 \
  --config=common.chipClass=Kestrel \
  --config=common.platform=Kestrel \
`

func WriteReduceStatsBash(wr io.Writer, tc *TopConfig, obj *PostprimaryObject, job Job) error {
	t := CreateTemplate(Template_reducestats, "")
	kv := make(map[string]string)
	UpdateWithConfig(kv, tc)
	job.outputPrefix = obj.OutputPrefixUrl // TODO
	UpdateJob(kv, job)
	//obj.OutputReduceStatsH5Url
	return t.Execute(wr, kv)
}
func CopyRsts(obj *PostprimaryObject, job Job) error {
	// obj.OutputStatsH5Url
	// obj.OutputStatsXmlUrl
	/*
		void PpaControllerOld::CopyRsts(const PPAJob& job)
		{
		    std::stringstream ss;
		    const std::string movieContext = job.movieContext;
		    const std::string rstsFilename = job.outputPrefix + ".rsts.h5";
		    if (PacBio::POSIX::IsFile(rstsFilename))
		    {
		        ss << "scp -o StrictHostKeyChecking=no " << rstsFilename
		           << " " << ppaConfig_.RstsDestinationPrefix() << "/" + movieContext;
		        PBLOG_INFO << ss.str();
		        const std::string capturedStdout = PacBio::System::Run(ss.str());
		        PBLOG_INFO << capturedStdout;

		        if (true)
		        {
		            if (unlink(rstsFilename.c_str()))
		            {
		                errors_++;
		                PBLOG_ERROR << "Could not delete " << rstsFilename;
		            }
		        }
		    }
		    else
		    {
		        PBLOG_WARN << "Won't copy and delete " << rstsFilename << " because it doesn't exist";
		    }
		}
	*/
	return nil
}
