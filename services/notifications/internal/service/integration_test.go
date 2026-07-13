package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/notifications/internal/events"
	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/repository"
	"github.com/mashmool0/inama/services/notifications/internal/schema"
)

const integrationDatabaseEnv = "INTEGRATION_DATABASE_URL"

type recordingDispatcher struct {
	notifications []model.Notification
}

func (r *recordingDispatcher) Dispatch(_ context.Context, notification model.Notification) {
	r.notifications = append(r.notifications, notification)
}

func TestProcessorIntegrationCreatesNotificationAndIgnoresDuplicateEvent(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)

	notifications := repository.NewNotificationRepository(pool)
	processedEvents := repository.NewProcessedEventRepository(pool)
	dispatcher := &recordingDispatcher{}
	processor := NewProcessor(pool, notifications, processedEvents, dispatcher)
	processor.now = func() time.Time { return time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC) }

	body := envelopeJSON(t, "event-1", events.EventTypePostLiked, events.PostLikedPayload{
		PostID:       "11111111-1111-1111-1111-111111111111",
		PostAuthorID: "22222222-2222-2222-2222-222222222222",
		ActorID:      "33333333-3333-3333-3333-333333333333",
	})

	if err := processor.Process(context.Background(), body); err != nil {
		t.Fatalf("Process() first call error = %v", err)
	}
	if err := processor.Process(context.Background(), body); err != nil {
		t.Fatalf("Process() second call error = %v", err)
	}

	var notificationCount int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM notifications`).Scan(&notificationCount); err != nil {
		t.Fatalf("count notifications error = %v", err)
	}
	if notificationCount != 1 {
		t.Fatalf("notifications count = %d, want 1", notificationCount)
	}

	var processedCount int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM processed_events`).Scan(&processedCount); err != nil {
		t.Fatalf("count processed_events error = %v", err)
	}
	if processedCount != 1 {
		t.Fatalf("processed_events count = %d, want 1", processedCount)
	}

	if len(dispatcher.notifications) != 1 {
		t.Fatalf("dispatcher notifications = %d, want 1", len(dispatcher.notifications))
	}
}

func TestProcessorIntegrationSkipsSelfNotification(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)

	processor := NewProcessor(
		pool,
		repository.NewNotificationRepository(pool),
		repository.NewProcessedEventRepository(pool),
		nil,
	)

	body := envelopeJSON(t, "event-2", events.EventTypeCommentCreated, events.CommentCreatedPayload{
		CommentID:    "44444444-4444-4444-4444-444444444444",
		PostID:       "11111111-1111-1111-1111-111111111111",
		PostAuthorID: "22222222-2222-2222-2222-222222222222",
		ActorID:      "22222222-2222-2222-2222-222222222222",
	})

	if err := processor.Process(context.Background(), body); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	var notificationCount int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM notifications`).Scan(&notificationCount); err != nil {
		t.Fatalf("count notifications error = %v", err)
	}
	if notificationCount != 0 {
		t.Fatalf("notifications count = %d, want 0", notificationCount)
	}
}

func TestReadServiceIntegrationMarkAllReadOnlyTouchesRecipient(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)
	loadSeed(t, pool)

	repo := repository.NewNotificationRepository(pool)
	svc := NewReadService(repo, 20, 100)

	if err := svc.MarkAllRead(context.Background(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"); err != nil {
		t.Fatalf("MarkAllRead() error = %v", err)
	}

	var recipientOneUnread int
	if err := pool.QueryRow(context.Background(), `
SELECT COUNT(*) FROM notifications WHERE recipient_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' AND is_read = FALSE
`).Scan(&recipientOneUnread); err != nil {
		t.Fatalf("count recipient-1 unread error = %v", err)
	}
	if recipientOneUnread != 0 {
		t.Fatalf("recipient-1 unread count = %d, want 0", recipientOneUnread)
	}

	var recipientTwoUnread int
	if err := pool.QueryRow(context.Background(), `
SELECT COUNT(*) FROM notifications WHERE recipient_id = 'dddddddd-dddd-dddd-dddd-dddddddddddd' AND is_read = FALSE
`).Scan(&recipientTwoUnread); err != nil {
		t.Fatalf("count recipient-2 unread error = %v", err)
	}
	if recipientTwoUnread != 1 {
		t.Fatalf("recipient-2 unread count = %d, want 1", recipientTwoUnread)
	}
}

func TestReadServiceIntegrationPagination(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)
	loadSeed(t, pool)

	repo := repository.NewNotificationRepository(pool)
	svc := NewReadService(repo, 2, 2)

	page1, err := svc.GetNotifications(context.Background(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", 2, "")
	if err != nil {
		t.Fatalf("GetNotifications() page1 error = %v", err)
	}
	if len(page1.Notifications) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1.Notifications))
	}
	if page1.NextCursor == "" {
		t.Fatal("page1 NextCursor should not be empty")
	}

	page2, err := svc.GetNotifications(context.Background(), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", 2, page1.NextCursor)
	if err != nil {
		t.Fatalf("GetNotifications() page2 error = %v", err)
	}
	if len(page2.Notifications) != 1 {
		t.Fatalf("page2 len = %d, want 1", len(page2.Notifications))
	}
	if page2.NextCursor != "" {
		t.Fatalf("page2 NextCursor = %q, want empty", page2.NextCursor)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(integrationDatabaseEnv)
	if dsn == "" {
		t.Skipf("%s is not set", integrationDatabaseEnv)
	}

	pool, err := repository.OpenPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("OpenPool() error = %v", err)
	}
	t.Cleanup(pool.Close)

	if err := schema.Bootstrap(context.Background(), pool); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	return pool
}

func resetDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), `TRUNCATE TABLE processed_events, notifications`); err != nil {
		t.Fatalf("TRUNCATE error = %v", err)
	}
}

func loadSeed(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	seedPath := filepath.Join("..", "..", "testdata", "seed.sql")
	payload, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", seedPath, err)
	}

	if _, err := pool.Exec(context.Background(), string(payload)); err != nil {
		t.Fatalf("Exec(seed.sql) error = %v", err)
	}
}

func envelopeJSON(t *testing.T, eventID, eventType string, payload any) []byte {
	t.Helper()

	body, err := json.Marshal(events.Envelope{
		EventID:    eventUUID(eventID),
		EventType:  eventType,
		OccurredAt: "2026-07-12T12:00:00Z",
		Data:       mustMarshalJSON(t, payload),
	})
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}
	return body
}

func eventUUID(seed string) string {
	switch seed {
	case "event-1":
		return "99999999-9999-9999-9999-999999999991"
	case "event-2":
		return "99999999-9999-9999-9999-999999999992"
	default:
		return "99999999-9999-9999-9999-999999999999"
	}
}
