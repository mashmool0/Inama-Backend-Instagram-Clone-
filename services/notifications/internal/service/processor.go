package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/notifications/internal/events"
	"github.com/mashmool0/inama/services/notifications/internal/model"
	"github.com/mashmool0/inama/services/notifications/internal/push"
)

type EventProcessor interface {
	Process(ctx context.Context, body []byte) error
}

type notificationWriter interface {
	InsertNotification(ctx context.Context, tx pgx.Tx, notification model.Notification) error
}

type processedEventWriter interface {
	InsertIfAbsent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error)
}

type Processor struct {
	pool            *pgxpool.Pool
	notifications   notificationWriter
	processedEvents processedEventWriter
	pushDispatcher  push.Dispatcher
	now             func() time.Time
}

func NewProcessor(pool *pgxpool.Pool, notifications notificationWriter, processedEvents processedEventWriter, pushDispatcher push.Dispatcher) *Processor {
	return &Processor{
		pool:            pool,
		notifications:   notifications,
		processedEvents: processedEvents,
		pushDispatcher:  pushDispatcher,
		now:             func() time.Time { return time.Now().UTC() },
	}
}

func (p *Processor) Process(ctx context.Context, body []byte) error {
	var envelope events.Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("%w: decode envelope: %v", ErrRejectMessage, err)
	}
	if envelope.EventID == "" || envelope.EventType == "" {
		return fmt.Errorf("%w: missing event metadata", ErrRejectMessage)
	}

	notification, skip, err := p.mapEnvelope(envelope)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	inserted, err := p.processedEvents.InsertIfAbsent(ctx, tx, envelope.EventID, envelope.EventType)
	if err != nil {
		return err
	}
	if !inserted {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		return nil
	}

	if err := p.notifications.InsertNotification(ctx, tx, notification); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	if p.pushDispatcher != nil {
		p.pushDispatcher.Dispatch(ctx, notification)
	}

	return nil
}

func (p *Processor) mapEnvelope(envelope events.Envelope) (model.Notification, bool, error) {
	switch envelope.EventType {
	case events.EventTypePostLiked:
		var payload events.PostLikedPayload
		if err := decodePayload(envelope.Data, &payload); err != nil {
			return model.Notification{}, false, err
		}
		if payload.ActorID == payload.PostAuthorID {
			return model.Notification{}, true, nil
		}
		postID := payload.PostID
		return p.newNotification(payload.PostAuthorID, payload.ActorID, model.NotificationTypeLike, &postID)
	case events.EventTypeCommentCreated:
		var payload events.CommentCreatedPayload
		if err := decodePayload(envelope.Data, &payload); err != nil {
			return model.Notification{}, false, err
		}
		if payload.ActorID == payload.PostAuthorID {
			return model.Notification{}, true, nil
		}
		postID := payload.PostID
		return p.newNotification(payload.PostAuthorID, payload.ActorID, model.NotificationTypeComment, &postID)
	case events.EventTypeUserFollowed:
		var payload events.UserFollowedPayload
		if err := decodePayload(envelope.Data, &payload); err != nil {
			return model.Notification{}, false, err
		}
		if payload.FollowerID == payload.FolloweeID {
			return model.Notification{}, true, nil
		}
		return p.newNotification(payload.FolloweeID, payload.FollowerID, model.NotificationTypeFollow, nil)
	default:
		return model.Notification{}, false, fmt.Errorf("%w: unsupported event type %s", ErrRejectMessage, envelope.EventType)
	}
}

func decodePayload(data []byte, target any) error {
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%w: decode payload: %v", ErrRejectMessage, err)
	}
	return nil
}

func (p *Processor) newNotification(recipientID, actorID string, typ model.NotificationType, postID *string) (model.Notification, bool, error) {
	if recipientID == "" || actorID == "" {
		return model.Notification{}, false, fmt.Errorf("%w: missing recipient or actor", ErrRejectMessage)
	}
	id, err := newUUID()
	if err != nil {
		return model.Notification{}, false, err
	}
	return model.Notification{
		ID:          id,
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        typ,
		PostID:      postID,
		IsRead:      false,
		CreatedAt:   p.now(),
	}, false, nil
}

func IsRejectError(err error) bool {
	return errors.Is(err, ErrRejectMessage)
}
