package messaging

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	js nats.JetStreamContext
}

func NewPublisher(nc *nats.Conn) (*Publisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return &Publisher{js: js}, nil
}

type PaymentEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

func (p *Publisher) PublishPaymentCompleted(evt PaymentEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, err = p.js.Publish("payment.completed", data)
	return err
}
