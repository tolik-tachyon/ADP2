package worker

import (
	"context"
	"encoding/json"
	"errors"
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
	sub, err := w.Js.Subscribe(
		"payment.completed",
		func(msg *nats.Msg) {

			var evt domain.Event
			if err := json.Unmarshal(msg.Data, &evt); err != nil {
				_ = msg.Ack()
				return
			}

			// handle pipeline
			if err := w.handle(ctx, evt); err != nil {
				// ❌ no ack → NATS retry (only for retryable failures)
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

// -------------------- CORE PIPELINE --------------------

func (w *Worker) handle(ctx context.Context, evt domain.Event) error {

	// 1. VALIDATION
	if err := validate(evt); err != nil {
		w.sendDLQ(evt, "validation failed: "+err.Error())
		return nil // already handled
	}

	// 2. IDEMPOTENCY
	key := "notif:" + evt.EventID
	ok, _ := w.Cache.SetNX(ctx, key, "processing", 24*time.Hour).Result()
	if !ok {
		return nil // already processed
	}

	// 3. RETRY EMAIL SENDING
	err := w.sendWithRetry(evt)
	if err != nil {
		w.sendDLQ(evt, "email failed after retries")
		return err
	}

	// 4. MARK DONE
	_ = w.Cache.Set(ctx, key, "done", 24*time.Hour).Err()

	return nil
}

// -------------------- RETRY LOGIC --------------------

func (w *Worker) sendWithRetry(evt domain.Event) error {
	var err error
	backoff := 2 * time.Second

	for i := 0; i < 5; i++ {
		err = w.Sender.Send(evt.CustomerEmail, "Payment Success", "Your payment was successful")

		if err == nil {
			return nil
		}

		time.Sleep(backoff)
		backoff *= 2
	}

	return err
}

// -------------------- DLQ --------------------

func (w *Worker) sendDLQ(evt domain.Event, reason string) {
	data, _ := json.Marshal(map[string]any{
		"event":  evt,
		"reason": reason,
		"time":   time.Now().UTC(),
	})

	_, _ = w.Js.Publish("payment.dlq", data)
}

// -------------------- VALIDATION --------------------

func validate(evt domain.Event) error {
	if evt.EventID == "" {
		return errors.New("missing event_id")
	}
	if evt.OrderID == "" {
		return errors.New("missing order_id")
	}
	if evt.CustomerEmail == "" {
		return errors.New("missing customer_email")
	}
	if evt.Amount <= 0 {
		return errors.New("invalid amount")
	}
	return nil
}
