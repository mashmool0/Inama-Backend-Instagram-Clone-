package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mashmool0/inama/services/notifications/internal/service"
)

type Worker interface {
	Run(ctx context.Context) error
}

type Config struct {
	URL         string
	Exchange    string
	QueueName   string
	RoutingKeys []string
}

type Broker struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

type BrokerWorker struct {
	broker    *Broker
	processor deliveryProcessor
}

type deliveryProcessor interface {
	Process(ctx context.Context, body []byte) error
}

func NewWorker(broker *Broker, processor deliveryProcessor) *BrokerWorker {
	return &BrokerWorker{broker: broker, processor: processor}
}

func (w *BrokerWorker) Run(ctx context.Context) error {
	deliveries, err := w.broker.channel.Consume(w.broker.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume deliveries: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			err := w.processor.Process(ctx, delivery.Body)
			switch {
			case err == nil:
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack delivery: %w", ackErr)
				}
			case service.IsRejectError(err):
				if nackErr := delivery.Nack(false, false); nackErr != nil {
					return fmt.Errorf("reject delivery: %w", nackErr)
				}
			default:
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					return fmt.Errorf("requeue delivery: %w", nackErr)
				}
			}
		}
	}
}

func Connect(ctx context.Context, cfg Config) (*Broker, error) {
	var lastErr error
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		conn, err := amqp.Dial(cfg.URL)
		if err == nil {
			ch, chErr := conn.Channel()
			if chErr != nil {
				_ = conn.Close()
				lastErr = chErr
			} else if err := declareTopology(ch, cfg); err != nil {
				_ = ch.Close()
				_ = conn.Close()
				lastErr = err
			} else {
				return &Broker{
					conn:      conn,
					channel:   ch,
					queueName: cfg.QueueName,
				}, nil
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connect rabbitmq: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}

	return nil, fmt.Errorf("connect rabbitmq: %w", lastErr)
}

func declareTopology(ch *amqp.Channel, cfg Config) error {
	if err := ch.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	queue, err := ch.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	for _, key := range cfg.RoutingKeys {
		if err := ch.QueueBind(queue.Name, key, cfg.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind queue for %s: %w", key, err)
		}
	}

	return nil
}

func (b *Broker) Close() error {
	if b == nil {
		return nil
	}
	if b.channel != nil {
		_ = b.channel.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
