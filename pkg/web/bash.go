package web

import (
	"io"
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
  --timeoutseconds {{.timeoutseconds}} \
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

	timeout := int(float64(numFrames) * 1.1 / tc.values.defaultFrameRate) // default
	if obj.MaxMovieSeconds != 0 {
		timeout = int(obj.MaxMovieSeconds)
	}
	kv["timeoutseconds"] = strconv.Itoa(timeout)

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
  --timeoutseconds {{.timeoutseconds}} \
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

	timeout := int(float64(numFrames) * 1.1 / tc.values.defaultFrameRate) // default
	if obj.MaxMovieSeconds != 0 {
		timeout = int(obj.MaxMovieSeconds)
	}
	kv["timeoutseconds"] = strconv.Itoa(timeout)

	return t.Execute(wr, kv)
}

var Template_basecaller = `
{{.Binary_smrt_basecaller}} \
  --statusfd 3 \
  --logoutput {{.logoutput}} \
  --logfilter INFO \
  --outputtrcfile {{.outputtrcfile}} \
  --outputbazfile {{.outputbazfile}} \
  --config source.WXIPCDataSource.sraIndex={{.sra}} \
  --config traceSaver.roi=roi_specification \
  --config source.WXIPCDataSource.acqConfig=Info-About-Chemistry \
  --config system.analyzerHardware=A100 \
  --config algorithm=forward-from-user \
`

// optional:
//   system.analyzerHardware
//   algorithm

// Doesn't this need the darkcalfile?
func WriteBasecallerBash(wr io.Writer, tc *TopConfig, obj *SocketBasecallerObject, SocketId string) error {
	t := CreateTemplate(Template_basecaller, "")
	kv := make(map[string]string)

	UpdateWithConfig(kv, tc)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	kv["outputtrcfile"] = obj.TraceFileUrl // TODO: Convert from URL!
	kv["outputbazfile"] = obj.BazUrl       // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl           // TODO: Convert from URL!

	// Skip --maxFrames for now?

	return t.Execute(wr, kv)
}

var Template_baz2bam = `
{{.Binary_baz2bam}} \
  {{.bazFile}} \
  --statusfd 3 \
  --metadata {{.metadataFile}} \
  --uuid {{.acqId}} \
  -j {{.baz2bamComputingThreads}} \
  -b {{.bamThreads}} \
  {{if .inlinePbi}}--inlinePbi{{end}} \
  --maxInputQueueMB {{.maxInputQueueMB}} \
  --zmwBatchMB {{.zmwBatchMB}} \
  --zmwHeaderBatchMB {{.headerBatchMB}} \
  --maxOutputQueueMB {{.baz2BamMaxOutputQueueMB}} \
`

func WriteBaz2bamBash(wr io.Writer, tc *TopConfig, obj *PostprimaryObject) error {
	t := CreateTemplate(Template_baz2bam, "")
	kv := make(map[string]string)
	UpdateWithConfig(kv, tc)
	kv["acqId"] = obj.Uuid
	kv["bazFile"] = obj.BazFileUrl                 // TODO
	kv["metadataFile"] = obj.SubreadsetMetadataXml // written into a file?
	kv["baz2bamComputingThreads"] = "16"
	kv["bamThreads"] = "16"
	kv["inlinePbi"] = "true"
	kv["maxInputQueueMB"] = "39"
	kv["zmwBatchMB"] = "40"
	kv["headerBatchMB"] = "41"
	kv["baz2BamMaxOutputQueueMB"] = "42"

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
  --config=common.chipClass={{.job_chipClass}} \
  --config=common.platform={{.job_platform}} \
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
