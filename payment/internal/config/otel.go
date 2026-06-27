package config

// internal/config/otel.go
// otelConfig — внутренняя структура пакета config (имя с маленькой буквы),
// наружу торчит только метод OTel() OTelConfig интерфейса.
type otelConfig struct {
	Endpoint       string  `yaml:"collector_endpoint"` // адрес OTEL Collector (OTLP gRPC)
	ServiceName    string  `yaml:"service_name"`       // имя сервиса в трейсах / метриках / логах
	Environment    string  `yaml:"environment"`
	ServiceVersion string  `yaml:"service_version"`
	SamplingRatio  float64 `yaml:"sampling_ratio"`
	EnableOTLP     bool    `yaml:"enable_otlp"`
}
