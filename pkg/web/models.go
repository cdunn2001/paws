package web

// Top level status of the pa-ws process
type PawsStatusObject struct {

	// Real time seconds that pa-ws has been running
	Uptime float64 `json:"uptime"`

	// Time that pa-ws has been running, formatted to be human readable as hours, minutes, seconds, etc
	UptimeMessage string `json:"uptimeMessage"`

	// Current epoch time in seconds as seen by pa-ws (UTC)
	Time float64 `json:"time"`

	// ISO8601 timestamp (with milliseconds) of time field
	Timestamp string `json:"timestamp"`

	// Version of software, including git hash of last commit
	Version string `json:"version"`
}
type LogLevelEnum string

const (
	Debug LogLevelEnum = "DEBUG"
	Info               = "INFO"
	Warn               = "WARN"
	Error              = "ERROR"
)

type BaseLabelEnum string

const (
	N BaseLabelEnum = "N"
	A               = "A"
	C               = "C"
	G               = "G"
	T               = "T"
)

type ExecutionStatusEnum string

const (
	Unknown  ExecutionStatusEnum = ""
	Ready                        = "READY"
	Running                      = "RUNNING"
	Complete                     = "COMPLETE"
)

type CompletionStatusEnum string

const (
	Incomplete CompletionStatusEnum = ""
	Success                         = "SUCCESS"
	Failed                          = "FAILED"
	Aborted                         = "ABORTED"
)

type ProcessStatusObject struct {

	// Status of the execution of the process
	ExecutionStatus ExecutionStatusEnum `json:"executionStatus"`

	// Status of the completion of the process after it exits. Only valid if the executionStatus is COMPLETE
	CompletionStatus CompletionStatusEnum `json:"completionStatus"`

	// ISO8601 timestamp (with milliseconds) of the latest status update
	// Example: 2017-01-31T01:59:49.103Z
	Timestamp string `json:"timestamp"`

	// The exit code of the process
	ExitCode int32 `json:"exitCode"`

	// The process ID of the process
	PID int `json:"PID"`
}

