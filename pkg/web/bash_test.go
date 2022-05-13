package web

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"pacb.com/seq/paws/pkg/config"
	"testing"
)

func hexComp(expected, got string) string {
	return fmt.Sprintf("\nGot\n%s Expected\n%s", hex.Dump([]byte(got)), hex.Dump([]byte(expected)))
}

func TestTranslateDiscardableUrl(t *testing.T) {
	{
		got := TranslateDiscardableUrl(nil, "--foo", "discard:")
		expected := ""
		assert.Equal(t, expected, got)
	}
	{
		got := TranslateDiscardableUrl(nil, "--foo", "/bar")
		expected := "--foo /bar"
		assert.Equal(t, expected, got)
	}
}
func TestWriteReduceStatsBash(t *testing.T) {
	if false {
		expected := `
echo 'PA_PPA_STATUS {"counter":0,"counterMax":1,"stageName":"Bash","stageNumber":0,"stageWeights":[100],"state":"progress","timeoutForNextStatus":300}' >&2

ppa-reducestats \
  --input PREFIX.sts.h5 \
  --output PREFIX.rsts.h5 \
`
		mid := "m123"
		obj := &PostprimaryObject{
			OutputPrefixUrl: "PREFIX",
			//Mid: mid, // does not matter in this case
		}
		var b bytes.Buffer
		tc := config.Top()
		so := GetLocalStorageObject("", "", "", mid)
		err := WriteReduceStatsBash(&b, tc, obj, so)
		assert.Nil(t, err)
		got := b.String()
		assert.Equal(t, expected, got)
	}
	{
		expected := `
echo 'PA_PPA_STATUS {"counter":0,"counterMax":1,"stageName":"Bash","stageNumber":0,"stageWeights":[100],"state":"progress","timeoutForNextStatus":300}' >&2

ppa-reducestats \
  --input nrt/0/m123/m123.sts.h5 \
  --output nrt/0/m123/m123.rsts.h5 \
`
		mid := "m123"
		obj := &PostprimaryObject{
			//OutputPrefixUrl: "PREFIX",
			Mid: mid,
		}
		var b bytes.Buffer
		tc := config.Top()
		partition := "0"
		so := GetLocalStorageObject("nrt", "icc", partition, mid)
		obj.OutputPrefixUrl = ChooseUrlThenRegister(so, obj.OutputPrefixUrl, StoragePathNrt, mid)
		//obj.LogUrl = ChooseUrlThenRegister(so, obj.LogUrl, StoragePathNrt, mid+".baz2bam.log")
		//log.Printf("OutputPrefixUrl: %q", obj.OutputPrefixUrl)

		//obj.LogReducdStatsUrl = ChooseUrlThenRegister(so, obj.LogReduceStatsUrl, StoragePathNrt, mid+".reducestats.log")
		//logoutput := TranslateUrl(so, obj.LogReduceStatsUrl)

		OutputStatsH5Url := obj.OutputPrefixUrl + ".sts.h5"
		obj.OutputStatsH5Url = ChooseUrlThenRegister(so, OutputStatsH5Url, StoragePathNrt, mid+".sts.h5")
		OutputReduceStatsH5Url := obj.OutputPrefixUrl + ".rsts.h5"
		obj.OutputReduceStatsH5Url = ChooseUrlThenRegister(so, OutputReduceStatsH5Url, StoragePathNrt, mid+".rsts.h5")
		//log.Printf("OutputReduceStatsH5Url: %q", OutputReduceStatsH5Url)

		err := WriteReduceStatsBash(&b, tc, obj, so)
		assert.Nil(t, err)
		got := b.String()
		assert.Equal(t, expected, got)
	}
}

