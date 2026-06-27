package config

type otelConfig struct {
	Endpoint       string  `yaml:"collector_endpoint"`
	ServiceName    string  `yaml:"service_name"`
	Environment    string  `yaml:"environment"`
	ServiceVersion string  `yaml:"service_version"`
	SamplingRatio  float64 `yaml:"sampling_ratio"`
	EnableOTLP     bool    `yaml:"enable_otlp"`
}
