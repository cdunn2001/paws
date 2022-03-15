package web

import (
	"io"
	"strconv"
	"text/template"
)

var (
	Binary_baz2bam         = "baz2bam"
	Binary_smrt_basecaller = "smrt-basecaller"
	Binary_pa_cal          = "kes-cal"
	Binary_reduce_stats    = "reduce-stats"
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
`

func WriteDarkcalBash(wr io.Writer, obj SocketDarkcalObject, SocketId string) error {
	t := CreateTemplate(Template_darkcal, "")
	kv := make(map[string]string)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	kv["movieNum"] = "0" // for now
	// assert if obj.movieNum not nil, then it is 0.

	kv["numFrames"] = strconv.Itoa(int(obj.MaxMovieFrames))
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = obj.CalibFileUrl // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl        // TODO: Convert from URL!

	// Skip --timeoutseconds for now. # but ask MarkL about this; might not need it anymore
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
`

func WriteLoadingcalBash(wr io.Writer, obj SocketLoadingcalObject, SocketId string) error {
	t := CreateTemplate(Template_loadingcal, "")
	kv := make(map[string]string)

	socketIdInt, err := strconv.Atoi(SocketId)
	if err != nil {
		return err
	}
	sra := socketIdInt - 1 // for now
	kv["sra"] = strconv.Itoa(sra)

	kv["movieNum"] = "0" // for now
	// assert if obj.movieNum not nil, then it is 0.

	kv["numFrames"] = strconv.Itoa(int(obj.MaxMovieFrames))
	// --numFrames # gets overridden w/ 128 or 512 for now, but setting prevents warning

	kv["outputFile"] = obj.CalibFileUrl           // TODO: Convert from URL!
	kv["logoutput"] = obj.LogUrl                  // TODO: Convert from URL!
	kv["inputDarkcalFile"] = obj.DarkFrameFileUrl // TODO: Convert from URL!

	// Skip --timeoutseconds for now. # but ask MarkL about this; might not need it anymore

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
  --config traceSaver.roi=<roi specification> \
  --config source.WXIPCDataSource.acqConfig=<Info-About-Chemistry> \
  --config system.analyzerHardware=A100 \ # optional
  --config algorithm=<forward-from-user> \ # optional
`

// Doesn't this need the darkcalfile?
func WriteBasecallerBash(wr io.Writer, obj SocketBasecallerObject, SocketId string) error {
	t := CreateTemplate(Template_basecaller, "")
	kv := make(map[string]string)

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

func WriteBaz2bamBash(wr io.Writer, obj PostprimaryObject) error {
	t := CreateTemplate(Template_baz2bam, "")
	kv := make(map[string]string)
	kv["acqId"] = obj.Uuid
	// kv["metadataFile"] = obj.SubreadsetMetadataXml // written into a file?
	// kv["baz2bamComputingThreads"] = ?
	// kv["bamThreads"] = ?
	// {{if .inlinePbi}}--inlinePbi{{end}} \
	// kv["maxInputQueueMB"] = ?
	// kv["zmwBatchMB"] = ?
	// kv["headerBatchMB"] = ?
	// kv["baz2BamMaxOutputQueueMB"] = ?

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

var Template_reduce_stats = `
{{.Binary_reduce_stats}} \
  --input {{.job_outputPrefix}}.sts.h5 \
  --output {{.job_outputPrefix}}.rsts.h5 \
  --config=common.chipClass={{.job_chipClass}} \
  --config=common.platform={{.job_platform}} \
`

func WriteReduceStatsBash(wr io.Writer, obj PostprimaryObject, job Job) error {
	t := CreateTemplate(Template_reduce_stats, "")
	kv := make(map[string]string)
	UpdateJob(kv, job)
	kv["Binary_reduce_stats"] = Binary_reduce_stats
	//obj.OutputReduceStatsH5Url
	return t.Execute(wr, kv)
}
func CopyRsts(obj PostprimaryObject, job Job) error {
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
