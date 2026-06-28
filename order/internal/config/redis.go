package config

type redisConfig struct {
	Address string `yaml:"redis_address" env:"REDIS_ADDRESS" env-default:"localhost:6379"`
	Rate    int    `yaml:"rate" env:"REDIS_RATE" env-default:"100"`
	Burst   int    `yaml:"burst" env:"REDIS_BURST" env-default:"200"`
}
