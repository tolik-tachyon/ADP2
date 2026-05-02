package consumer

import "github.com/nats-io/nats.go"

func SetupStream(js nats.JetStreamContext) error {
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     "PAYMENTS",
		Subjects: []string{"payment.completed"},
		Storage:  nats.FileStorage,
	})

	if err != nil {
		return err
	}

	_, err = js.AddConsumer("PAYMENTS", &nats.ConsumerConfig{
		Durable:        "NOTIFICATION_CONSUMER",
		AckPolicy:      nats.AckExplicitPolicy,
		AckWait:        10 * 1e9, // 10s
		MaxDeliver:     3,
		DeliverSubject: "payment.completed",
	})

	return err
}
