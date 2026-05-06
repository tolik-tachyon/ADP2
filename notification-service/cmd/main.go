package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"notification-service/internal/consumer"
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

	if err := consumer.SetupStream(js); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/test-publish", func(w http.ResponseWriter, r *http.Request) {
		type Event struct {
			EventID       string `json:"event_id"`
			OrderID       string `json:"order_id"`
			Amount        int64  `json:"amount"`
			CustomerEmail string `json:"customer_email"`
			Status        string `json:"status"`
		}

		var evt Event
		_ = json.NewDecoder(r.Body).Decode(&evt)

		data, _ := json.Marshal(evt)

		nc, _ := nats.Connect(os.Getenv("NATS_URL"))
		js, _ := nc.JetStream()

		_, err := js.Publish("payment.completed", data)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte("sent"))
	})

	go http.ListenAndServe(":8090", nil)

	c := consumer.NewConsumer(js)

	ctx, cancel := context.WithCancel(context.Background())

	go c.Listen(ctx)

	log.Println("Notification service running...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down notification service...")

	cancel()
	nc.Close()
}
