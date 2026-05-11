package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"notification-service/internal/consumer"
	"notification-service/internal/provider"
	"notification-service/internal/worker"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	// setup streams (PAYMENTS + DLQ)
	if err := consumer.SetupStream(js); err != nil {
		log.Fatal(err)
	}

	// Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Provider (Adapter Pattern)
	sender := &provider.MockSender{}

	// Worker (Background processing)
	w := &worker.Worker{
		Js:     js,
		Cache:  redisClient,
		Sender: sender,
	}

	ctx, cancel := context.WithCancel(context.Background())

	// start worker
	go w.Listen(ctx)

	log.Println("Notification service running...")

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down notification service...")

	cancel()
	nc.Close()
}
