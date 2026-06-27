package config

import (
	"time"
)

type keepaliveGRPCConfig struct {
	Time    int64 `yaml:"timeMin" env:"GRPC_KEEPALIVE_TIME" env-default:"5"`
	Timeout int64 `yaml:"timeoutSec" env:"GRPC_KEEPALIVE_TIMEOUT" env-default:"1"`
}

type grpcConfig struct {
	InventoryGRPCConfig inventoryGRPCConfig `yaml:"inventory_client"`
	PaymentGRPCConfig   paymentGRPCConfig   `yaml:"payment_client"`
	AuthGRPCConfig      authGRPCConfig      `yaml:"iam_client"`
}

type inventoryGRPCConfig struct {
	Address   string              `yaml:"address" env:"GRPC_INVENTORY_ADDRESS" env-default:"localhost:50051"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *inventoryGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *inventoryGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}

type paymentGRPCConfig struct {
	Address   string              `yaml:"address" env:"GRPC_PAYMENT_ADDRESS" env-default:"localhost:50052"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *paymentGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *paymentGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}

type authGRPCConfig struct {
	Address   string              `yaml:"address" env:"GRPC_AUTH_ADDRESS" env-default:"localhost:50053"`
	Keepalive keepaliveGRPCConfig `yaml:"keepalive"`
}

func (c *authGRPCConfig) Time() time.Duration {
	return time.Duration(c.Keepalive.Time) * time.Minute
}

func (c *authGRPCConfig) Timeout() time.Duration {
	return time.Duration(c.Keepalive.Timeout) * time.Second
}
