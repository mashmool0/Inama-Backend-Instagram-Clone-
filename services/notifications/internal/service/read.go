package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/pagination"
)

type ReadService interface {
	GetNotifications(ctx context.Context, recipientID string, limit int32, cursor string) (model.NotificationPage, error)
	MarkAsRead(ctx context.Context, recipientID, notificationID string) error
	MarkAllRead(ctx context.Context, recipientID string) error
}

type readRepository interface {
	ListNotificationsPage(ctx context.Context, recipientID string, limit int32, cursor *pagination.EdgeCursor) ([]model.Notification, *pagination.EdgeCursor, error)
	NotificationBelongsToRecipient(ctx context.Context, notificationID, recipientID string) (bool, error)
	MarkAsRead(ctx context.Context, notificationID, recipientID string) error
	MarkAllRead(ctx context.Context, recipientID string) error
}

type ReadManager struct {
	repo         readRepository
	defaultLimit int32
	maxLimit     int32
}

func NewReadService(repo readRepository, defaultLimit, maxLimit int32) *ReadManager {
	return &ReadManager{repo: repo, defaultLimit: defaultLimit, maxLimit: maxLimit}
}

func (s *ReadManager) GetNotifications(ctx context.Context, recipientID string, limit int32, cursor string) (model.NotificationPage, error) {
	if recipientID == "" {
		return model.NotificationPage{}, ErrUnauthenticated
	}

	decoded, err := pagination.Decode(cursor)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return model.NotificationPage{}, ErrInvalidCursor
		}
		return model.NotificationPage{}, err
	}

	notifications, next, err := s.repo.ListNotificationsPage(ctx, recipientID, s.normalizeLimit(limit), decoded)
	if err != nil {
		return model.NotificationPage{}, err
	}

	nextCursor, err := pagination.Encode(next)
	if err != nil {
		return model.NotificationPage{}, err
	}

	return model.NotificationPage{
		Notifications: notifications,
		NextCursor:    nextCursor,
	}, nil
}

func (s *ReadManager) MarkAsRead(ctx context.Context, recipientID, notificationID string) error {
	if recipientID == "" {
		return ErrUnauthenticated
	}
	if notificationID == "" {
		return NewInvalidArgument("notification_id is required")
	}

	belongs, err := s.repo.NotificationBelongsToRecipient(ctx, notificationID, recipientID)
	if err != nil {
		return err
	}
	if !belongs {
		return ErrNotFound
	}

	if err := s.repo.MarkAsRead(ctx, notificationID, recipientID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ReadManager) MarkAllRead(ctx context.Context, recipientID string) error {
	if recipientID == "" {
		return ErrUnauthenticated
	}
	return s.repo.MarkAllRead(ctx, recipientID)
}

func (s *ReadManager) normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return s.defaultLimit
	}
	if limit > s.maxLimit {
		return s.maxLimit
	}
	return limit
}
