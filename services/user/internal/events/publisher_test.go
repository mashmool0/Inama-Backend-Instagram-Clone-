package events

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type recordingBroker struct {
	bodies   [][]byte
	keys     []string
	failures int
}

func (b *recordingBroker) Publish(_ context.Context, routingKey string, body []byte) error {
	b.keys = append(b.keys, routingKey)
	b.bodies = append(b.bodies, append([]byte(nil), body...))
	if len(b.bodies) <= b.failures {
		return errors.New("broker unavailable")
	}
	return nil
}

func TestRabbitPublisherBuildsStandardFollowEnvelope(t *testing.T) {
	t.Parallel()

	broker := &recordingBroker{}
	publisher := NewRabbitPublisher(broker, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := publisher.UserFollowed(context.Background(), UserFollowedPayload{
		FollowerID: "actor-id",
		FolloweeID: "recipient-id",
	}); err != nil {
		t.Fatalf("UserFollowed() error = %v", err)
	}

	if len(broker.bodies) != 1 || broker.keys[0] != EventTypeUserFollowed {
		t.Fatalf("published keys = %v, want [%s]", broker.keys, EventTypeUserFollowed)
	}

	var envelope struct {
		EventID    string              `json:"event_id"`
		EventType  string              `json:"event_type"`
		OccurredAt string              `json:"occurred_at"`
		Data       UserFollowedPayload `json:"data"`
	}
	if err := json.Unmarshal(broker.bodies[0], &envelope); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if envelope.EventID == "" || envelope.OccurredAt == "" {
		t.Fatal("event_id and occurred_at must be populated")
	}
	if envelope.EventType != EventTypeUserFollowed {
		t.Fatalf("event_type = %q, want %q", envelope.EventType, EventTypeUserFollowed)
	}
	if envelope.Data.FollowerID != "actor-id" || envelope.Data.FolloweeID != "recipient-id" {
		t.Fatalf("payload = %+v", envelope.Data)
	}
}

func TestRabbitPublisherRetriesSameEventEnvelope(t *testing.T) {
	t.Parallel()

	broker := &recordingBroker{failures: 2}
	publisher := NewRabbitPublisher(broker, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := publisher.UserFollowed(context.Background(), UserFollowedPayload{
		FollowerID: "actor-id",
		FolloweeID: "recipient-id",
	}); err != nil {
		t.Fatalf("UserFollowed() error = %v", err)
	}

	if len(broker.bodies) != 3 {
		t.Fatalf("publish attempts = %d, want 3", len(broker.bodies))
	}
	for i := 1; i < len(broker.bodies); i++ {
		if string(broker.bodies[i]) != string(broker.bodies[0]) {
			t.Fatal("retries must preserve event_id and envelope bytes")
		}
	}
}
