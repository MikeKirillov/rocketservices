package config

import "net"

// Go-структуры конфига содержат оба набора тегов — yaml для маппинга из YAML
// и env/env-default для маппинга из переменных окружения.
type grpcConfig struct {
	Host string `yaml:"host" env:"GRPC_HOST" env-default:"localhost"`
	Port string `yaml:"port" env:"GRPC_PORT" env-default:"50052"`
}

func (c *grpcConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
