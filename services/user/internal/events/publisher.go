package events

import "context"

type UserFollowedPayload struct {
	FollowerID string
	FolloweeID string
}

type UserUpdatedPayload struct {
	UserID    string
	Username  string
	AvatarURL string
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
