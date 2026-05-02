package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"notification-service/internal/consumer"
)

func main() {
	natsURL := os.Getenv("NATS_URL")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	c := consumer.NewConsumer(nc)

	go c.Listen()

	log.Println("Notification service running...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down notification service...")
}
