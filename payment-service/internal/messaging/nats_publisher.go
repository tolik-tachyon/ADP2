package messaging

import (
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	js nats.JetStreamContext
}

type PaymentEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func NewPublisher(nc *nats.Conn) (*Publisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	// ✅ SAFE: ensure stream exists
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "PAYMENTS",
		Subjects: []string{"payment.completed"},
		Storage:  nats.FileStorage,
	})

	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Println("stream init error:", err)
	}

	return &Publisher{js: js}, nil
}

func (p *Publisher) PublishPaymentCompleted(evt PaymentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, err = p.js.Publish("payment.completed", data)
	if err != nil {
		return err
	}

	log.Println("[NATS] event published:", evt.EventID)

	return nil
}
