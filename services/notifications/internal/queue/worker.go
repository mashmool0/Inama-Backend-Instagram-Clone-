package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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
	broker *Broker
}

func NewWorker(broker *Broker) *BrokerWorker {
	return &BrokerWorker{broker: broker}
}

func (w *BrokerWorker) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
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
