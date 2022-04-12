package web

import (
	"bytes"
	"encoding/json"
	"pacb.com/seq/paws/pkg/config"
	"testing"
)

func TestTranslateDiscardableUrl(t *testing.T) {
	{
		got := TranslateDiscardableUrl("--foo", "discard:")
		expected := ""
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}
	{
		got := TranslateDiscardableUrl("--foo", "/bar")
		expected := "--foo /bar"
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}
	{
		got := TranslateUrl("file:/bar")
		expected := "/bar"
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}
	{
		got := TranslateUrl("local")
		expected := "local"
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}
}
func TestWriteReduceStatsBash(t *testing.T) {
	expected := `
ppa-reducestats \
  --input PREFIX.sts.h5 \
  --output PREFIX.rsts.h5 \
`
	obj := &PostprimaryObject{
		OutputPrefixUrl: "PREFIX",
	}
	var b bytes.Buffer
	tc := config.Top()
	err := WriteReduceStatsBash(&b, tc, obj)
	check(err)
	got := b.String()
	if got != expected {
		t.Errorf("Got %s", got)
	}
}

const (
	basecallerObjJson = `
{"uuid":"eea6a94e-8e8b-4203-9844-8540d49662a8","bazUrl":"/data/nrta/0/m84003_220325_032134_s1.baz","traceFileUrl":"file:/data/nrta/0/m84003_220325_032134_s1.trc.h5","darkcalFileUrl":"file:/data/nrta/0/m84003_220325_032134_s1.darkcal_220325_032954.h5","chiplayout":"Spider_1p0_NTO","pixelSpreadFunction":[[0.0009999999,0.00390000013,0.00735,0.0044,0.00195000006],[0.00199999986,0.0201000012,0.05845,0.02015,0.0049],[0.0055,0.0528,0.634799957,0.04755,0.0088],[0.00445,0.021949999,0.0569000021,0.02155,0.0035],[0.00195000006,0.0044,0.0059,0.00390000013,0.00195000006]],"crosstalkFilter":null,"analogs":[{"baseLabel":"A","relativeAmp":0.68000000715255737,"interPulseDistanceSec":0.11999999731779099,"excessNoiseCv":0.23999999463558197,"pulseWidthSec":0.16599999368190765,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},{"baseLabel":"C","relativeAmp":1.0,"interPulseDistanceSec":0.10999999940395355,"excessNoiseCv":0.36000001430511475,"pulseWidthSec":0.20900000631809235,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},{"baseLabel":"G","relativeAmp":0.27000001072883606,"interPulseDistanceSec":0.10999999940395355,"excessNoiseCv":0.27000001072883606,"pulseWidthSec":0.19300000369548798,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0},{"baseLabel":"T","relativeAmp":0.43000000715255737,"interPulseDistanceSec":0.11999999731779099,"excessNoiseCv":0.34000000357627869,"pulseWidthSec":0.16300000250339508,"pw2SlowStepRatio":3.2000000476837158,"ipd2SlowStepRatio":0.0}],"sequencingRoi":[[32,64,1080,1920]],"traceFileRoi":[[135,288,1,32],[135,768,1,32],[135,1248,1,32],[135,1728,1,32],[405,288,1,32],[405,768,1,32],[405,1248,1,32],[405,1728,1,32],[675,288,1,32],[675,768,1,32],[675,1248,1,32],[675,1728,1,32],[945,288,1,32],[945,768,1,32],[945,1248,1,32],[945,1728,1,32]],"expectedFrameRate":100,"photoelectronSensitivity":6.6666665077209473,"refSnr":15.300000190734863,"simulationFileUrl":"","smrtBasecallerConfig":null,"movieMaxFrames":60000,"movieMaxSeconds":660.0,"movieNumber":1,"mid":"m84003_220325_032134_s1","logUrl":"/data/nrta/0/m84003_220325_032134_s1.basecaller.log","logLevel":"INFO"}
`
)

func TestWriteBasecallerBash(t *testing.T) {
	obj := &SocketBasecallerObject{}
	err := json.Unmarshal([]byte(basecallerObjJson), &obj)
	check(err)

	{
		expected := `
smrt-basecaller-launch.sh \
  --config multipleBazFiles=false \
  --statusfd 3 \
  --logoutput /data/nrta/0/m84003_220325_032134_s1.basecaller.log \
  --logfilter INFO \
  --outputtrcfile /data/nrta/0/m84003_220325_032134_s1.trc.h5 \
  --config traceSaver.roi='[[135,288,1,32],[135,768,1,32],[135,1248,1,32],[135,1728,1,32],[405,288,1,32],[405,768,1,32],[405,1248,1,32],[405,1728,1,32],[675,288,1,32],[675,768,1,32],[675,1248,1,32],[675,1728,1,32],[945,288,1,32],[945,768,1,32],[945,1248,1,32],[945,1728,1,32]]' \
  --outputbazfile /data/nrta/0/m84003_220325_032134_s1.baz \
  --config /tmp/3/m84003_220325_032134_s1.basecaller.config.json \
  --config source.WXIPCDataSourceConfig.sraIndex=3 \
   \
   \
  --config system.analyzerHardware=A100 \
  --maxFrames 60000 \
`
		var b bytes.Buffer
		tc := config.Top()
		err = WriteBasecallerBash(&b, tc, obj, "4")
		check(err)
		got := b.String()
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}

	obj.TraceFileRoi = obj.TraceFileRoi[:0] // or nil; both have len()==0
	{
		expected := `
smrt-basecaller-launch.sh \
  --config multipleBazFiles=false \
  --statusfd 3 \
  --logoutput /data/nrta/0/m84003_220325_032134_s1.basecaller.log \
  --logfilter INFO \
   \
   \
  --outputbazfile /data/nrta/0/m84003_220325_032134_s1.baz \
  --config /tmp/3/m84003_220325_032134_s1.basecaller.config.json \
  --config source.WXIPCDataSourceConfig.sraIndex=3 \
   \
   \
  --config system.analyzerHardware=A100 \
  --maxFrames 60000 \
`
		var b bytes.Buffer
		tc := config.Top()
		err = WriteBasecallerBash(&b, tc, obj, "4")
		check(err)
		got := b.String()
		if got != expected {
			t.Errorf("Got %s", got)
		}
	}
}
