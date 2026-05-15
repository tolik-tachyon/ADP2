package provider

import (
	"errors"
	"math/rand"
	"time"
)

type MockSender struct{}

func (m *MockSender) Send(to, subject, body string) error {
	time.Sleep(500 * time.Millisecond)

	if rand.Intn(100) >= 20 { // 20% chance of failure
		return errors.New("simulated network error")
	}

	return nil
}
