package model

import "time"

type NotificationType int16

const (
	NotificationTypeUnspecified NotificationType = 0
	NotificationTypeLike        NotificationType = 1
	NotificationTypeComment     NotificationType = 2
	NotificationTypeFollow      NotificationType = 3
)

type Notification struct {
	ID         string
	RecipientID string
	ActorID    string
	Type       NotificationType
	PostID     *string
	IsRead     bool
	CreatedAt  time.Time
}

type NotificationPage struct {
	Notifications []Notification
	NextCursor    string
}
