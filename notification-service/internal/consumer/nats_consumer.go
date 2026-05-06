package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

type Event struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type Consumer struct {
	js   nats.JetStreamContext
	seen map[string]bool
	mu   sync.Mutex
}

func NewConsumer(js nats.JetStreamContext) *Consumer {
	return &Consumer{
		js:   js,
		seen: make(map[string]bool),
	}
}

func (c *Consumer) sendToDLQ(evt Event, reason string) {
	data, _ := json.Marshal(map[string]any{
		"event":  evt,
		"reason": reason,
	})

	_, err := c.js.Publish("payment.dlq", data)
	if err != nil {
		log.Println("failed to send to DLQ:", err)
	}
}

func (c *Consumer) Listen(ctx context.Context) {
	sub, err := c.js.Subscribe(
		"payment.completed",
		func(msg *nats.Msg) {

			var evt Event
			if err := json.Unmarshal(msg.Data, &evt); err != nil {
				log.Println("invalid message:", err)
				_ = msg.Ack()
				return
			}

			c.mu.Lock()
			if c.seen[evt.EventID] {
				c.mu.Unlock()
				_ = msg.Ack()
				return
			}
			c.seen[evt.EventID] = true
			c.mu.Unlock()

			if evt.OrderID == "fail-test" {
				log.Println("[Notification] sending to DLQ")

				c.sendToDLQ(evt, "forced failure test")
				_ = msg.Ack()
				return
			}

			if evt.Amount < 0 {
				log.Println("[Notification] invalid amount → DLQ")

				c.sendToDLQ(evt, "invalid amount")
				_ = msg.Ack()
				return
			}

			log.Printf(
				"[Notification] Sent email to %s for Order #%s. Amount: %d\n",
				evt.CustomerEmail,
				evt.OrderID,
				evt.Amount,
			)

			_ = msg.Ack()
		},
		nats.Durable("NOTIFICATION_CONSUMER"),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.AckWait(10e9),
		nats.MaxDeliver(3),
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Notification consumer started")

	<-ctx.Done()

	_ = sub.Unsubscribe()
	log.Println("Notification consumer stopped")
}
