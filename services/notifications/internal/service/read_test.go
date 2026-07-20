package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/pagination"
)

type stubReadRepository struct {
	listNotificationsPageFn        func(context.Context, string, int32, *pagination.EdgeCursor) ([]model.Notification, *pagination.EdgeCursor, error)
	notificationBelongsToRecipient func(context.Context, string, string) (bool, error)
	markAsReadFn                   func(context.Context, string, string) error
	markAllReadFn                  func(context.Context, string) error
}

func (s stubReadRepository) ListNotificationsPage(ctx context.Context, recipientID string, limit int32, cursor *pagination.EdgeCursor) ([]model.Notification, *pagination.EdgeCursor, error) {
	return s.listNotificationsPageFn(ctx, recipientID, limit, cursor)
}

func (s stubReadRepository) NotificationBelongsToRecipient(ctx context.Context, notificationID, recipientID string) (bool, error) {
	return s.notificationBelongsToRecipient(ctx, notificationID, recipientID)
}

func (s stubReadRepository) MarkAsRead(ctx context.Context, notificationID, recipientID string) error {
	return s.markAsReadFn(ctx, notificationID, recipientID)
}

func (s stubReadRepository) MarkAllRead(ctx context.Context, recipientID string) error {
	return s.markAllReadFn(ctx, recipientID)
}

func TestReadManagerGetNotificationsRejectsInvalidCursor(t *testing.T) {
	t.Parallel()

	svc := NewReadService(stubReadRepository{
		listNotificationsPageFn: func(context.Context, string, int32, *pagination.EdgeCursor) ([]model.Notification, *pagination.EdgeCursor, error) {
			return nil, nil, nil
		},
	}, 20, 100)

	_, err := svc.GetNotifications(context.Background(), "recipient-1", 10, "not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("GetNotifications() error = %v, want ErrInvalidCursor", err)
	}
}

func TestReadManagerMarkAsReadRejectsForeignNotification(t *testing.T) {
	t.Parallel()

	svc := NewReadService(stubReadRepository{
		notificationBelongsToRecipient: func(context.Context, string, string) (bool, error) {
			return false, nil
		},
		markAsReadFn: func(context.Context, string, string) error {
			t.Fatal("MarkAsRead should not be called when ownership check fails")
			return nil
		},
	}, 20, 100)

	err := svc.MarkAsRead(context.Background(), "recipient-1", "notif-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("MarkAsRead() error = %v, want ErrNotFound", err)
	}
}

func TestReadManagerMarkAsReadMapsMissingRowToNotFound(t *testing.T) {
	t.Parallel()

	svc := NewReadService(stubReadRepository{
		notificationBelongsToRecipient: func(context.Context, string, string) (bool, error) {
			return true, nil
		},
		markAsReadFn: func(context.Context, string, string) error {
			return pgx.ErrNoRows
		},
	}, 20, 100)

	err := svc.MarkAsRead(context.Background(), "recipient-1", "notif-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("MarkAsRead() error = %v, want ErrNotFound", err)
	}
}

func TestReadManagerNormalizeLimitCapsValues(t *testing.T) {
	t.Parallel()

	svc := NewReadService(stubReadRepository{}, 20, 100)
	if got := svc.normalizeLimit(0); got != 20 {
		t.Fatalf("normalizeLimit(0) = %d, want 20", got)
	}
	if got := svc.normalizeLimit(500); got != 100 {
		t.Fatalf("normalizeLimit(500) = %d, want 100", got)
	}
	if got := svc.normalizeLimit(25); got != 25 {
		t.Fatalf("normalizeLimit(25) = %d, want 25", got)
	}
}
