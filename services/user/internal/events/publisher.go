package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeUserFollowed = "user.followed"
	EventTypeUserUpdated  = "user.updated"
)

type UserFollowedPayload struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

type UserUpdatedPayload struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type Publisher interface {
	UserFollowed(ctx context.Context, payload UserFollowedPayload) error
	UserUpdated(ctx context.Context, payload UserUpdatedPayload) error
}

type NopPublisher struct{}

func (NopPublisher) UserFollowed(context.Context, UserFollowedPayload) error {
	return nil
}

func (NopPublisher) UserUpdated(context.Context, UserUpdatedPayload) error {
	return nil
}

type eventBroker interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
}

type RabbitPublisher struct {
	broker   eventBroker
	logger   *slog.Logger
	attempts int
}

func NewRabbitPublisher(broker eventBroker, logger *slog.Logger) *RabbitPublisher {
	return &RabbitPublisher{broker: broker, logger: logger, attempts: 3}
}

func (p *RabbitPublisher) UserFollowed(ctx context.Context, payload UserFollowedPayload) error {
	return p.publish(ctx, EventTypeUserFollowed, payload)
}

func (p *RabbitPublisher) UserUpdated(ctx context.Context, payload UserUpdatedPayload) error {
	return p.publish(ctx, EventTypeUserUpdated, payload)
}

func (p *RabbitPublisher) publish(ctx context.Context, eventType string, data any) error {
	envelope := struct {
		EventID    string    `json:"event_id"`
		EventType  string    `json:"event_type"`
		OccurredAt time.Time `json:"occurred_at"`
		Data       any       `json:"data"`
	}{
		EventID:    uuid.NewString(),
		EventType:  eventType,
		OccurredAt: time.Now().UTC(),
		Data:       data,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal %s event: %w", eventType, err)
	}

	for attempt := 1; attempt <= p.attempts; attempt++ {
		if err = p.broker.Publish(ctx, eventType, body); err == nil {
			return nil
		}
		if attempt < p.attempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
			}
		}
	}

	p.logger.Error("failed to publish user event", "event_type", eventType, "error", err)
	return fmt.Errorf("publish %s event: %w", eventType, err)
}
