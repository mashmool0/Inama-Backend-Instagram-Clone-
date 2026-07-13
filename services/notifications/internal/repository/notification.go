package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/pagination"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) InsertNotification(ctx context.Context, tx pgx.Tx, notification model.Notification) error {
	_, err := tx.Exec(ctx, `
INSERT INTO notifications (id, recipient_id, actor_id, type, post_id, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, notification.ID, notification.RecipientID, notification.ActorID, int16(notification.Type), notification.PostID, notification.IsRead, notification.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListNotificationsPage(ctx context.Context, recipientID string, limit int32, cursor *pagination.EdgeCursor) ([]model.Notification, *pagination.EdgeCursor, error) {
	baseQuery := `
SELECT id, recipient_id, actor_id, type, post_id, is_read, created_at
FROM notifications
WHERE recipient_id = $1
`
	args := []any{recipientID}

	if cursor != nil {
		baseQuery += ` AND (created_at < $2 OR (created_at = $2 AND id < $3))`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}

	baseQuery += `
ORDER BY created_at DESC, id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list notifications page: %w", err)
	}
	defer rows.Close()

	notifications := make([]model.Notification, 0, limit)
	var next *pagination.EdgeCursor

	for rows.Next() {
		var notification model.Notification
		var notifType int16
		if err := rows.Scan(
			&notification.ID,
			&notification.RecipientID,
			&notification.ActorID,
			&notifType,
			&notification.PostID,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan notification row: %w", err)
		}
		notification.Type = model.NotificationType(notifType)
		notifications = append(notifications, notification)
		next = &pagination.EdgeCursor{CreatedAt: notification.CreatedAt, ID: notification.ID}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate notification rows: %w", err)
	}

	if int32(len(notifications)) < limit {
		next = nil
	}

	return notifications, next, nil
}

func (r *NotificationRepository) NotificationBelongsToRecipient(ctx context.Context, notificationID, recipientID string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
SELECT EXISTS(
    SELECT 1 FROM notifications
    WHERE id = $1 AND recipient_id = $2
)
`, notificationID, recipientID).Scan(&exists); err != nil {
		return false, fmt.Errorf("notification belongs to recipient: %w", err)
	}
	return exists, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID, recipientID string) error {
	tag, err := r.pool.Exec(ctx, `
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND recipient_id = $2
`, notificationID, recipientID)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, recipientID string) error {
	if _, err := r.pool.Exec(ctx, `
UPDATE notifications
SET is_read = TRUE
WHERE recipient_id = $1 AND is_read = FALSE
`, recipientID); err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}
	return nil
}

func (r *NotificationRepository) InsertNotificationFields(ctx context.Context, tx pgx.Tx, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	columns := make([]string, 0, len(fields))
	placeholders := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))
	idx := 1
	for column, value := range fields {
		columns = append(columns, column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		args = append(args, value)
		idx++
	}
	query := fmt.Sprintf("INSERT INTO notifications (%s) VALUES (%s)", strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return err
	}
	return nil
}

var ErrDuplicateProcessedEvent = errors.New("duplicate processed event")