// For composition in others
type socketCommonObject struct {

	// Movie context ID used to create this object
	// Example: m123456_987654
	Mid string `json:"mid"`

	// Movie length in frames. The values movieMaxFrames and movieMaxSeconds should be similar, but not exactly the same, depending on whether true elapsed time or accurate frame count is desired. One value should be the desired amount and the other value should be an emergency stop amount.
	MaxMovieFrames int32 `json:"maxMovieFrames"`

	// Movie length in seconds. The values movieMaxFrames and movieMaxSeconds should be similar, but not exactly the same, depending on whether true elapsed time or accurate frame count is desired. One value should be the desired amount and the other value should be an emergency stop amount.
	MaxMovieSeconds float64 `json:"maxMovieSeconds"`

	// Arbitrary movie number to delimite the start and end
	MovieNumber int32 `json:"movieNumber"`

	// Destination URL of the log file
	LogUrl string `json:"logUrl"`

	// Log severity threshold
	LogLevel LogLevelEnum `json:"logLevel"`

	ProcessStatus ProcessStatusObject `json:"processStatus"`
}
type SocketDarkcalObject struct {

	// Destination URL of the calibration file
	// Example: http://localhost:23632/storages/m123456_987654/loadingcal.h5
	CalibFileUrl string `json:"calibFileUrl"`

	socketCommonObject
}
type SocketLoadingcalObject struct {

	// Source URL of the dark_frame calibration file
	// Example: http://localhost:23632/storages/m123456_987654/darkcal.h5
	DarkFrameFileUrl string `json:"darkFrameFileUrl"`

	// Destination URL of the calibration file
	// Example: http://localhost:23632/storages/m123456_987654/loadingcal.h5
	CalibFileUrl string `json:"calibFileUrl"`

	socketCommonObject
}
type SocketBasecallerObject struct {

	// subreadset UUID
	// Example: 123e4567-e89b-12d3-a456-426614174000
	Uuid string `json:"uuid"`

	// Destination URL for the baz file
	// Example: http://localhost:23632/storages/m123456_987654/thefile.baz
	BazUrl string `json:"bazUrl"`

	// Destination URL for the trace file (optional)
	// Example: "discard:"
	TraceFileUrl string `json:"traceFileUrl"`

	// Controlled name of the sensor chip unit cell layout
	// Example: Minesweeper1.0
	Chiplayout string `json:"chiplayout"`

	// Source URL for the dark calibration file
	// Example: http://localhost:23632/storages/m123456_987654/darkcal.h5
	DarkCalFileUrl string `json:"darkcalFileUrl"`

	// This is required and a function of the sensor NFC tag
	// Example: List [ List [ 0, 0.1, 0 ], List [ 0.1, 0.6, 0.1 ], List [ 0, 0.1, 0 ] ]
	PixelSpreadFunction [][]float64 `json:"pixelSpreadFunction"`

	// Optional kernel definition of the crosstalk deconvolution. THe pixelSpreadFunction is used to automatically calculate one if this is not specified.
	// Example: List [ List [ 0, 0.1, 0 ], List [ 0.1, 0.6, 0.1 ], List [ 0, 0.1, 0 ] ]
	CrosstalkFilter [][]float64 `json:"crosstalkFilter"`

	Analogs []AnalogObject `json:"analogs"`

	// ROI of the ZMWs that will be used for basecalling
	// 0,0,2048,1980
	SequencingRoi [][]int32 `json:"sequencingRoi"`

	// ROI of the ZMWs that will be used for trace file writing
	// 0,0,256,32
	TraceFileRoi [][]int32 `json:"traceFileRoi"`

	// The expected (not measured) canonical frame rate
	// Example: 100
	ExpectedFrameRate int32 `json:"expectedFrameRate"`

	// The inversion of photoelectron gain of the sensor pixels.
	// Example: 1.4
	PhotoelectronSensitivity float64 `json:"photoelectronSensitivity"`

	// Reference SNR
	// Example: 10.0
	RefSnr float64 `json:"refSnr"`

	// Source URL for the file to use for transmission of simulated data. Only local files are supported currently.
	// Example: file://localhost/data/pa/sample_file.trc.h5
	SimulationFileUrl string `json:"simulationFileUrl"`

	// SmrtBasecallerConfig. Passed to smrt_basecaller --config. TODO: This will be a JSON object, but is a string here as a placeholder.
	// Example: null
	SmrtBasecallerConfig string `json:"smrtBasecallerConfig"`

	RtMetrics SocketBasecallerRTMetricsObject

	socketCommonObject
}
type SocketBasecallerRTMetricsObject struct {

	// Source URL of the most recent RT Metrics file. When the file is updated, the URL will change with the embedded timestamp
	// Example: http://localhost:23632/storages/m123456_987654/rtmetrics_20210625_123456.xml
	Url string `json:"url"`
}
type AnalogObject struct {

	// The nucleotide that the analog is attached to
	// Example: C
	BaseLabel BaseLabelEnum `json:"baseLabel"`

	// The relative amplitude in terms of pulse height.
	// Example: 0.3
	RelativeAmp float64 `json:"relativeAmp"`

	// Average time in seconds between the falling edge of the previous pulse and rising edge of the next pulse
	// Example: 0.14
	InterPulseDistanceSec float64 `json:"interPulseDistanceSec"`

	// Coefficient of variation of excess noise
	// Example: 3
	ExcessNoiseCv float64 `json:"excessNoiseCv"`

	// Average time in seconds of the width of pulses of this analog
	// Example: 0.11
	PulseWidthSec float64 `json:"pulseWidthSec"`

	// Rate constant ratio for two-step distribution of pulse width
	// Example: 0.19
	Pw2SlowStepRatio float64 `json:"pw2SlowStepRatio"`

	// Rate constant ratio for two-step distribution of interPulse distance
	// Example: 0.14
	Ipd2SlowStepRatio float64 `json:"ipd2SlowStepRatio"`
}
type SocketObject struct {
	// The socket identifier, typically "1" thru "4".
	SocketId   string                 `json:"socketId"`
	Darkcal    SocketDarkcalObject    `json:"darkcal"`
	Loadingcal SocketLoadingcalObject `json:"loadingcal"`
	Basecaller SocketBasecallerObject `json:"basecaller"`
}
type PostprimaryStatusObject struct {

	// A list of all of the URLS of the files generated by postprimary for this object
	// Example: List [ "http://localhost:23632/m123456_98765/foo.bam", "http://localhost:23632/m123456_98765/foo.baz2bam.log" ]
	OutputUrls []string `json:"outputUrls"`

	// progress of job completion. Range is [0.0, 1.0]
	// Example: 0.74
	Progress float64 `json:"progress"`

	// The rate of ZMW processing performed by baz2bam
	// Example: 3.6e6
	Baz2bamZmwsPerMin float64 `json:"baz2bamZmwsPerMin"`

	// The rate of ZMW processing performed by ccs
	// Example: 0.4e6
	Ccs2bamZmwsPerMin float64 `json:"ccs2bamZmwsPerMin"`

	// The total number of ZMWs processed so far
	// Example: 25000000
	NumZmws int64 `json:"numZmws"`

	// The peak RSS memory usage in GiB used by baz2bam
	// Example: 5.6
	Baz2bamPeakRssGb float64 `json:"baz2bamPeakRssGb"`

	// The peak RSS memory usage in GiB used by ccs
	// Example: 1.1
	Ccs2bamPeakRssGb float64 `json:"ccs2bamPeakRssGb"`
}
type PostprimaryObject struct {

	// Movie context ID used to create this object
	// Example: m123456_987654
	Mid string `json:"mid"`

	// Source URL for the BAZ file
	// Example: http://localhost:23632/m123456_98765/foo.baz
	BazFileUrl string `json:"bazFileUrl"`

	// movie UUID, used for logging purposes only (might be deprecated)
	// 123e4567-e89b-12d3-a456-426614174000
	Uuid string `json:"uuid"`

	// Destination URL of the log file
	LogUrl string `json:"logUrl"`

	// Log severity threshold
	LogLevel LogLevelEnum `json:"logLevel"`

	// Destination URL for the prefix of all output files from baz2bam and/or ccs
	// Example: http://localhost:23632/storages/0/m12346
	OutputPrefixUrl string `json:"outputPrefixUrl"`

	// Destination URL for the stats.xml file
	// Example: http://localhost:23632/storages/0/m12346.stats.xml
	OutputStatsXmlUrl string `json:"outputStatsXmlUrl"`

	// Destination URL for the stats.h5 file
	// Example: http://localhost:23632/storages/0/m12346.sts.h5
	OutputStatsH5Url string `json:"outputStatsH5Url"`

	// Destination URL for the reduced stats.h5 file
	// Example: http://localhost:23632/storages/0/m12346.rsts.h5
	OutputReduceStatsH5Url string `json:"outputReduceStatsH5Url"`

	// Controlled name of the sensor chip unit cell layout
	// Example: Minesweeper1.0
	Chiplayout string `json:"chiplayout"`

	// The subreadset metadata, derived from the original run metadata
	// Example: <SubreadSets><SubreadSet xmln= [snip] </SubreadSets>
	SubreadsetMetadataXml string `json:"subreadsetMetadataXml"`

	// Include kinetics in the run if true
	// Example: true
	IncludeKinetics bool `json:"includeKinetics"`

	// Run CCS on instrument if true
	// Example: false
	CcsOnInstrument bool `json:"ccsOnInstrument"`

	PostprimaryStatus PostprimaryStatusObject `json:"status"`

	ProcessStatus ProcessStatusObject `json:"processStatus"`
}
type StorageItemObject struct {

	// URL of this object
	// Example: http://localhost:23632/storages/m123456_987654/foobar1.bam
	Url string `json:"url"`

	// ISO8601 timestamp (with milliseconds) of file write time
	// Example: 2017-01-31T01:59:49.103998Z
	Timestamp string `json:"timestamp"`

	// size of the file
	// Example: 6593845929837
	Size int64 `json:"size"`

	// The category for this particular item in the StorageObject
	// Example: BAM
	//   [ UNKNOWN, BAM, BAZ, CAL ] TODO
	Category string `json:"category"`

	// information about the source of this file
	// Example: null
	SourceInfo string `json:"sourceInfo"`
}
type StorageDiskReportObject struct {

	// Total space allocated in bytes for this StorageObject, include used and unused space
	// Example: 6593845929837
	TotalSpace int64 `json:"totalSpace"`

	// Total unused space in bytes of this StorageObject
	// Example: 6134262344238
	FreeSpace int64 `json:"freeSpace"`
}
type StorageObject struct {

	// Movie context ID used to create this object
	// Example: m123456_987654
	Mid string `json:"mid"`

	// symbolic link to storage directory which points back to this StorageObject
	// Example: http://localhost:23632/storages/m123456_987654
	RootUrl string `json:"rootUrl"`

	// physical path to storage directory (should only be used for debugging and logging)
	// Example: file:/data/pa/m123456_987654
	LinuxPath string `json:"linuxPath"`

	// Destination URL for the log file. Logging happens during construction and freeing.
	// Example: http://localhost:23632/storages/m123456_987654/storage.log
	LogUrl string `json:"logUrl"`

	// Log severity threshold
	// Example: "INFO"
	LogLevel LogLevelEnum `json:"logLevel"`

	Files         []StorageItemObject       `json:"files"`
	Space         []StorageDiskReportObject `json:"space"`
	ProcessStatus ProcessStatusObject       `json:"processStatus"`
}

func CreateSocketBasecallerObject() (result *SocketBasecallerObject) {
	result = new(SocketBasecallerObject)
	result.ProcessStatus.ExecutionStatus = Ready
	return result
}
func CreateSocketDarkcalObject() (result *SocketDarkcalObject) {
	result = new(SocketDarkcalObject)
	result.ProcessStatus.ExecutionStatus = Ready
	return result
}
func CreateSocketLoadingcalObject() (result *SocketLoadingcalObject) {
	result = new(SocketLoadingcalObject)
	result.ProcessStatus.ExecutionStatus = Ready
	return result
}
