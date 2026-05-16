package config

import (
	"net"
	"time"
)

type httpConfig struct {
	Host    string        `yaml:"host" env:"HTTP_HOST" env-default:"localhost"`
	Port    string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	Timeout timeoutConfig `yaml:"timeout"`
}

func (c *httpConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c *httpConfig) TimeoutReadHeader() time.Duration {
	return time.Duration(c.Timeout.ReadHeader) * time.Second
}

func (c *httpConfig) TimeoutRead() time.Duration {
	return time.Duration(c.Timeout.Read) * time.Second
}

func (c *httpConfig) TimeoutWrite() time.Duration {
	return time.Duration(c.Timeout.Write) * time.Second
}

func (c *httpConfig) TimeoutIdle() time.Duration {
	return time.Duration(c.Timeout.Idle) * time.Second
}

func (c *httpConfig) TimeoutShutdown() time.Duration {
	return time.Duration(c.Timeout.Shutdown) * time.Second
}

func (c *httpConfig) TimeoutMiddlware() time.Duration {
	return time.Duration(c.Timeout.Middlware) * time.Second
}

type timeoutConfig struct {
	ReadHeader int64 `yaml:"readHeaderSec" env:"HTTP_READ_HEADER" env-default:"5000"`
	Read       int64 `yaml:"readSec" env:"HTTP_READ" env-default:"15000"`
	Write      int64 `yaml:"writeSec" env:"HTTP_WRITE" env-default:"15000"`
	Idle       int64 `yaml:"idleSec" env:"HTTP_IDLE" env-default:"60000"`
	Shutdown   int64 `yaml:"shutdownSec" env:"HTTP_SHUTDOWN" env-default:"10000"`
	Middlware  int64 `yaml:"middlwareSec" env:"HTTP_MIDDLWARE" env-default:"10000"`
}
