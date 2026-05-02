package consumer

import (
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

func NewConsumer(nc *nats.Conn) *Consumer {
	js, _ := nc.JetStream()
	return &Consumer{
		js:   js,
		seen: make(map[string]bool),
	}
}

func (c *Consumer) Listen() {
	sub, _ := c.js.Subscribe("payment.completed",
		func(msg *nats.Msg) {
			var evt Event
			_ = json.Unmarshal(msg.Data, &evt)

			// IDEMPOTENCY CHECK
			c.mu.Lock()
			if c.seen[evt.EventID] {
				c.mu.Unlock()
				_ = msg.Ack()
				return
			}
			c.seen[evt.EventID] = true
			c.mu.Unlock()

			// simulate failure for DLQ test
			if evt.OrderID == "fail-test" {
				log.Println("[Notification] forced failure")
				return // NO ACK → retry → DLQ
			}

			log.Printf("[Notification] Sent email to %s for Order #%s. Amount: %d\n",
				evt.CustomerEmail, evt.OrderID, evt.Amount)

			_ = msg.Ack()
		},
		nats.ManualAck(),
		nats.AckExplicit(),
	)

	if sub != nil {
		log.Fatal(sub)
	}
}
