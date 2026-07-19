package model

import "time"

type Profile struct {
	ID             string
	Username       string
	Bio            string
	AvatarURL      string
	FollowerCount  int64
	FollowingCount int64
	CreatedAt      time.Time
}

type UpdateProfileInput struct {
	Username  *string
	Bio       *string
	AvatarURL *string
}

type UserIDPage struct {
	UserIDs    []string
	NextCursor string
}
