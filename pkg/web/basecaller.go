package web

// This is the JSON definition of the Analogs that smrt-basecaller uses.
// It's not the same as the JSON definition that /sockets/#/basecaller/start uses. sigh.
// The prefix Smrt_ indicates that this is used by smrt-basecaller
type Smrt_AnalogObject struct {
	BaseLabel          string  `json:"baseLabel"`
	ExcessNoiseCV      float64 `json:"excessNoiseCV"`
	InterPulseDistance float64 `json:"interPulseDistance"`
	Ipd2SlowStepRatio  float64 `json:"ipd2SlowStepRatio"`
	PulseWidth         float64 `json:"pulseWidth"`
	Pw2SlowStepRatio   float64 `json:"pw2SlowStepRatio"`
	RelAmplitude       float64 `json:"relAmplitude"`
}

// The prefix Smrt_ indicates that this is used by smrt-basecaller
type Smrt_AcqConfig struct {
	A                        Smrt_AnalogObject
	C                        Smrt_AnalogObject
	G                        Smrt_AnalogObject
	T                        Smrt_AnalogObject
	ChipLayoutName           string  `json:"chipLayoutName"`
	RefSnr                   float64 `json:"refSnr"`
	PhotoelectronSensitivity float64 `json:"photoelectronSensitivity"`
}

// The prefix Smrt_ indicates that this is used by smrt-basecaller
type Smrt_WXIPCDataSourceConfig struct {
	AcqConfig Smrt_AcqConfig `json:"acqConfig"`
}

// The prefix Smrt_ indicates that this is used by smrt-basecaller
type Smrt_Source struct {
	WXIPCDataSourceConfig Smrt_WXIPCDataSourceConfig `json:"WXIPCDataSourceConfig"`
}

// This is the top level configuration object for smrt-basecaller, passed to the --config argument.
type SmrtBasecallerConfigObject struct {
	Source Smrt_Source `json:"source"`
}
