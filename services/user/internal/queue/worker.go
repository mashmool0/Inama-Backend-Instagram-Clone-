package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mashmool0/inama/services/user/internal/service"
)

type Config struct {
	URL         string
	Exchange    string
	QueueName   string
	RoutingKeys []string
}

type Broker struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange  string
	queueName string
}

type deliveryProcessor interface {
	Process(context.Context, []byte) error
}

type Worker struct {
	broker    *Broker
	processor deliveryProcessor
}

func NewWorker(broker *Broker, processor deliveryProcessor) *Worker {
	return &Worker{broker: broker, processor: processor}
}

func (w *Worker) Run(ctx context.Context) error {
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
				err = delivery.Ack(false)
			case service.IsRejectMessage(err):
				err = delivery.Nack(false, false)
			default:
				err = delivery.Nack(false, true)
			}
			if err != nil {
				return fmt.Errorf("settle delivery: %w", err)
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
			channel, channelErr := conn.Channel()
			if channelErr == nil {
				if topologyErr := declareTopology(channel, cfg); topologyErr == nil {
					return &Broker{conn: conn, channel: channel, exchange: cfg.Exchange, queueName: cfg.QueueName}, nil
				} else {
					lastErr = topologyErr
				}
				_ = channel.Close()
			} else {
				lastErr = channelErr
			}
			_ = conn.Close()
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

func declareTopology(channel *amqp.Channel, cfg Config) error {
	if err := channel.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	queue, err := channel.QueueDeclare(cfg.QueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	for _, key := range cfg.RoutingKeys {
		if err := channel.QueueBind(queue.Name, key, cfg.Exchange, false, nil); err != nil {
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
