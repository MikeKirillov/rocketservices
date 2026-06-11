package config

import (
	"net"
	"time"
)

type grpcConfig struct {
	Host      string              `yaml:"host" env:"GRPC_HOST" env-default:"localhost"`
	Port      string              `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *grpcConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *grpcConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *grpcConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}

func (c *grpcConfig) MaxConnectionIdle() time.Duration {
	return time.Duration(c.Keepalive.MaxConnectionIdle) * time.Minute
}

func (c *grpcConfig) MaxConnectionAge() time.Duration {
	return time.Duration(c.Keepalive.MaxConnectionAge) * time.Minute
}

func (c *grpcConfig) MaxConnectionAgeGrace() time.Duration {
	return time.Duration(c.Keepalive.MaxConnectionAgeGrace) * time.Second
}

func (c *grpcConfig) MinPingInterval() time.Duration {
	return time.Duration(c.Keepalive.MinPingInterval) * time.Minute
}

func (c *grpcConfig) ShutdownTimeout() time.Duration {
	return time.Duration(c.Keepalive.ShutdownTimeout) * time.Second
}

type keepaliveGRPCConfig struct {
	Time                  int64 `yaml:"timeMinutes" env:"GRPC_TIME" env-default:"5"`
	Timeout               int64 `yaml:"timeoutSec" env:"GRPC_TIMEOUT" env-default:"1"`
	MaxConnectionIdle     int64 `yaml:"maxConnIdleMinutes" env:"GRPC_IDLE" env-default:"15"`
	MaxConnectionAge      int64 `yaml:"maxConnAgeMinutes" env:"GRPC_AGE" env-default:"30"`
	MaxConnectionAgeGrace int64 `yaml:"maxConnAgeGraceSec" env:"GRPC_AGE_GRACE" env-default:"5"`
	MinPingInterval       int64 `yaml:"minPingIntervalMinutes" env:"GRPC_PING_INTERVAL" env-default:"5"`
	ShutdownTimeout       int64 `yaml:"shutdownTimeoutSec" env:"GRPC_SHUTDOWN_TIMEOUT" env-default:"5"`
}

type grpcClientConfig struct {
	AuthGRPCConfig authGRPCConfig `yaml:"auth"`
}

type authGRPCConfig struct {
	Host      string              `yaml:"host" env:"GRPC_PAYMENT_HOST" env-default:"localhost"`
	Port      string              `yaml:"port" env:"GRPC_PAYMENT_PORT" env-default:"50053"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *authGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *authGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *authGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}
