package events

import "encoding/json"

const (
	EventTypeUserRegistered  = "user.registered"
	EventTypeUsernameUpdated = "user.username_updated"
)

type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type UserRegisteredPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UsernameUpdatedPayload struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
