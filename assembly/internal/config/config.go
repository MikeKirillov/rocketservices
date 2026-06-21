package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const defaultConfig = "config.local.yaml"

type config struct {
	Logger loggerConfig `yaml:"logger"`
	Kafka  kafkaConfig  `yaml:"kafka"`
	Otel   otelConfig   `yaml:"otel"`
}

var appConfig *config

func ResolveConfigPath() string {
	var cfgFlag string
	flag.StringVar(&cfgFlag, "config", "", "путь к YAML-конфигу")
	flag.Parse()

	if cfgFlag != "" {
		return cfgFlag
	}

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	return defaultConfig
}

func Load(path string) (*config, error) {
	var cfg config

	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return nil, fmt.Errorf("не удалось загрузить конфиг из %q: %w", path, err)
		}
		appConfig = &cfg
		return appConfig, nil
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфиг из env: %w", err)
	}
	appConfig = &cfg
	return appConfig, nil
}

func AppConfig() *config {
	return appConfig
}
