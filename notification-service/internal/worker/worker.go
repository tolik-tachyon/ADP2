package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"notification-service/internal/domain"
	"notification-service/internal/provider"
)

type Worker struct {
	Js     nats.JetStreamContext
	Cache  *redis.Client
	Sender provider.EmailSender
}

func (w *Worker) Listen(ctx context.Context) {
	sub, err := w.Js.Subscribe("payment.completed", func(msg *nats.Msg) {

		var evt domain.Event
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			_ = msg.Ack()
			return
		}

		err := w.process(evt)
		if err != nil {
			// no ack → NATS will retry
			return
		}

		_ = msg.Ack()
	},
		nats.ManualAck(),
	)

	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	_ = sub.Unsubscribe()
}

func (w *Worker) process(evt domain.Event) error {
	ctx := context.Background()
	key := "notif:" + evt.EventID

	// atomic idempotency (prevents race condition)
	ok, _ := w.Cache.SetNX(ctx, key, "processing", 24*time.Hour).Result()
	if !ok {
		return nil
	}

	var err error
	backoff := 2 * time.Second

	for i := 0; i < 3; i++ {
		err = w.Sender.Send(evt.CustomerEmail, "Payment", "Success")
		if err == nil {
			w.Cache.Set(ctx, key, "done", 24*time.Hour)
			return nil
		}
		time.Sleep(backoff)
		backoff *= 2
	}

	return err
}
