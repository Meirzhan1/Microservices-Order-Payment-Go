package email

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type MockEmailSender struct {
	failProbability float64
}

func NewMockEmailSender(failProbability float64) *MockEmailSender {
	return &MockEmailSender{failProbability: failProbability}
}

func (s *MockEmailSender) Send(ctx context.Context, to, subject, body string) error {
	time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)

	if rand.Float64() < s.failProbability {
		return errors.New("temporary external provider failure")
	}

	log.Printf("[MockEmail] Email sent to %s: %s", to, subject)
	return nil
}
