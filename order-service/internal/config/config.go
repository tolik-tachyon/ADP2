package config

import (
	"log"
	"os"
)

type Config struct {
	DB_DSN      string
	RedisAddr   string
	PaymentGRPC string
	Port        string
}

func Load() *Config {
	cfg := &Config{
		DB_DSN:      os.Getenv("ORDER_DB_DSN"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		PaymentGRPC: os.Getenv("PAYMENT_GRPC_URL"),
		Port:        os.Getenv("ORDER_PORT"),
	}

	if cfg.DB_DSN == "" {
		log.Fatal("ORDER_DB_DSN is required")
	}
	if cfg.RedisAddr == "" {
		log.Fatal("REDIS_ADDR is required")
	}
	if cfg.PaymentGRPC == "" {
		cfg.PaymentGRPC = "payment-service:50051"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}
