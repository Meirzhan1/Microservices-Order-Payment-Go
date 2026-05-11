package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"notification-service/internal/infrastructure/email"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PaymentEvent struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type Consumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	queue        string
	exchange     string
	dlq          string
	dlx          string
	redisClient  *redis.Client
	emailSender  email.EmailSender
}

func NewConsumer(url, exchange, queue, dlx, dlq string, rdb *redis.Client, sender email.EmailSender) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		dlx,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLX: %w", err)
	}

	_, err = ch.QueueDeclare(
		dlq,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	err = ch.QueueBind(dlq, dlq, dlx, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind DLQ: %w", err)
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
		return nil, fmt.Errorf("failed to declare main exchange: %w", err)
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
		return nil, fmt.Errorf("failed to declare main queue: %w", err)
	}

	err = ch.QueueBind(
		queue,
		queue,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind main queue: %w", err)
	}

	return &Consumer{
		conn:        conn,
		channel:     ch,
		queue:       queue,
		exchange:    exchange,
		dlx:         dlx,
		dlq:         dlq,
		redisClient: rdb,
		emailSender: sender,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	err := c.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				c.handleDelivery(d)
			}
		}
	}()

	return nil
}

func (c *Consumer) handleDelivery(d amqp.Delivery) {
	var event PaymentEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		d.Nack(false, false)
		return
	}

	idempotencyKey := fmt.Sprintf("notification:payment:%s", event.PaymentID)
	ctx := context.Background()
	
	val, err := c.redisClient.Get(ctx, idempotencyKey).Result()
	if err == nil && val == "processed" {
		log.Printf("[Idempotency] Payment #%s for Order #%s already processed. Skipping.", event.PaymentID, event.OrderID)
		d.Ack(false)
		return
	}

	maxRetries := 5
	baseDelay := 2 * time.Second
	
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lastErr = c.emailSender.Send(ctx, event.CustomerEmail, 
			fmt.Sprintf("Payment for Order #%s", event.OrderID),
			fmt.Sprintf("Your payment of $%d was successful.", event.Amount))
		
		if lastErr == nil {
			break
		}

		delay := time.Duration(math.Pow(2, float64(i))) * baseDelay
		log.Printf("[Retry] Attempt %d failed for Order #%s: %v. Retrying in %v...", i+1, event.OrderID, lastErr, delay)
		time.Sleep(delay)
	}

	if lastErr != nil {
		log.Printf("[Failure] Failed to send notification for Order #%s after %d attempts: %v. Moving to DLQ.", 
			event.OrderID, maxRetries, lastErr)
		d.Nack(false, false)
		return
	}

	err = c.redisClient.Set(ctx, idempotencyKey, "processed", 24*time.Hour).Err()
	if err != nil {
		log.Printf("[Warning] Failed to set idempotency key in Redis: %v", err)
	}

	log.Printf("[Notification] Successfully processed Order #%s", event.OrderID)

	if err := d.Ack(false); err != nil {
		log.Printf("Error acknowledging message: %v", err)
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
