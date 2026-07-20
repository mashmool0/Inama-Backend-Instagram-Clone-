package queue

import (
	"context"
	"fmt"
	"sync"
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
	publisher *amqp.Channel
	confirms  <-chan amqp.Confirmation
	exchange  string
	queueName string
	publishMu sync.Mutex
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
					publisher, publisherErr := conn.Channel()
					if publisherErr == nil {
						if confirmErr := publisher.Confirm(false); confirmErr == nil {
							confirms := publisher.NotifyPublish(make(chan amqp.Confirmation, 1))
							return &Broker{conn: conn, channel: channel, publisher: publisher, confirms: confirms, exchange: cfg.Exchange, queueName: cfg.QueueName}, nil
						} else {
							lastErr = confirmErr
						}
						_ = publisher.Close()
					} else {
						lastErr = publisherErr
					}
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

func (b *Broker) Publish(ctx context.Context, routingKey string, body []byte) error {
	b.publishMu.Lock()
	defer b.publishMu.Unlock()

	if err := b.publisher.PublishWithContext(ctx, b.exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	}); err != nil {
		return fmt.Errorf("publish %s: %w", routingKey, err)
	}

	select {
	case confirmation, ok := <-b.confirms:
		if !ok {
			return fmt.Errorf("publish %s: confirmation channel closed", routingKey)
		}
		if !confirmation.Ack {
			return fmt.Errorf("publish %s: broker rejected delivery", routingKey)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("publish %s: %w", routingKey, ctx.Err())
	}
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
	if b.publisher != nil {
		_ = b.publisher.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
