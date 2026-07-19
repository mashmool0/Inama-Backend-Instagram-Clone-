package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mashmool0/inama/services/notifications/internal/events"
	"github.com/mashmool0/inama/services/notifications/internal/model"
)

func TestProcessorMapEnvelopeSkipsSelfNotifications(t *testing.T) {
	t.Parallel()

	processor := &Processor{}
	envelope := events.Envelope{
		EventType: events.EventTypePostLiked,
		Data: mustMarshalJSON(t, events.PostLikedPayload{
			PostID:       "post-1",
			PostAuthorID: "user-1",
			ActorID:      "user-1",
		}),
	}

	_, skip, err := processor.mapEnvelope(envelope)
	if err != nil {
		t.Fatalf("mapEnvelope() error = %v", err)
	}
	if !skip {
		t.Fatal("mapEnvelope() should skip self notifications")
	}
}

func TestProcessorMapEnvelopeRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	processor := &Processor{}
	_, _, err := processor.mapEnvelope(events.Envelope{
		EventType: events.EventTypeCommentCreated,
		Data:      []byte(`{"post_id":123}`),
	})
	if !IsRejectError(err) {
		t.Fatalf("mapEnvelope() error = %v, want reject error", err)
	}
}

func TestProcessorMapEnvelopeBuildsFollowNotification(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 12, 15, 0, 0, 0, time.UTC)
	processor := &Processor{
		now: func() time.Time { return now },
	}

	notification, skip, err := processor.mapEnvelope(events.Envelope{
		EventType: events.EventTypeUserFollowed,
		Data: mustMarshalJSON(t, events.UserFollowedPayload{
			FollowerID: "user-1",
			FolloweeID: "user-2",
		}),
	})
	if err != nil {
		t.Fatalf("mapEnvelope() error = %v", err)
	}
	if skip {
		t.Fatal("mapEnvelope() unexpectedly skipped follow event")
	}
	if notification.RecipientID != "user-2" || notification.ActorID != "user-1" {
		t.Fatalf("notification recipient/actor = %q/%q, want user-2/user-1", notification.RecipientID, notification.ActorID)
	}
	if notification.Type != model.NotificationTypeFollow {
		t.Fatalf("notification.Type = %v, want follow", notification.Type)
	}
	if notification.PostID != nil {
		t.Fatalf("notification.PostID = %v, want nil", *notification.PostID)
	}
	if notification.CreatedAt != now {
		t.Fatalf("notification.CreatedAt = %v, want %v", notification.CreatedAt, now)
	}
	if notification.ID == "" {
		t.Fatal("notification.ID should not be empty")
	}
}

func TestProcessorProcessRejectsMalformedEnvelope(t *testing.T) {
	t.Parallel()

	processor := &Processor{}
	err := processor.Process(context.Background(), []byte(`{"event_type":`))
	if !IsRejectError(err) {
		t.Fatalf("Process() error = %v, want reject error", err)
	}
}

func TestIsRejectErrorRecognizesWrappedErrors(t *testing.T) {
	t.Parallel()

	err := errors.Join(ErrRejectMessage, errors.New("bad payload"))
	if !IsRejectError(err) {
		t.Fatalf("IsRejectError(%v) = false, want true", err)
	}
}

func mustMarshalJSON(t *testing.T, payload any) []byte {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}
