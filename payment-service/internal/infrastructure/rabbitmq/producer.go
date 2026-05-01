package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment-service/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func NewProducer(url, exchange, queue, dlx, dlq string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare an exchange: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    dlx,
		"x-dead-letter-routing-key": dlq,
	}
	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = ch.QueueBind(
		queue,
		queue,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	return &Producer{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (p *Producer) PublishPaymentCompleted(ctx context.Context, event usecase.PaymentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = p.channel.PublishWithContext(ctx,
			p.exchange,
			p.queue,
			true,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
			})

		if err == nil {
			break
		}
		
		log.Printf("Warning: failed to publish message, attempt %d/%d: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
		}
	}

	if err != nil {
		return fmt.Errorf("failed to publish a message after %d retries: %w", maxRetries, err)
	}

	log.Printf(" [x] Sent %s", body)
	return nil
}

func (p *Producer) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
