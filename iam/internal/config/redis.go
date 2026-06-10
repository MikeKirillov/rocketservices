package config

import (
	"net"
	"time"
)

type redisConfig struct {
	Host              string  `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port              string  `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	ConnectionTimeout int64   `yaml:"connection_timeout" env:"REDIS_CONN_TIMEOUT" env-default:"5"`
	Session           session `yaml:"session"`
}

type session struct {
	TTL int64 `yaml:"ttl env:"SESSION_TTL" env-default:"24"`
}

func (r *redisConfig) Address() string {
	return net.JoinHostPort(r.Host, r.Port)
}

func (r *redisConfig) SessionTTL() time.Duration {
	return time.Duration(r.Session.TTL) * time.Hour
}

func (r *redisConfig) ConnTimeout() time.Duration {
	return time.Duration(r.ConnectionTimeout) * time.Second
}
