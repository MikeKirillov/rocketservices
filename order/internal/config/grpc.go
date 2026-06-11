package config

import (
	"net"
	"time"
)

type keepaliveGRPCConfig struct {
	Time    int64 `yaml:"timeMin" env:"GRPC_KEEPALIVE_TIME" env-default:"5"`
	Timeout int64 `yaml:"timeoutSec" env:"GRPC_KEEPALIVE_TIMEOUT" env-default:"1"`
}

type grpcConfig struct {
	InventoryGRPCConfig inventoryGRPCConfig `yaml:"inventory"`
	PaymentGRPCConfig   paymentGRPCConfig   `yaml:"payment"`
	AuthGRPCConfig      authGRPCConfig      `yaml:"auth"`
}

type inventoryGRPCConfig struct {
	Host      string              `yaml:"host" env:"GRPC_INVENTORY_HOST" env-default:"localhost"`
	Port      string              `yaml:"port" env:"GRPC_INVENTORY_PORT" env-default:"50051"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *inventoryGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *inventoryGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *inventoryGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}

type paymentGRPCConfig struct {
	Host      string              `yaml:"host" env:"GRPC_PAYMENT_HOST" env-default:"localhost"`
	Port      string              `yaml:"port" env:"GRPC_PAYMENT_PORT" env-default:"50052"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *paymentGRPCConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *paymentGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *paymentGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}

type authGRPCConfig struct {
	Host      string              `yaml:"host" env:"GRPC_PAYMENT_HOST" env-default:"localhost"`
	Port      string              `yaml:"port" env:"GRPC_PAYMENT_PORT" env-default:"50052"`
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