const (
	basecallerObjJson = `
{"uuid":"eea6a94e-8e8b-4203-9844-8540d49662a8",
"bazUrl":"/data/nrta/3/m123/m123.baz",
"traceFileUrl":"file:/data/nrta/3/m123/m123.trc.h5",
"rtMetricsUrl": "file:/data/nrta/3/m123/m123.rtmetrics.json",
"darkcalFileUrl":"file:/data/nrta/3/m123/m123.darkcal_220325_032954.h5",
"chiplayout":"Spider_1p0_NTO",
"pixelSpreadFunction":[[0.0009999999,0.00390000013,0.00735,0.0044,0.00195000006],[0.00199999986,0.0201000012,0.05845,0.02015,0.0049],[0.0055,0.0528,0.634799957,0.04755,0.0088],[0.00445,0.021949999,0.0569000021,0.02155,0.0035],
[0.00195000006,0.0044,0.0059,0.00390000013,0.00195000006]],
"crosstalkFilter":null,
"analogs":[{"baseLabel":"A","relativeAmp":0.68000000715255737,"interPulseDistanceSec":0.11999999731779099,"excessNoiseCv":0.23999999463558197,
            "pulseWidthSec":0.16599999368190765,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},
           {"baseLabel":"C","relativeAmp":1.0,"interPulseDistanceSec":0.10999999940395355,"excessNoiseCv":0.36000001430511475,
           "pulseWidthSec":0.20900000631809235,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},
           {"baseLabel":"G","relativeAmp":0.27000001072883606,"interPulseDistanceSec":0.10999999940395355,"excessNoiseCv":0.27000001072883606,
           "pulseWidthSec":0.19300000369548798,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},
           {"baseLabel":"T","relativeAmp":0.43000000715255737,"interPulseDistanceSec":0.11999999731779099,"excessNoiseCv":0.34000000357627869,
           "pulseWidthSec":0.16300000250339508,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0}],
           "sequencingRoi":[[32,64,1080,1920]
           ],
    "traceFileRoi":[[135,288,1,32],[135,768,1,32],[135,1248,1,32],[135,1728,1,32],[405,288,1,32],[405,768,1,32],[405,1248,1,32],
                    [405,1728,1,32],[675,288,1,32],[675,768,1,32],[675,1248,1,32],[675,1728,1,32],[945,288,1,32],[945,768,1,32],
                    [945,1248,1,32],[945,1728,1,32]],
    "expectedFrameRate":100,
    "photoelectronSensitivity":6.6666665077209473,
    "refSnr":15.300000190734863,
    "simulationFileUrl":"",
    "smrtBasecallerConfig":null,
    "movieMaxFrames":60000,
    "movieMaxSeconds":660.0,
    "movieNumber":1,
    "mid":"m123",
    "logUrl":"/data/nrta/3/m123/m123.basecaller.log",
    "logLevel":"INFO"}
`
)

