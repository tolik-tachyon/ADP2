package consumer

import "github.com/nats-io/nats.go"

func SetupStream(js nats.JetStreamContext) error {
	// основной stream
	_, err := js.AddStream(&nats.StreamConfig{
		Name:      "PAYMENTS",
		Subjects:  []string{"payment.completed"},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return err
	}

	// DLQ stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "PAYMENTS_DLQ",
		Subjects:  []string{"payment.dlq"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return err
	}

	return nil
}
