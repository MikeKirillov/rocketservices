package config

type otelConfig struct {
	Endpoint       string  `yaml:"endpoint"`
	ServiceName    string  `yaml:"service_name"`
	Environment    string  `yaml:"environment"`
	ServiceVersion string  `yaml:"service_version"`
	SamplingRatio  float64 `yaml:"service_ration"`
	OTLP           bool    `yaml:"otlp"`
}
