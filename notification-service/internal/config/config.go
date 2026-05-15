package config

import (
	"log"
	"os"
)

type Config struct {
	NATSURL   string
	RedisAddr string
}

func Load() *Config {
	cfg := &Config{
		NATSURL:   os.Getenv("NATS_URL"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
	}

	if cfg.NATSURL == "" {
		log.Fatal("NATS_URL is required")
	}
	if cfg.RedisAddr == "" {
		log.Fatal("REDIS_ADDR is required")
	}

	return cfg
}
