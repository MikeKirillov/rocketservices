package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const defaultConfigPath = "config.local.yaml"

type config struct {
	GRPC   grpcConfig   `yaml:"grpc"`
	Logger loggerConfig `yaml:"logger"`
}

var appConfig *config

// ResolveConfigPath определяет путь к конфиг-файлу по цепочке приоритетов:
// флаг -config > env CONFIG_PATH > "config.local.yaml".
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

	return defaultConfigPath
}

func Load(path string) (*config, error) {
	var cfg config

	// ReadConfig читает YAML-файл, а затем перетирает значения из env-переменных
	// Приоритет: env > yaml > env-default
	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return nil, fmt.Errorf("не удалось загрузить конфиг из %q: %w", path, err)
		}
		appConfig = &cfg
		return appConfig, nil
	}

	// Если путь не указан — читаем только из env-переменных
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфиг из env: %w", err)
	}

	appConfig = &cfg
	return appConfig, nil
}

func AppConfig() *config {
	return appConfig
}
