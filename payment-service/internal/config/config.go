package config

import (
	"log"
	"os"
)

type Config struct {
	DB_DSN  string
	NATSURL string
	Port    string
}

func Load() *Config {
	cfg := &Config{
		DB_DSN:  os.Getenv("PAYMENT_DB_DSN"),
		NATSURL: os.Getenv("NATS_URL"),
		Port:    "50051",
	}

	if cfg.DB_DSN == "" {
		log.Fatal("PAYMENT_DB_DSN is required")
	}
	if cfg.NATSURL == "" {
		log.Fatal("NATS_URL is required")
	}

	return cfg
}
