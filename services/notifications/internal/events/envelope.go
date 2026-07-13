package events

import "encoding/json"

const (
	EventTypePostLiked      = "post.liked"
	EventTypeCommentCreated = "comment.created"
	EventTypeUserFollowed   = "user.followed"
)

type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type PostLikedPayload struct {
	PostID       string `json:"post_id"`
	PostAuthorID string `json:"post_author_id"`
	ActorID      string `json:"actor_id"`
}

type CommentCreatedPayload struct {
	CommentID    string `json:"comment_id"`
	PostID       string `json:"post_id"`
	PostAuthorID string `json:"post_author_id"`
	ActorID      string `json:"actor_id"`
}

type UserFollowedPayload struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}