func TestWriteBasecallerBash(t *testing.T) {
	// Clean up for previous Bamboo runs.
	defer func() {
		_ = os.Remove("/tmp/pawsgo/TestWriteBasecallerBash/m123/m123.basecaller.config.json")
	}()
	obj := &SocketBasecallerObject{}
	err := json.Unmarshal([]byte(basecallerObjJson), &obj)
	assert.Nil(t, err)

	{
		// NUMA_NODE and GPU_ID are currently implemented as (sraIndex % 2) which may change in the future.
		expected := `
export NUMA_NODE=1
export GPU_ID=1
smrt-basecaller-launch.sh \
  --config multipleBazFiles=false \
  --statusfd 2 \
  --logoutput /data/nrta/3/m123/m123.basecaller.log \
  --logfilter INFO \
  --outputtrcfile /data/nrta/3/m123/m123.trc.h5 \
  --config traceSaver.roi='[[135,288,1,32],[135,768,1,32],[135,1248,1,32],[135,1728,1,32],[405,288,1,32],[405,768,1,32],[405,1248,1,32],[405,1728,1,32],[675,288,1,32],[675,768,1,32],[675,1248,1,32],[675,1728,1,32],[945,288,1,32],[945,768,1,32],[945,1248,1,32],[945,1728,1,32]]' \
  --outputbazfile /data/nrta/3/m123/m123.baz \
  --config realTimeMetrics.jsonOutputFile=/data/nrta/3/m123/m123.rtmetrics.json \
  --config algorithm.modelEstimationMode=FixedEstimations \
  --config /tmp/pawsgo/TestWriteBasecallerBash/m123/m123.basecaller.config.json \
  --config source.WXIPCDataSourceConfig.sraIndex=3 \
  --config dataSource.darkCalFileName=/data/nrta/3/m123/m123.darkcal_220325_032954.h5 \
  --config dataSource.imagePsfKernel=[[0.0009999999,0.00390000013,0.00735,0.0044,0.00195000006],[0.00199999986,0.0201000012,0.05845,0.02015,0.0049],[0.0055,0.0528,0.634799957,0.04755,0.0088],[0.00445,0.021949999,0.0569000021,0.02155,0.0035],[0.00195000006,0.0044,0.0059,0.00390000013,0.00195000006]] \
   \
  --config system.analyzerHardware=A100 \
  --maxFrames 60000 \
`
		var b bytes.Buffer
		tc := config.Top()
		DataDir = "/tmp/pawsgo/TestWriteBasecallerBash" // Note: global side-effect
		mid := "m123"
		so := GetLocalStorageObject("/data/nrta", "/data/icc", "3", mid)
		err = WriteBasecallerBash(&b, tc, obj, "4", so)
		assert.Nil(t, err)
		got := b.String()
		assert.Equal(t, expected, got)
	}
	{
		dat, err := os.ReadFile("/tmp/pawsgo/TestWriteBasecallerBash/m123/m123.basecaller.config.json")
		assert.Nil(t, err)
		// fmt.Print(got)
		got := string(dat)

		expected := `{
    "source": {
        "WXIPCDataSourceConfig": {
            "acqConfig": {
                "A": {
                    "baseLabel": "A",
                    "excessNoiseCV": 0.23999999463558197,
                    "interPulseDistance": 0.11999999731779099,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.16599999368190765,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.6800000071525574
                },
                "C": {
                    "baseLabel": "C",
                    "excessNoiseCV": 0.36000001430511475,
                    "interPulseDistance": 0.10999999940395355,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.20900000631809235,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 1
                },
                "G": {
                    "baseLabel": "G",
                    "excessNoiseCV": 0.27000001072883606,
                    "interPulseDistance": 0.10999999940395355,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.19300000369548798,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.27000001072883606
                },
                "T": {
                    "baseLabel": "T",
                    "excessNoiseCV": 0.3400000035762787,
                    "interPulseDistance": 0.11999999731779099,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.16300000250339508,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.4300000071525574
                },
                "chipLayoutName": "KestrelRTO2",
                "refSnr": 15.300000190734863,
                "photoelectronSensitivity": 6.666666507720947
            }
        }
    }
}`
		assert.Equal(t, expected, got, "basecaller.config.json[1]")
	}

	obj.TraceFileRoi = obj.TraceFileRoi[:0] // or nil; both have len()==0
	obj.PhotoelectronSensitivity = 6.0
	obj.PixelSpreadFunction = nil
	obj.CrosstalkFilter = [][]float64{{0.0, 0.0, 0.0}, {0.0, 0.0, 1.0}, {0.0, 0.0, 0.0}}
	{
		expected := `
export NUMA_NODE=0
export GPU_ID=0
smrt-basecaller-launch.sh \
  --config multipleBazFiles=false \
  --statusfd 2 \
  --logoutput /data/nrta/3/m123/m123.basecaller.log \
  --logfilter INFO \
   \
   \
  --outputbazfile /data/nrta/3/m123/m123.baz \
  --config realTimeMetrics.jsonOutputFile=/data/nrta/3/m123/m123.rtmetrics.json \
  --config algorithm.modelEstimationMode=FixedEstimations \
  --config /tmp/pawsgo/TestWriteBasecallerBash/m123/m123.basecaller.config.json \
  --config source.WXIPCDataSourceConfig.sraIndex=2 \
  --config dataSource.darkCalFileName=/data/nrta/3/m123/m123.darkcal_220325_032954.h5 \
   \
  --config dataSource.crosstalkFilterKernel=[[0,0,0],[0,0,1],[0,0,0]] \
  --config system.analyzerHardware=A100 \
  --maxFrames 60000 \
`
		var b bytes.Buffer
		tc := config.Top()
		mid := "m123"
		so := GetLocalStorageObject("/data/nrta", "/data/icc", "2", mid)
		err = WriteBasecallerBash(&b, tc, obj, "3", so)
		assert.Nil(t, err)
		got := b.String()
		assert.Equal(t, expected, got)
	}
	{
		dat, err := os.ReadFile("/tmp/pawsgo/TestWriteBasecallerBash/m123/m123.basecaller.config.json")
		assert.Nil(t, err)
		// fmt.Print(got)
		got := string(dat)
		expected := `{
    "source": {
        "WXIPCDataSourceConfig": {
            "acqConfig": {
                "A": {
                    "baseLabel": "A",
                    "excessNoiseCV": 0.23999999463558197,
                    "interPulseDistance": 0.11999999731779099,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.16599999368190765,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.6800000071525574
                },
                "C": {
                    "baseLabel": "C",
                    "excessNoiseCV": 0.36000001430511475,
                    "interPulseDistance": 0.10999999940395355,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.20900000631809235,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 1
                },
                "G": {
                    "baseLabel": "G",
                    "excessNoiseCV": 0.27000001072883606,
                    "interPulseDistance": 0.10999999940395355,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.19300000369548798,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.27000001072883606
                },
                "T": {
                    "baseLabel": "T",
                    "excessNoiseCV": 0.3400000035762787,
                    "interPulseDistance": 0.11999999731779099,
                    "ipd2SlowStepRatio": 0,
                    "pulseWidth": 0.16300000250339508,
                    "pw2SlowStepRatio": 3.200000047683716,
                    "relAmplitude": 0.4300000071525574
                },
                "chipLayoutName": "KestrelRTO2",
                "refSnr": 15.300000190734863,
                "photoelectronSensitivity": 6
            }
        }
    }
}`
		assert.Equal(t, expected, got, "basecaller.config.json[2]")
	}
}
func TestGetPostprimaryHostname(t *testing.T) {
	{
		got := GetPostprimaryHostname("snafu", "/data/nrta/5")
		expected := ""
		assert.Equal(t, expected, got)
	}
	{
		got := GetPostprimaryHostname("rt-84006.fubar.com", "/data/nrta/5")
		expected := "nrta"
		assert.Equal(t, expected, got)
	}
}
func TestUniqueLabel(t *testing.T) {
	try := func(so *StorageObject, expected string) {
		got := UniqueLabel(so)
		assert.Equal(t, expected, got)
	}
	so := &StorageObject{}
	try(so, ".00")
	try(so, ".01")
	try(so, ".02")
}
